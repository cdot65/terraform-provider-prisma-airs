package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringValueOrEnv_configValue(t *testing.T) {
	val := types.StringValue("from-config")
	result := stringValueOrEnv(val, "TEST_UNUSED_ENV_VAR")
	if result != "from-config" {
		t.Errorf("expected 'from-config', got %q", result)
	}
}

func TestStringValueOrEnv_envFallback(t *testing.T) {
	t.Setenv("TEST_STRING_VALUE_OR_ENV", "from-env")
	val := types.StringNull()
	result := stringValueOrEnv(val, "TEST_STRING_VALUE_OR_ENV")
	if result != "from-env" {
		t.Errorf("expected 'from-env', got %q", result)
	}
}

func TestStringValueOrEnv_unknownFallsToEnv(t *testing.T) {
	t.Setenv("TEST_STRING_VALUE_UNKNOWN", "env-val")
	val := types.StringUnknown()
	result := stringValueOrEnv(val, "TEST_STRING_VALUE_UNKNOWN")
	if result != "env-val" {
		t.Errorf("expected 'env-val', got %q", result)
	}
}

func TestStringValueOrEnv_noConfigNoEnv(t *testing.T) {
	val := types.StringNull()
	result := stringValueOrEnv(val, "TEST_NONEXISTENT_VAR")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
