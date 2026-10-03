// Package products composes product-owned implementations into one provider.
package products

import (
	"context"

	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/cdot65/prisma-airs-provider/internal/products/gateway"
	"github.com/cdot65/prisma-airs-provider/internal/products/redteam"
	"github.com/cdot65/prisma-airs-provider/internal/products/runtime"
	"github.com/cdot65/prisma-airs-provider/internal/products/supplychain"
)

func All() []product.Definition {
	return []product.Definition{
		runtime.Definition(), redteam.Definition(),
		gateway.Definition(),
		supplychain.Definition(),
	}
}

func Metadata(ctx context.Context) []product.Metadata {
	metadata := make([]product.Metadata, 0, 4)
	for _, definition := range All() {
		metadata = append(metadata, definition.Metadata(ctx, "prisma-airs"))
	}
	return metadata
}
