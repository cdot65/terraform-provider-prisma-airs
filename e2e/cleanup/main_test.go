package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec"
)

func TestCleanupSnapshotsBeforeDeletionShiftsPages(t *testing.T) {
	var items []fixture
	for i := 0; i < 101; i++ {
		name := "unrelated"
		if i == 0 || i == 100 {
			name = "owned-run"
		}
		items = append(items, fixture{fmt.Sprint(i), name})
	}
	list := func(_ context.Context, offset int) ([]fixture, int, error) {
		end := offset + 100
		if end > len(items) {
			end = len(items)
		}
		if offset >= len(items) {
			return nil, 0, nil
		}
		next := 0
		if end < len(items) {
			next = end
		}
		return append([]fixture(nil), items[offset:end]...), next, nil
	}
	deleted := map[string]bool{}
	remove := func(_ context.Context, id string) error {
		deleted[id] = true
		for i, item := range items {
			if item.id == id {
				items = append(items[:i], items[i+1:]...)
				break
			}
		}
		return nil
	}
	if err := cleanup(context.Background(), "fixture", "owned-run", false, list, remove); err != nil {
		t.Fatal(err)
	}
	if !deleted["0"] || !deleted["100"] || len(deleted) != 2 {
		t.Fatal("deletion skipped a shifted page or touched unrelated inventory")
	}
}

func TestCleanupNeverAcceptsUnconfirmedSuccess(t *testing.T) {
	for _, err := range []error{nil, aisec.NewHTTPError("already absent", aisec.ClientSideError, 404)} {
		list := func(context.Context, int) ([]fixture, int, error) {
			return []fixture{{"owned-id", "owned-run"}}, 0, nil
		}
		if cleanup(context.Background(), "fixture", "owned-run", false, list, func(context.Context, string) error { return err }) == nil {
			t.Fatal("present object treated as deleted")
		}
	}
}

func TestInventoryRejectsRepeatedPages(t *testing.T) {
	list := func(context.Context, int) ([]fixture, int, error) { return []fixture{{"same-id", "owned-run"}}, 7, nil }
	if _, err := inventory(context.Background(), list); err == nil {
		t.Fatal("repeated inventory page accepted")
	}
}
