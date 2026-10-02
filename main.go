package main

import (
	"context"
	"flag"
	"log"

	"github.com/cdot65/prisma-airs-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate python3 docs-site/scripts/schema_reference.py

var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/cdot65/prisma-airs",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
