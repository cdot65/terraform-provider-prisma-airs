// cleanup deletes disposable runtime profiles (all revisions) and custom topics.
// Use an exact run-specific prefix; other domains are verified by acceptance fixtures.
// Usage: go run ./e2e/cleanup --prefix=e2e-tf --dry-run
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cdot65/prisma-airs-go/aisec"
	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
)

type fixture struct{ id, name string }
type listing func(context.Context, int) ([]fixture, int, error)

// Snapshot the complete inventory before deleting. Deleting while paging by
// offset can shift the inventory and skip objects from later pages.
func inventory(ctx context.Context, list listing) ([]fixture, error) {
	var result []fixture
	seen := map[string]bool{}
	for offset := 0; ; {
		page, next, err := list(ctx, offset)
		if err != nil {
			return nil, err
		}
		if len(page) == 0 && next > offset {
			return nil, fmt.Errorf("empty page with advancing cursor")
		}
		for _, item := range page {
			if item.id == "" || seen[item.id] {
				return nil, fmt.Errorf("missing identity or repeated page")
			}
			seen[item.id] = true
			result = append(result, item)
		}
		if next > offset {
			offset = next
		} else if len(page) >= 100 {
			offset += len(page)
		} else {
			return result, nil
		}
	}
}

func provisionalDelete(err error) bool {
	if err == nil || aisec.IsNotFound(err) {
		return true
	}
	var sdk *aisec.AISecSDKError
	return errors.As(err, &sdk) && sdk.StatusCode >= 200 && sdk.StatusCode < 300 && sdk.ErrorType == aisec.AISecSDKInternalError
}

func cleanup(ctx context.Context, kind, prefix string, dryRun bool, list listing, remove func(context.Context, string) error) error {
	items, err := inventory(ctx, list)
	if err != nil {
		return fmt.Errorf("list %s: %w", kind, err)
	}
	var failures []error
	for _, item := range items {
		if !strings.HasPrefix(item.name, prefix) {
			continue
		}
		if dryRun {
			fmt.Printf("would delete %s name=%q id=%s\n", kind, item.name, item.id)
			continue
		}
		deleteErr := remove(ctx, item.id)
		if !provisionalDelete(deleteErr) {
			failures = append(failures, fmt.Errorf("delete %s %s: %w", kind, item.id, deleteErr))
			continue
		}
		remaining, err := inventory(ctx, list)
		if err != nil {
			failures = append(failures, fmt.Errorf("verify %s %s: %w", kind, item.id, err))
			continue
		}
		absent := true
		for _, candidate := range remaining {
			if candidate.id == item.id {
				absent = false
				break
			}
		}
		if !absent {
			failures = append(failures, fmt.Errorf("%s %s remains after delete", kind, item.id))
			continue
		}
		fmt.Printf("deleted %s name=%q id=%s confirmed=true\n", kind, item.name, item.id)
	}
	return errors.Join(failures...)
}

func main() {
	prefix := flag.String("prefix", "e2e-tf", "exact disposable run prefix to delete")
	dryRun := flag.Bool("dry-run", false, "list without deleting")
	flag.Parse()
	if *prefix == "" {
		fmt.Fprintln(os.Stderr, "ERROR: a nonempty disposable prefix is required")
		os.Exit(1)
	}
	client, err := airsruntime.NewClient(airsruntime.Opts{ClientID: os.Getenv("PANW_MGMT_CLIENT_ID"), ClientSecret: os.Getenv("PANW_MGMT_CLIENT_SECRET"), TsgID: os.Getenv("PANW_MGMT_TSG_ID")})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	profiles := func(ctx context.Context, offset int) ([]fixture, int, error) {
		latest := false
		page, err := client.Profiles.ListWithOptions(ctx, airsruntime.ProfileListOpts{ListOpts: airsruntime.ListOpts{Limit: 100, Offset: offset}, Latest: &latest})
		if err != nil {
			return nil, 0, err
		}
		var items []fixture
		for _, p := range page.Items {
			items = append(items, fixture{p.ProfileID, p.ProfileName})
		}
		return items, page.NextOffset, nil
	}
	topics := func(ctx context.Context, offset int) ([]fixture, int, error) {
		page, err := client.Topics.List(ctx, airsruntime.ListOpts{Limit: 100, Offset: offset})
		if err != nil {
			return nil, 0, err
		}
		var items []fixture
		for _, p := range page.Items {
			items = append(items, fixture{p.TopicID, p.TopicName})
		}
		return items, page.NextOffset, nil
	}
	profileErr := cleanup(ctx, "profile revision", *prefix, *dryRun, profiles, func(ctx context.Context, id string) error {
		_, err := client.Profiles.ForceDelete(ctx, id, "terraform-e2e-cleanup")
		return err
	})
	topicErr := cleanup(ctx, "topic", *prefix, *dryRun, topics, func(ctx context.Context, id string) error {
		_, err := client.Topics.ForceDelete(ctx, id, "terraform-e2e-cleanup")
		return err
	})
	if err := errors.Join(profileErr, topicErr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *dryRun {
		fmt.Println("Dry-run inventory complete; no deletion requested.")
	} else {
		fmt.Println("Cleanup inventory complete; requested deletions confirmed.")
	}
}
