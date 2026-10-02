package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	runtime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"os"
	"strings"
	"time"
)

var prefix string

func finish(kind string, count int) {
	fmt.Printf("AUDIT prefix=%q kind=%s active_matches=%d read_only=true\n", prefix, kind, count)
	if count != 0 {
		panic("active disposable fixtures remain")
	}
}

// audit performs read-only fixture inventory. Nonzero active prefix matches
// fail the gate; recorded tombstone IDs and an optional channel are checked.
func main() {
	prefixFlag := flag.String("prefix", "", "disposable fixture prefix (required)")
	domain := flag.String("domain", "runtime", "runtime or modelsecurity")
	channel := flag.String("channel", "", "optional borrowed Network Broker UUID to verify")
	tombstones := flag.String("tombstones", "", "comma-separated recorded group UUIDs to verify")
	flag.Parse()
	prefix = *prefixFlag
	if prefix == "" {
		panic("a nonempty disposable prefix is required")
	}
	if *domain != "runtime" && *domain != "modelsecurity" {
		panic("unsupported domain")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if *domain == "modelsecurity" {
		c, err := modelsecurity.NewClient(modelsecurity.Opts{ClientID: os.Getenv("PANW_MGMT_CLIENT_ID"), ClientSecret: os.Getenv("PANW_MGMT_CLIENT_SECRET"), TsgID: os.Getenv("PANW_MGMT_TSG_ID")})
		if err != nil {
			panic(err)
		}
		count := 0
		seen := map[string]bool{}
		for skip := 0; ; {
			p, err := c.SecurityGroups.List(ctx, modelsecurity.GroupListOpts{Skip: skip, Limit: 100})
			if err != nil {
				panic(err)
			}
			for _, g := range p.Items {
				if seen[g.UUID] {
					panic("repeated group page")
				}
				seen[g.UUID] = true
				if strings.HasPrefix(g.Name, prefix) {
					if g.IsTombstone {
						fmt.Printf("RESIDUAL kind=model-security-group id=%s tombstoned=true confirmed=true\n", g.UUID)
					} else {
						count++
						fmt.Printf("ACTIVE group id=%s\n", g.UUID)
					}
				}
			}
			skip += len(p.Items)
			if len(p.Items) == 0 || (p.Metadata.TotalItems != nil && skip >= *p.Metadata.TotalItems) {
				break
			}
		}
		known := map[string]bool{}
		for _, id := range strings.Split(*tombstones, ",") {
			if id != "" {
				known[id] = true
			}
		}
		for id := range known {
			g, err := c.SecurityGroups.Get(ctx, id)
			if err != nil {
				panic(err)
			}
			if !g.IsTombstone {
				panic("recorded disposable group is not tombstoned")
			}
			fmt.Printf("RESIDUAL kind=model-security-group id=%s tombstoned=true confirmed=true\n", g.UUID)
		}
		finish("model-security-group", count)
		return
	}
	r, err := runtime.NewClient(runtime.Opts{ClientID: os.Getenv("PANW_MGMT_CLIENT_ID"), ClientSecret: os.Getenv("PANW_MGMT_CLIENT_SECRET"), TsgID: os.Getenv("PANW_MGMT_TSG_ID")})
	if err != nil {
		panic(err)
	}
	for _, kind := range []string{"profile-revision", "topic", "key", "customer-app"} {
		count := 0
		seen := map[string]bool{}
		for offset := 0; ; {
			names := map[string]string{}
			next := 0
			length := 0
			switch kind {
			case "profile-revision":
				latest := false
				p, e := r.Profiles.ListWithOptions(ctx, runtime.ProfileListOpts{ListOpts: runtime.ListOpts{Limit: 100, Offset: offset}, Latest: &latest})
				if e != nil {
					panic(e)
				}
				next = p.NextOffset
				length = len(p.Items)
				for _, x := range p.Items {
					names[x.ProfileID] = x.ProfileName
				}
			case "topic":
				p, e := r.Topics.List(ctx, runtime.ListOpts{Limit: 100, Offset: offset})
				if e != nil {
					panic(e)
				}
				next = p.NextOffset
				length = len(p.Items)
				for _, x := range p.Items {
					names[x.TopicID] = x.TopicName
				}
			case "key":
				p, e := r.ApiKeys.List(ctx, runtime.ListOpts{Limit: 100, Offset: offset})
				if e != nil {
					panic(e)
				}
				next = p.NextOffset
				length = len(p.Items)
				for _, x := range p.Items {
					names[x.ApiKeyID] = x.ApiKeyName
				}
			case "customer-app":
				p, e := r.CustomerApps.List(ctx, runtime.ListOpts{Limit: 100, Offset: offset})
				if e != nil {
					panic(e)
				}
				next = p.NextOffset
				length = len(p.Items)
				for _, x := range p.Items {
					names[x.CustomerAppID] = x.AppName
				}
			}
			if length == 0 && next > offset {
				panic("empty advancing page")
			}
			for id, name := range names {
				if seen[id] {
					panic("repeated runtime page")
				}
				seen[id] = true
				if strings.HasPrefix(name, prefix) {
					count++
					fmt.Printf("ACTIVE kind=%s id=%s\n", kind, id)
				}
			}
			if next > offset {
				offset = next
			} else if length >= 100 {
				offset += length
			} else {
				break
			}
		}
		finish(kind, count)
	}
	c, err := redteam.NewClient(redteam.Opts{ClientID: os.Getenv("PANW_MGMT_CLIENT_ID"), ClientSecret: os.Getenv("PANW_MGMT_CLIENT_SECRET"), TsgID: os.Getenv("PANW_MGMT_TSG_ID")})
	if err != nil {
		panic(err)
	}
	count := 0
	seen := map[string]bool{}
	for skip := 0; ; {
		page, err := c.Targets.List(ctx, redteam.TargetListOpts{Skip: skip, Limit: 100})
		if err != nil {
			panic(err)
		}
		for _, item := range page.Data {
			if seen[item.UUID] {
				panic("repeated target page")
			}
			seen[item.UUID] = true
			if strings.HasPrefix(item.Name, prefix) {
				count++
				fmt.Printf("ACTIVE kind=target id=%s\n", item.UUID)
			}
		}
		skip += len(page.Data)
		if len(page.Data) == 0 || skip >= page.Pagination.Total {
			break
		}
	}
	finish("target", count)
	for _, archived := range []bool{false, true} {
		count = 0
		seen = map[string]bool{}
		for skip := 0; ; {
			page, err := c.CustomAttacks.ListPromptSets(ctx, redteam.PromptSetListOpts{Skip: skip, Limit: 100, Archive: &archived})
			if err != nil {
				panic(err)
			}
			for _, item := range page.Data {
				if seen[item.UUID] {
					panic("repeated promptset page")
				}
				seen[item.UUID] = true
				if strings.HasPrefix(item.Name, prefix) {
					count++
					fmt.Printf("RESIDUAL kind=promptset id=%s archived=%t confirmed=true\n", item.UUID, archived)
				}
			}
			skip += len(page.Data)
			if len(page.Data) == 0 || skip >= page.Pagination.Total {
				break
			}
		}
		if !archived {
			finish("promptset", count)
		} else {
			fmt.Printf("AUDIT kind=promptset archived_matches=%d read_only=true\n", count)
		}
	}
	if *channel != "" {
		g, err := c.NetworkBroker.Get(ctx, *channel)
		if err != nil {
			panic(err)
		}
		if g.UUID == nil || *g.UUID != *channel {
			panic("borrowed channel identity mismatch")
		}
		fmt.Printf("BORROWED fixture=network-broker-channel id=%s read_only=true still_present=true\n", *g.UUID)
	}
}
