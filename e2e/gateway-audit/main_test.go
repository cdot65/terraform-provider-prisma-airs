package main

import (
	"context"
	"fmt"
	"testing"
)

func TestAuditIncludesLaterPagesAndRejectsRepeatedPages(t *testing.T) {
	first := make([]item, 100)
	for i := range first {
		first[i] = item{ID: fmt.Sprint(i), Name: "unrelated", Status: "active"}
	}
	total := int64(101)
	for _, repeat := range []bool{false, true} {
		t.Run(fmt.Sprint(repeat), func(t *testing.T) {
			calls := 0
			c := collection{name: "fixture", paged: true, read: func(_ context.Context, page int64) (any, error) {
				calls++
				if page == 1 || repeat {
					return struct {
						Data  []item `json:"data"`
						Total int64  `json:"total"`
					}{first, total}, nil
				}
				return struct {
					Data  []item `json:"data"`
					Total int64  `json:"total"`
				}{[]item{{ID: "owned", Name: "disposable-owned", Status: "active"}}, total}, nil
			}}
			if e := audit(context.Background(), c, "disposable-"); e == nil || calls != 2 {
				t.Fatal("later-page fixture or repeated page was ignored")
			}
		})
	}
}
func TestAuditDoesNotAcceptListingFailureAsAbsence(t *testing.T) {
	c := collection{name: "fixture", read: func(context.Context, int64) (any, error) { return nil, fmt.Errorf("not authorized") }}
	if audit(context.Background(), c, "disposable-") == nil {
		t.Fatal("listing failure was accepted")
	}
}
