// Package product defines the provider's product composition seam.
package product

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Credentials are shared defaults; each product owns construction of its SDK client.
type Credentials struct {
	ClientID, ClientSecret, TsgID, TokenEndpoint string
}

func (c Credentials) Complete() bool {
	return c.ClientID != "" && c.ClientSecret != "" && c.TsgID != ""
}

// Data contains opaque product-owned client state, never a cross-product SDK bag.
type Data map[string]any

// Client retrieves only the calling product's concrete SDK client.
func Client[T comparable](data any, id, label string) (T, diag.Diagnostics) {
	var empty T
	var diagnostics diag.Diagnostics
	products, ok := data.(Data)
	if !ok {
		diagnostics.AddError("Invalid provider configuration", "Expected product-owned provider data.")
		return empty, diagnostics
	}
	client, ok := products[id].(T)
	if !ok || client == empty {
		diagnostics.AddError(label+" client not configured", "OAuth2 credentials required: client_id, client_secret and tsg_id.")
		return empty, diagnostics
	}
	return client, diagnostics
}

type Endpoint struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Description string `json:"description"`
}

type Resource struct {
	New   func() resource.Resource
	Guide string
}

type DataSource struct {
	New   func() datasource.DataSource
	Guide string
}

// Definition owns a product's configuration, registrations and documentation.
type Definition struct {
	ID          string
	Label       string
	Implemented bool
	Endpoints   []Endpoint
	Resources   []Resource
	DataSources []DataSource
	Configure   func(Credentials, map[string]string) (any, error)
}

type TypeMetadata struct {
	Name  string `json:"name"`
	Guide string `json:"guide"`
}

type Metadata struct {
	ID          string         `json:"id"`
	Label       string         `json:"label"`
	Implemented bool           `json:"implemented"`
	Endpoints   []Endpoint     `json:"endpoints"`
	Resources   []TypeMetadata `json:"resources"`
	DataSources []TypeMetadata `json:"data_sources"`
}

// Metadata derives public names from actual Terraform constructors, rather than
// maintaining another inventory of schema names.
func (d Definition) Metadata(ctx context.Context, providerName string) Metadata {
	m := Metadata{ID: d.ID, Label: d.Label, Implemented: d.Implemented, Endpoints: d.Endpoints,
		Resources: []TypeMetadata{}, DataSources: []TypeMetadata{}}
	if m.Endpoints == nil {
		m.Endpoints = []Endpoint{}
	}
	for _, entry := range d.Resources {
		var response resource.MetadataResponse
		entry.New().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: providerName}, &response)
		m.Resources = append(m.Resources, TypeMetadata{Name: response.TypeName, Guide: entry.Guide})
	}
	for _, entry := range d.DataSources {
		var response datasource.MetadataResponse
		entry.New().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: providerName}, &response)
		m.DataSources = append(m.DataSources, TypeMetadata{Name: response.TypeName, Guide: entry.Guide})
	}
	return m
}
