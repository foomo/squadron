package jsonschema_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/foomo/squadron/internal/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerger(t *testing.T) {
	// Example usage
	baseURL := "https://raw.githubusercontent.com/foomo/squadron/refs/heads/main/squadron.schema.json"
	overrideURL := "https://raw.githubusercontent.com/foomo/helm-charts/refs/heads/main/charts/namespace/values.schema.json"

	// Create the JSONSchema object
	js := jsonschema.New()
	err := js.LoadBaseSchema(context.TODO(), baseURL)
	require.NoError(t, err)

	// Override the base schema
	err = js.SetSquadronUnitSchema(context.TODO(), "site", "namespace", overrideURL)
	require.NoError(t, err)

	// Print the resulting schema
	actual, err := js.String()
	require.NoError(t, err)

	expected := `{"$defs":{"Build":{"additionalProperties":false,"properties":{"add_host":{"items":{"type":"string"},"type":"array"},"allow":{"items":{"type":"string"},"type":"array"},"attest":{"items":{"type":"string"},"type":"array"},"build_arg":{"items":{"type":"string"},"type":"array"},"build_context":{"items":{"type":"string"},"type":"array"},"builder":{"type":"string"},"cache_from":{"type":"string"},"cache_to":{"type":"string"},"cgroup_parent":{"type":"string"},"context":{"type":"string"},"dependencies":{"items":{"type":"string"},"type":"array"},"file":{"type":"string"},"iidfile":{"type":"string"},"image":{"type":"string"},"label":{"items":{"type":"string"},"type":"array"},"load":{"type":"boolean"},"metadata_file":{"type":"string"},"network":{"type":"string"},"no_cache":{"type":"boolean"},"no_cache_filter":{"items":{"type":"string"},"type":"array"},"output":{"type":"string"},"platform":{"type":"string"},"secret":{"items":{"type":"string"},"type":"array"},"shm_size":{"type":"string"},"ssh":{"type":"string"},"tag":{"type":"string"},"target":{"type":"string"},"ulimit":{"type":"string"}},"type":"object"},"Chart":{"additionalProperties":false,"properties":{"alias":{"type":"string"},"name":{"type":"string"},"repository":{"type":"string"},"version":{"type":"string"}},"type":"object"},"Config":{"additionalProperties":false,"properties":{"builds":{"additionalProperties":{"$ref":"#/$defs/Build"},"type":"object"},"global":{"type":"object"},"squadron":{"additionalProperties":{"additionalProperties":{"$ref":"#/$defs/Unit"},"type":"object"},"properties":{"site":{"additionalProperties":{"$ref":"#/$defs/Unit"},"properties":{"namespace":{"anyOf":[{"$ref":"#/$defs/Unit"},{"properties":{"values":{"$ref":"#/$defs/raw.githubusercontent.com-foomo-helm-charts-refs-heads-main-charts-namespace-values.schema.json"}},"type":"object"}]}},"type":"object"}},"type":"object"},"vars":{"type":"object"},"version":{"type":"string"}},"required":["version"],"type":"object"},"Tags":{"items":{"type":"string"},"type":"array"},"Unit":{"additionalProperties":false,"properties":{"builds":{"additionalProperties":{"$ref":"#/$defs/Build"},"type":"object"},"chart":{"anyOf":[{"type":"string"},{"$ref":"#/$defs/Chart"}]},"kustomize":{"type":"string"},"tags":{"$ref":"#/$defs/Tags"},"values":{"type":"object"}},"type":"object"},"raw.githubusercontent.com-foomo-helm-charts-refs-heads-main-charts-namespace-values.schema.json":{"properties":{"fullnameOverride":{"type":"string"},"nameOverride":{"type":"string"},"namespaceOverride":{"type":"string"},"secrets":{"properties":{"dockerConfigs":{"type":"object"},"opaque":{"type":"object"},"tls":{"type":"object"}},"type":"object"},"serviceAccounts":{"type":"object"}},"type":"object"}},"$id":"https://github.com/foomo/squadron/internal/config/config","$ref":"#/$defs/Config","$schema":"https://json-schema.org/draft/2020-12/schema"}`
	if !assert.JSONEq(t, expected, actual) {
		fmt.Println(actual)
	}
}
