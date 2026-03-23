// cleanup deletes orphaned e2e resources by name prefix.
//
// Usage:
//
//	go run ./e2e/cleanup [--prefix=e2e-tf] [--dry-run]
//
// Requires environment variables: PANW_MGMT_CLIENT_ID, PANW_MGMT_CLIENT_SECRET, PANW_MGMT_TSG_ID
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
)

func main() {
	prefix := flag.String("prefix", "e2e-tf", "name prefix to match for deletion")
	dryRun := flag.Bool("dry-run", false, "list matches without deleting")
	flag.Parse()

	clientID := os.Getenv("PANW_MGMT_CLIENT_ID")
	clientSecret := os.Getenv("PANW_MGMT_CLIENT_SECRET")
	tsgID := os.Getenv("PANW_MGMT_TSG_ID")

	if clientID == "" || clientSecret == "" || tsgID == "" {
		fmt.Fprintln(os.Stderr, "ERROR: PANW_MGMT_CLIENT_ID, PANW_MGMT_CLIENT_SECRET, PANW_MGMT_TSG_ID required")
		os.Exit(1)
	}

	client, err := airsruntime.NewClient(airsruntime.Opts{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TsgID:        tsgID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR creating client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	deleted := 0

	// Clean up security profiles
	fmt.Printf("Scanning security profiles with prefix %q...\n", *prefix)
	offset := 0
	for {
		resp, err := client.Profiles.List(ctx, airsruntime.ListOpts{Limit: 100, Offset: offset})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR listing profiles: %v\n", err)
			break
		}
		for _, p := range resp.Items {
			if strings.HasPrefix(p.ProfileName, *prefix) {
				if *dryRun {
					fmt.Printf("  [dry-run] would delete profile: %s (id=%s)\n", p.ProfileName, p.ProfileID)
				} else {
					fmt.Printf("  deleting profile: %s (id=%s)... ", p.ProfileName, p.ProfileID)
					if _, err := client.Profiles.Delete(ctx, p.ProfileID); err != nil {
						// SDK bug: delete succeeds but returns unparseable empty body
						if strings.Contains(err.Error(), "failed to parse response JSON") {
							fmt.Println("OK (empty response)")
							deleted++
						} else {
							fmt.Printf("FAILED: %v\n", err)
						}
					} else {
						fmt.Println("OK")
						deleted++
					}
				}
			}
		}
		if len(resp.Items) < 100 {
			break
		}
		offset += 100
	}

	// Clean up custom topics
	fmt.Printf("Scanning custom topics with prefix %q...\n", *prefix)
	offset = 0
	for {
		resp, err := client.Topics.List(ctx, airsruntime.ListOpts{Limit: 100, Offset: offset})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR listing topics: %v\n", err)
			break
		}
		for _, t := range resp.Items {
			if strings.HasPrefix(t.TopicName, *prefix) {
				if *dryRun {
					fmt.Printf("  [dry-run] would delete topic: %s (id=%s)\n", t.TopicName, t.TopicID)
				} else {
					fmt.Printf("  deleting topic: %s (id=%s)... ", t.TopicName, t.TopicID)
					if _, err := client.Topics.Delete(ctx, t.TopicID); err != nil {
						fmt.Printf("FAILED: %v\n", err)
					} else {
						fmt.Println("OK")
						deleted++
					}
				}
			}
		}
		if len(resp.Items) < 100 {
			break
		}
		offset += 100
	}

	if *dryRun {
		fmt.Println("Dry run complete. No resources deleted.")
	} else {
		fmt.Printf("Cleanup complete. Deleted %d resource(s).\n", deleted)
	}
}
