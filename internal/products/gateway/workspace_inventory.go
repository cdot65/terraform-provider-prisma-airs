package gateway

import (
	"context"
	"fmt"
	"math"

	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
)

// The captured workspace route has no paging parameters. Positive identities
// are useful, but absence/uniqueness requires a complete, consistent envelope.
func (r *workspaceResource) inventory(ctx context.Context, status string) ([]document, int64, bool, *bool, error) {
	reply, e := r.client.sdk.Workspaces.List(ctx, gw.WorkspaceListOptions{Plane: gw.WorkspaceAdmin, Status: status})
	if e != nil {
		return nil, 0, false, nil, e
	}
	if reply.Total < 0 || math.IsNaN(reply.Total) || math.IsInf(reply.Total, 0) || reply.Total != math.Trunc(reply.Total) || reply.Total > float64(math.MaxInt64-1024) {
		return nil, 0, false, nil, fmt.Errorf("invalid workspace total")
	}
	rows := make([]document, len(reply.Data))
	seen := map[string]bool{}
	for i, row := range reply.Data {
		if row.ID == "" || row.Slug == "" || seen[row.ID] {
			return nil, 0, false, nil, fmt.Errorf("invalid or duplicate workspace identity")
		}
		seen[row.ID] = true
		rows[i], e = decodeDocument(row)
		if e != nil {
			return nil, 0, false, nil, e
		}
	}
	complete := reply.HasMore != nil && !*reply.HasMore && reply.Total == float64(len(rows))
	return rows, int64(reply.Total), complete, reply.HasMore, nil
}
func (r *workspaceResource) findScopeWorkspace(ctx context.Context, scope string) (document, error) {
	found := map[string]document{}
	for _, status := range []string{"active", "archived"} {
		rows, _, complete, _, e := r.inventory(ctx, status)
		if e != nil {
			return nil, e
		}
		if !complete {
			return nil, &inputShapeError{field: "scope_name", detail: "cannot be reconciled from an incomplete workspace inventory. Inspect the retained identities and import the exact workspace UUID before retrying"}
		}
		for _, row := range rows {
			if row["scope_name"] == scope {
				found[row["id"].(string)] = row
			}
		}
	}
	if len(found) > 1 {
		return nil, &inputShapeError{field: "scope_name", detail: "references multiple workspaces; dedicated ownership cannot be established"}
	}
	for _, row := range found {
		return row, nil
	}
	return nil, nil
}
