package grafana

import (
	"fmt"
	"path/filepath"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// GrafanaRuleGroupHandler is a Grizzly Handler for Grafana alert rules
type GrafanaRuleGroupHandler struct {
	Provider Provider
}

// NewGrafanaRuleGroupHandler returns configuration defining a new Grafana Rule Group Handler
func NewGrafanaRuleGroupHandler(provider Provider) *GrafanaRuleGroupHandler {
	return &GrafanaRuleGroupHandler{
		Provider: provider,
	}
}

// Kind returns the kind for this handler
func (h *GrafanaRuleGroupHandler) Kind() string {
	return "GrafanaRuleGroup"
}

// Validate returns the uid of resource
func (h *GrafanaRuleGroupHandler) Validate(resource grizzly.Resource) error {
	return nil
}

// APIVersion returns group and version of the provider of this resource
func (h *GrafanaRuleGroupHandler) APIVersion() string {
	return h.Provider.APIVersion()
}

// GetExtension returns the file name extension for a datasource
func (h *GrafanaRuleGroupHandler) GetExtension() string {
	return "yaml"
}

const (
	grafanaruleGlob    = "grafanarulegroups/grafanarulegroup-*"
	grafanarulePattern = "grafanarulegroups/grafanarulegroup-%s.%s"
)

// FindResourceFiles identifies files within a directory that this handler can process
func (h *GrafanaRuleGroupHandler) FindResourceFiles(dir string) ([]string, error) {
	path := filepath.Join(dir, grafanaruleGlob)
	return filepath.Glob(path)
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *GrafanaRuleGroupHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(grafanarulePattern, resource.UID(), filetype)
}

// Parse parses a manifest object into a struct for this resource type
func (h *GrafanaRuleGroupHandler) Parse(m manifest.Manifest) (grizzly.Resources, error) {
	resource := grizzly.Resource(m)
	return grizzly.Resources{resource}, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *GrafanaRuleGroupHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *GrafanaRuleGroupHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetUID returns the UID for a resource
func (h *GrafanaRuleGroupHandler) GetUID(resource grizzly.Resource) (string, error) {
	if !resource.HasMetadata("folder") {
		return "", fmt.Errorf("%s %s requires a folder metadata entry", h.Kind(), resource.Name())
	}
	return fmt.Sprintf("%s.%s", resource.GetMetadata("folder"), resource.Name()), nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID (name)
func (h *GrafanaRuleGroupHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	return getRemoteGrafanaRuleGroup(UID)
}

// GetRemote retrieves a rule group as a Resource
func (h *GrafanaRuleGroupHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	uid := fmt.Sprintf("%s.%s", resource.GetMetadata("folder"), resource.Name())
	return getRemoteGrafanaRuleGroup(uid)
}

// ListRemote retrieves as list of UIDs (names) of all remote rule groups
func (h *GrafanaRuleGroupHandler) ListRemote() ([]string, error) {
	return getRemoteGrafanaRuleGroupList()
}

// Add pushes a rule group to Grafana via the API
func (h *GrafanaRuleGroupHandler) Add(resource grizzly.Resource) error {
	return postGrafanaRuleGroup(resource)
}

// Update pushes a rule group to Grafana via the API
func (h *GrafanaRuleGroupHandler) Update(existing, resource grizzly.Resource) error {
	return postGrafanaRuleGroup(resource)
}
