package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec/gateway"
)

type client struct {
	sdk          *gateway.Client
	organisation string
}
type document map[string]any

func decodeDocument(v any) (document, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	var d document
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if e = dec.Decode(&d); e != nil {
		return nil, e
	}
	// Generic upstream responses use an envelope; SCM returns flat records.
	if nested, ok := d["data"].(map[string]any); ok {
		for k, v := range nested {
			if _, exists := d[k]; !exists {
				d[k] = v
			}
		}
		delete(d, "data")
	}
	for _, key := range []string{"config", "configurations"} {
		if encoded, ok := d[key].(string); ok {
			var value any
			dec := json.NewDecoder(bytes.NewBufferString(encoded))
			dec.UseNumber()
			if e := dec.Decode(&value); e != nil {
				return nil, fmt.Errorf("invalid Gateway configuration response")
			}
			d[key] = value
		}
	}
	return d, nil
}

func writeSDK[Q, R any](ctx context.Context, body document, call func(context.Context, Q) (*R, error)) (document, error) {
	b, e := json.Marshal(body)
	if e != nil {
		return nil, e
	}
	var q Q
	if e = json.Unmarshal(b, &q); e != nil {
		var typed *json.UnmarshalTypeError
		if errors.As(e, &typed) {
			// SDK field paths contain schema names, never submitted values.
			return nil, &inputShapeError{field: typed.Field, detail: "has a type incompatible with the Gateway SDK contract"}
		}
		return nil, &inputShapeError{field: "settings", detail: "do not match the Gateway SDK contract"}
	}
	r, e := call(ctx, q)
	if e != nil {
		return nil, e
	}
	return decodeDocument(r)
}
func readSDK[R any](r *R, err error) (document, error) {
	if err != nil {
		return nil, err
	}
	return decodeDocument(r)
}

func listSDK[R any](r *R, err error) ([]document, int64, error) {
	if err != nil {
		return nil, 0, err
	}
	b, e := json.Marshal(r)
	if e != nil {
		return nil, 0, e
	}
	var v map[string]json.RawMessage
	if e = json.Unmarshal(b, &v); e != nil {
		return nil, 0, e
	}
	var items []document
	if len(v["data"]) == 0 {
		v["data"] = v["workspaces"]
	}
	dec := json.NewDecoder(bytes.NewReader(v["data"]))
	dec.UseNumber()
	if len(v["data"]) > 0 && string(v["data"]) != "null" {
		if e = dec.Decode(&items); e != nil {
			return nil, 0, e
		}
	}
	if items == nil {
		items = []document{}
	}
	total := int64(len(items))
	if len(v["total"]) > 0 {
		if e = json.Unmarshal(v["total"], &total); e != nil {
			return nil, 0, e
		}
	}
	return items, total, nil
}

type inputShapeError struct{ field, detail string }

func (e *inputShapeError) Error() string {
	field := e.field
	for _, c := range field {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.[]", c) {
			field = "settings"
			break
		}
	}
	return "Gateway attribute " + field + " " + e.detail + ". No input values are included."
}

// Absence is authoritative only after a complete workspace listing. The current
// SDK exposes no paging knobs here, so retain state if metadata reports truncation.
func bindingItems[R any](response *R, err error) ([]document, bool, error) {
	items, total, err := listSDK(response, err)
	if err != nil {
		return nil, false, err
	}
	envelope, err := decodeDocument(response)
	if err != nil {
		return nil, false, err
	}
	complete := total <= int64(len(items))
	if value, present := envelope["has_more"]; present && value != nil {
		more, ok := value.(bool)
		if !ok {
			return nil, false, fmt.Errorf("cannot verify workspace list paging metadata")
		}
		complete = complete && !more
	}
	return items, complete, nil
}
