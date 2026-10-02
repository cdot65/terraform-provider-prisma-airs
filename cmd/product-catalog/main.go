// product-catalog emits the same inventory used by provider registration.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/cdot65/prisma-airs-provider/internal/products"
)

func main() {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(products.Metadata(context.Background())); err != nil {
		log.Fatal(err)
	}
}
