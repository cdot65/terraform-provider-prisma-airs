package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Live acceptance helpers. Credentials come only from the PANW_MGMT_* environment
// variables (never from files in the repo). Every helper that mutates the tenant
// is scoped to a uniquely named disposable resource created by the test itself.

func accCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Minute)
}

func accMgmtClient(t *testing.T) *airsruntime.Client {
	t.Helper()
	c, err := airsruntime.NewClient(airsruntime.Opts{
		ClientID:      os.Getenv("PANW_MGMT_CLIENT_ID"),
		ClientSecret:  os.Getenv("PANW_MGMT_CLIENT_SECRET"),
		TsgID:         os.Getenv("PANW_MGMT_TSG_ID"),
		APIEndpoint:   os.Getenv("PANW_MGMT_ENDPOINT"),
		TokenEndpoint: os.Getenv("PANW_MGMT_TOKEN_ENDPOINT"),
	})
	if err != nil {
		t.Fatalf("management client: %v", err)
	}
	return c
}

func accModelSecClient(t *testing.T) *modelsecurity.Client {
	t.Helper()
	c, err := modelsecurity.NewClient(modelsecurity.Opts{
		ClientID:      os.Getenv("PANW_MGMT_CLIENT_ID"),
		ClientSecret:  os.Getenv("PANW_MGMT_CLIENT_SECRET"),
		TsgID:         os.Getenv("PANW_MGMT_TSG_ID"),
		DataEndpoint:  os.Getenv("PANW_MODEL_SEC_DATA_ENDPOINT"),
		MgmtEndpoint:  os.Getenv("PANW_MODEL_SEC_MGMT_ENDPOINT"),
		TokenEndpoint: os.Getenv("PANW_MGMT_TOKEN_ENDPOINT"),
	})
	if err != nil {
		t.Fatalf("model security client: %v", err)
	}
	return c
}

func accRedTeamClient(t *testing.T) *redteam.Client {
	t.Helper()
	c, err := redteam.NewClient(redteam.Opts{
		ClientID:      os.Getenv("PANW_MGMT_CLIENT_ID"),
		ClientSecret:  os.Getenv("PANW_MGMT_CLIENT_SECRET"),
		TsgID:         os.Getenv("PANW_MGMT_TSG_ID"),
		DataEndpoint:  os.Getenv("PANW_RED_TEAM_DATA_ENDPOINT"),
		MgmtEndpoint:  os.Getenv("PANW_RED_TEAM_MGMT_ENDPOINT"),
		TokenEndpoint: os.Getenv("PANW_MGMT_TOKEN_ENDPOINT"),
	})
	if err != nil {
		t.Fatalf("red team client: %v", err)
	}
	return c
}

// idRecorder captures a resource attribute during a Check so a later step's
// PreConfig can delete that object remotely.
type idRecorder struct{ value string }

func (r *idRecorder) capture(attr string, resourceAddr string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %s not in state", resourceAddr)
		}
		v := rs.Primary.Attributes[attr]
		if v == "" {
			return fmt.Errorf("%s.%s is empty", resourceAddr, attr)
		}
		r.value = v
		return nil
	}
}

func (r *idRecorder) different(attr, addr string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		value := s.RootModule().Resources[addr].Primary.Attributes[attr]
		if value == "" || value == r.value {
			return fmt.Errorf("remote deletion did not produce a new identity")
		}
		fmt.Printf("RECOVERY resource=%s prior=%s new=%s confirmed=true\n", addr, r.value, value)
		return nil
	}
}

// destroyCheck builds a CheckDestroy that runs absent for every managed
// resource of resourceType, failing when the remote object still exists.
func destroyCheck(resourceType, idAttr string, absent func(ctx context.Context, id string) (bool, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		ctx, cancel := accCtx()
		defer cancel()
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			id := rs.Primary.Attributes[idAttr]
			gone, err := absent(ctx, id)
			if err != nil {
				return fmt.Errorf("verify destroy of %s %s: %w", resourceType, id, err)
			}
			if !gone {
				return fmt.Errorf("%s %s still exists after destroy", resourceType, id)
			}
			outcome := "deleted"
			if resourceType == "prisma-airs_supply_chain_security_group" {
				outcome = "tombstoned"
			}
			if resourceType == "prisma-airs_red_team_custom_prompt_set" {
				outcome = "archived"
			}
			fmt.Printf("CLEANUP resource=%s id=%s outcome=%s confirmed=true\n", resourceType, id, outcome)
		}
		return nil
	}
}

func goneOnNotFound(err error) (bool, error) {
	if err == nil {
		return false, nil
	}
	if aisec.IsNotFound(err) {
		return true, nil
	}
	return false, err
}

func redteamArchive() redteam.CustomPromptSetArchiveRequest {
	return redteam.CustomPromptSetArchiveRequest{Archive: true}
}

func accCustomerApp(ctx context.Context, client *airsruntime.Client, name string) (*airsruntime.CustomerApp, error) {
	for offset := 0; ; {
		page, err := client.CustomerApps.List(ctx, airsruntime.ListOpts{Limit: 100, Offset: offset})
		if err != nil {
			return nil, err
		}
		for i := range page.Items {
			if page.Items[i].AppName == name {
				return &page.Items[i], nil
			}
		}
		if page.NextOffset > offset {
			offset = page.NextOffset
		} else if len(page.Items) >= 100 {
			offset += len(page.Items)
		} else {
			return nil, aisec.NewHTTPError("customer app absent", aisec.ClientSideError, 404)
		}
	}
}

func accProfileCleanup(t *testing.T, client *airsruntime.Client, names ...string) {
	t.Helper()
	ctx, cancel := accCtx()
	defer cancel()
	for _, name := range names {
		revs, err := accProfileRevisions(ctx, client, name)
		if err != nil {
			t.Fatal("profile fixture ownership check failed")
		}
		if len(revs) > 0 {
			t.Fatal("disposable fixture name already exists; refusing cleanup ownership")
		}
	}
	t.Cleanup(func() {
		ctx, cancel := accCtx()
		defer cancel()
		for _, name := range names {
			revs, err := accProfileRevisions(ctx, client, name)
			if err != nil {
				t.Error("fixture cleanup history lookup failed")
				continue
			}
			for _, p := range revs {
				_, err := client.Profiles.ForceDelete(ctx, p.ProfileID, "terraform-acc")
				if err != nil && !gone404(err) {
					t.Errorf("fixture cleanup failed UUID=%s", p.ProfileID)
				}
			}
			left, err := accProfileRevisions(ctx, client, name)
			if err != nil || len(left) > 0 {
				t.Errorf("fixture cleanup unconfirmed name=%s", name)
			}
		}
	})
}
