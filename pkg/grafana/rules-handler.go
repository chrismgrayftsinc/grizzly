package grafana

import (
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// RuleHandler is a Grizzly Handler for Prometheus Rules
type RuleHandler struct {
	Provider Provider
}

// NewRuleHandler returns a new Grizzly Handler for Prometheus Rules
func NewRuleHandler(provider Provider) *RuleHandler {
	return &RuleHandler{
		Provider: provider,
	}
}

// Kind returns the name for this handler
func (h *RuleHandler) Kind() string {
	return "PrometheusRuleGroup"
}

// APIVersion returns the group and version for the provider of which this handler is a part
func (h *RuleHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a rule grouping
func (h *RuleHandler) GetExtension() string {
	return "yaml"
}

func (h *RuleHandler) newRuleGroupingResource(m manifest.Manifest) grizzly.Resource {
	resource := grizzly.Resource{
		UID:     m.Metadata().Name(),
		Handler: h,
		Detail:  m,
	}
	return resource
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *RuleHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *RuleHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetRemoteByUID retrieves a dashboard as a resource
func (h *RuleHandler) GetRemoteByUID(uid string) (*grizzly.Resource, error) {
	m, err := getRemoteRuleGroup(uid)
	if err != nil {
		return nil, err
	}
	return grizzly.NewResource(*m, h), nil
}

// GetRemote retrieves a dashboard as a resource
func (h *RuleHandler) GetRemote(existing grizzly.Resource) (*grizzly.Resource, error) {
	uid := fmt.Sprintf("%s.%s", existing.Detail.Metadata().Name(), existing.Detail.Metadata().Namespace())
	return h.GetRemoteByUID(uid)
}

// Add pushes a datasource to Grafana via the API
func (h *RuleHandler) Add(resource grizzly.Resource) error {
	return writeRuleGroup(resource.Detail)
}

// Update pushes a datasource to Grafana via the API
func (h *RuleHandler) Update(existing, resource grizzly.Resource) error {
	return writeRuleGroup(resource.Detail)
}