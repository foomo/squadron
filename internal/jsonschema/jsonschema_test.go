package jsonschema_test

import (
	"fmt"
	"testing"

	"github.com/foomo/squadron/internal/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSchema(t *testing.T) {
	// Example usage
	baseURL := "https://raw.githubusercontent.com/foomo/squadron/refs/tags/v2.4.0/squadron.schema.json"
	overrideURL := "https://raw.githubusercontent.com/foomo/helm-charts/refs/tags/namespace-0.1.2/charts/namespace/values.schema.json"

	// Create the JSONSchema object
	js := jsonschema.New()
	err := js.LoadBaseSchema(t.Context(), baseURL)
	require.NoError(t, err)

	// Override the base schema
	err = js.SetSquadronUnitSchema(t.Context(), "site", "namespace", overrideURL)
	require.NoError(t, err)

	// Print the resulting schema
	actual, err := js.String()
	require.NoError(t, err)

	expected := `{"$defs":{"Build":{"additionalProperties":false,"properties":{"add_host":{"description":"AddHost add a custom host-to-IP mapping (format: \"host:ip\")","items":{"type":"string"},"type":"array"},"allow":{"description":"Allow extra privileged entitlement (e.g., \"network.host\", \"security.insecure\")","items":{"type":"string"},"type":"array"},"attest":{"description":"Attest parameters (format: \"type=sbom,generator=image\")","items":{"type":"string"},"type":"array"},"build_arg":{"description":"BuildArg set build-time variables","items":{"type":"string"},"type":"array"},"build_context":{"description":"BuildContext additional build contexts (e.g., name=path)","items":{"type":"string"},"type":"array"},"builder":{"description":"Builder override the configured builder instance","type":"string"},"cache_from":{"description":"CacheFrom external cache sources (e.g., \"user/app:cache\", \"type=local,src=path/to/dir\")","type":"string"},"cache_to":{"description":"CacheTo cache export destinations (e.g., \"user/app:cache\", \"type=local,dest=path/to/dir\")","type":"string"},"cgroup_parent":{"description":"CGroupParent optional parent cgroup for the container","type":"string"},"context":{"description":"Build context","type":"string"},"dependencies":{"description":"Dependencies list of build names defined in the squadron configuration","items":{"type":"string"},"type":"array"},"file":{"description":"File name of the Dockerfile (default: \"PATH/Dockerfile\")","type":"string"},"iidfile":{"description":"IIDFile write the image ID to the file","type":"string"},"image":{"description":"Image name","type":"string"},"label":{"description":"Label wet metadata for an image","items":{"type":"string"},"type":"array"},"load":{"description":"Load shorthand for \"--output=type=docker\"","type":"boolean"},"metadata_file":{"description":"MetadataFile write build result metadata to the file","type":"string"},"network":{"description":"Network set the networking mode for the \"RUN\" instructions during build (default \"default\")","type":"string"},"no_cache":{"description":"NoCache do not use cache when building the image","type":"boolean"},"no_cache_filter":{"description":"NoCacheFilter do not cache specified stages","items":{"type":"string"},"type":"array"},"output":{"description":"Output destination (format: \"type=local,dest=path\")","type":"string"},"platform":{"description":"Platform set target platform for build","type":"string"},"pull":{"description":"Always attempt to pull all referenced images","type":"boolean"},"push":{"description":"Shorthand for \"--output=type=registry\"","type":"boolean"},"quiet":{"description":"Suppress the build output and print image ID on succes","type":"boolean"},"secret":{"description":"Secret to expose to the build (format: \"id=mysecret[,src=/local/secret]\")","items":{"type":"string"},"type":"array"},"shm_size":{"description":"ShmSize size of \"/dev/shm\"","type":"string"},"ssh":{"description":"SSH agent socket or keys to expose to the build (format: \"default|\u003cid\u003e[=\u003csocket\u003e|\u003ckey\u003e[,\u003ckey\u003e]]\")","type":"string"},"tag":{"description":"Tag name and optionally a tag (format: \"name:tag\")","type":"string"},"target":{"description":"Target set the target build stage to build","type":"string"},"ulimit":{"description":"ULimit ulimit options (default [])","type":"string"}},"type":"object"},"Chart":{"additionalProperties":false,"properties":{"alias":{"description":"Chart alias","type":"string"},"name":{"description":"Chart name","type":"string"},"repository":{"description":"Chart repository","type":"string"},"schema":{"description":"Values schema json","type":"string"},"version":{"description":"Chart version","type":"string"}},"type":"object"},"Config":{"additionalProperties":false,"properties":{"builds":{"additionalProperties":{"$ref":"#/$defs/Build"},"description":"Global builds that can be referenced as dependencies","type":"object"},"global":{"description":"Global values to be injected into all squadron values","type":"object"},"squadron":{"additionalProperties":{"additionalProperties":{"$ref":"#/$defs/Unit"},"type":"object"},"description":"Squadron definitions","properties":{"site":{"additionalProperties":{"$ref":"#/$defs/Unit"},"properties":{"namespace":{"anyOf":[{"$ref":"#/$defs/Unit"},{"properties":{"values":{"$ref":"#/$defs/raw.githubusercontent.com-foomo-helm-charts-refs-tags-namespace-0.1.2-charts-namespace-values.schema.json"}},"type":"object"}]}},"type":"object"}},"type":"object"},"vars":{"description":"Global values to be injected into all squadron values","type":"object"},"version":{"description":"Version of the schema","pattern":"^[0-9]\\.[0-9]$","type":"string"}},"required":["version"],"type":"object"},"Tags":{"items":{"type":"string"},"type":"array"},"Unit":{"additionalProperties":false,"properties":{"builds":{"additionalProperties":{"$ref":"#/$defs/Build"},"description":"Map of containers to build","type":"object"},"chart":{"anyOf":[{"type":"string"},{"$ref":"#/$defs/Chart"}],"description":"Chart settings"},"extends":{"description":"Extend chart values","type":"string"},"kustomize":{"description":"Kustomize files path","type":"string"},"tags":{"$ref":"#/$defs/Tags","description":"List of tags"},"values":{"description":"Chart values","type":"object"}},"type":"object"},"raw.githubusercontent.com-foomo-helm-charts-refs-tags-namespace-0.1.2-charts-namespace-values.schema.json":{"properties":{"fullnameOverride":{"type":"string"},"nameOverride":{"type":"string"},"namespaceOverride":{"type":"string"},"secrets":{"properties":{"dockerConfigs":{"type":"object"},"opaque":{"type":"object"},"tls":{"type":"object"}},"type":"object"},"serviceAccounts":{"type":"object"}},"type":"object"}},"$id":"https://github.com/foomo/squadron/internal/config/config","$ref":"#/$defs/Config","$schema":"https://json-schema.org/draft/2020-12/schema"}`
	if !assert.JSONEq(t, expected, actual) {
		fmt.Println(actual)
	}
}
