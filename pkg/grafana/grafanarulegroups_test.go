package grafana

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestGrafanaRuleGroups(t *testing.T) {
	os.Setenv("GRAFANA_URL", getUrl())

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&Provider{},
		})

	ticker := pingService(getUrl())
	defer ticker.Stop()

	t.Run("get remote grafana rule group - not found", func(t *testing.T) {
		resource, err := getRemoteGrafanaRuleGroup("ns.dummy")
		require.EqualError(t, err, "not found")
		require.Empty(t, resource)
	})

	t.Run("post remote grafana rule group - success", func(t *testing.T) {
		datasources, err := getRemoteDatasourceList()
		require.NoError(t, err)

		testdata, err := os.ReadFile("testdata/test_json/post_grafanarulegroup.json")
		require.NoError(t, err)

		datasource := datasources[0]
		testdataStr := strings.Replace(string(testdata), "__DATASOURCE__", datasource, -1)

		var resource grizzly.Resource
		err = json.Unmarshal([]byte(testdataStr), &resource)
		require.NoError(t, err)

		err = postGrafanaRuleGroup(resource)
		require.NoError(t, err)

		group, err := getRemoteGrafanaRuleGroup("Azure Data Explorer.rulegroup")
		require.NoError(t, err)
		require.NotNil(t, group)
		require.Equal(t, "1m", group.Spec()["interval"])
		require.Equal(t, "rulegroup", group.Name())
	})

	_ = os.Unsetenv("GRAFANA_URL")
}
