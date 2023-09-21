package grafana

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/grafana/grizzly/pkg/grizzly"
)

// getRemoteGrafanaRuleGroup retrieves a rule group from Grafana
func getRemoteGrafanaRuleGroup(uid string) (*grizzly.Resource, error) {
	parts := strings.SplitN(uid, ".", 2)
	folder := parts[0]
	name := parts[1]

	client := new(http.Client)
	h := GrafanaRuleGroupHandler{}
	grafanaURL, err := getGrafanaURL(fmt.Sprintf("api/ruler/grafana/api/v1/rules/%s/%s", folder, name))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", grafanaURL, nil)
	if err != nil {
		return nil, err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("not found")
	default:
		if resp.StatusCode >= 400 {
			return nil, errors.New(resp.Status)
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var g RuleGroupConfig
	if err := json.Unmarshal([]byte(string(body)), &g); err != nil {
		return nil, grizzly.APIErr{Err: err, Body: body}
	}

	delete(g, "name")
	for _, r := range g.Rules() {
		rule := r.(map[string]interface{})
		alert := rule["grafana_alert"].(map[string]interface{})
		delete(alert, "namespace_id")
		delete(alert, "namespace_uid")
		delete(alert, "orgId")
		delete(alert, "rule_group")
		delete(alert, "id")
		delete(alert, "updated")
		delete(alert, "version")
	}

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), name, g)
	resource.SetMetadata("folder", folder)
	return &resource, nil
}

func getRemoteGrafanaRuleGroupList() ([]string, error) {
	client := new(http.Client)
	UIDs := []string{}
	grafanaURL, err := getGrafanaURL("api/ruler/grafana/api/v1/rules")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", grafanaURL, nil)
	if err != nil {
		return nil, err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, fmt.Errorf("couldn't fetch rule list from remote: %w", grizzly.ErrNotFound)
	case resp.StatusCode >= 400:
		return nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	groupings := map[string][]GettableRuleGroupConfig{}
	if err := json.Unmarshal([]byte(string(body)), &groupings); err != nil {
		return nil, grizzly.APIErr{Err: err, Body: body}
	}

	for folder, grouping := range groupings {
		for _, group := range grouping {
			uid := fmt.Sprintf("%s.%s", folder, group.Name())
			UIDs = append(UIDs, uid)
		}
	}

	return UIDs, nil
}

func postGrafanaRuleGroup(resource grizzly.Resource) error {
	folder := resource.GetMetadata("folder")

	client := new(http.Client)
	grafanaURL, err := getGrafanaURL(fmt.Sprintf("api/ruler/grafana/api/v1/rules/%s", folder))
	if err != nil {
		return err
	}

	resource.Spec()["name"] = resource.Name()
	bs, err := json.Marshal(resource["spec"])
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", grafanaURL, bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusAccepted:
		return nil
	default:
		r, err := updateResourceToHaveUids(&resource)
		if err != nil {
			return NewErrNon200Response("rules", resource.Name(), resp)
		}
		return postGrafanaRuleGroup(*r)

	}
}

func updateResourceToHaveUids(resource *grizzly.Resource) (*grizzly.Resource, error) {
	folder := resource.GetMetadata("folder")
	r, err := getRemoteGrafanaRuleGroup(folder + "." + resource.Name())
	if err != nil {
		return nil, err
	}
	t := make(map[string]string)
	for _, rule := range r.Spec()["rules"].([]interface{}) {
		if grafana_alert, ok := (rule.(map[string]interface{}))["grafana_alert"]; ok {
			if uid, ok := (grafana_alert.(map[string]interface{}))["uid"]; ok {
				// Find the matching grafana_alert in resource and update it.
				if title, ok := grafana_alert.(map[string]interface{})["title"]; ok {
					t[title.(string)] = uid.(string)
				}
			}
		}
	}
	uidsAdded := false
	for _, rule := range resource.Spec()["rules"].([]interface{}) {
		if grafana_alert, ok := (rule.(map[string]interface{}))["grafana_alert"]; ok {
			// Find the matching grafana_alert in resource and update it.
			g := grafana_alert.(map[string]interface{})
			if _, ok := g["uid"]; ok {
				continue
			}

			if title, ok := g["title"]; ok {
				if uid, ok := t[title.(string)]; ok {
					g["uid"] = uid
					uidsAdded = true
				}
			}
		}
	}
	if !uidsAdded {
		return nil, errors.New("No UIDs to add")
	}
	return resource, nil
}

func deleteGrafanaRuleGroup(uid string) error {
	parts := strings.SplitN(uid, ".", 2)
	folder := parts[0]
	name := parts[1]

	client := new(http.Client)
	grafanaURL, err := getGrafanaURL(fmt.Sprintf("api/ruler/grafana/api/v1/rules/%s/%s", folder, name))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", grafanaURL, nil)
	if err != nil {
		return err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New(resp.Status)
	}
	return nil
}

type RuleGroupConfig map[string]interface{}

func (g *RuleGroupConfig) Rules() []interface{} {
	rules, _ := (*g)["rules"]
	return rules.([]interface{})
}

type GettableExtendedRuleNode map[string]interface{}

func (r *GettableExtendedRuleNode) GrafanaAlert() map[string]interface{} {
	rules, _ := (*r)["grafana_alert"]
	return rules.(map[string]interface{})
}

type GettableRuleGroupConfig map[string]interface{}

func (d *GettableRuleGroupConfig) Name() string {
	name, ok := (*d)["name"]
	if !ok {
		return ""
	}
	return name.(string)
}

type GrafanaRule map[string]interface{}

func (d *GrafanaRule) UID() string {
	uid, ok := (*d)["uid"]
	if !ok {
		return ""
	}
	return uid.(string)
}
