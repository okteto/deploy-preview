package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed default-message.md.gotmpl
var defaultCommentTemplate string

type contexts struct {
	Current  string             `json:"current-context"`
	Contexts map[string]context `json:"contexts"`
}

type context struct {
	Name *url.URL `json:"name"`
}

func (p *context) UnmarshalJSON(data []byte) error {
	type Context context

	tmp := struct {
		Name string `json:"name"`
		*Context
	}{
		Context: (*Context)(p),
	}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	p.Name, err = url.Parse(tmp.Name)
	if err != nil {
		return err
	}

	return nil
}

// Endpoint represents an Okteto statefulset
type Endpoint struct {
	URL     *url.URL `json:"url"`
	Divert  bool     `json:"divert"`
	Private bool     `json:"private"`
}

func (p *Endpoint) UnmarshalJSON(data []byte) error {
	type endpoint Endpoint

	tmp := struct {
		URL string `json:"url"`
		*endpoint
	}{
		endpoint: (*endpoint)(p),
	}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	p.URL, err = url.Parse(tmp.URL)
	if err != nil {
		return err
	}

	return nil
}

func generateMessage(previewName string, previewSucceeded bool, commentTemplate string) (string, error) {
	oktetoURL, err := getOktetoURL()
	if err != nil {
		return "", err
	}

	commentTemplate, err = getCommentTemplate(commentTemplate)
	if err != nil {
		return "", err
	}

	endpoints, err := getEndpoints(previewName)
	if err != nil {
		return "", err
	}

	previewURLSuffix := fmt.Sprintf("%s.%s", previewName, oktetoURL.Host)
	templateVars := map[string]interface{}{
		"OktetoURL":        oktetoURL.String(),
		"PreviewURL":       fmt.Sprintf("%s/#/previews/%s", oktetoURL, previewName),
		"PreviewName":      previewName,
		"PreviewURLSuffix": previewURLSuffix,
		"PreviewSuccess":   previewSucceeded,
		"Endpoints":        endpoints,
		"EndpointsMap":     getEndpointsMap(previewURLSuffix, endpoints),
	}

	return parseTemplate(commentTemplate, templateVars)
}

func getCommentTemplate(commentTemplate string) (string, error) {
	if commentTemplate == "" {
		commentTemplate = defaultCommentTemplate
	}

	if commentTemplate[0:1] == "@" {
		file, err := os.Open(commentTemplate[1:])
		if err != nil {
			return "", err
		}

		fileContents, err := io.ReadAll(file)
		if err != nil {
			return "", err
		}

		return string(fileContents), nil
	}

	return commentTemplate, nil
}

func getEndpointsMap(previewURLSuffix string, endpoints []*url.URL) map[string]string {
	endpointsMap := make(map[string]string, len(endpoints))
	for _, endpoint := range endpoints {
		name := strings.TrimSuffix(endpoint.Host, "-"+previewURLSuffix)
		endpointsMap[name] = endpoint.String()
	}

	return endpointsMap
}

func getOktetoURL() (*url.URL, error) {
	contextsPath := filepath.Join(os.Getenv("HOME"), ".okteto", "context", "config.json")
	b, err := os.ReadFile(contextsPath)
	if err != nil {
		return nil, err
	}

	contexts := &contexts{}
	if err := json.Unmarshal(b, contexts); err != nil {
		return nil, err
	}

	if val, ok := contexts.Contexts[contexts.Current]; ok {
		return val.Name, nil
	}

	return nil, fmt.Errorf("context %s is missing", contexts.Current)
}

func getEndpoints(name string) ([]*url.URL, error) {
	cmd := exec.Command("/home/hsadiq/.zinit/plugins/okteto---okteto/okteto", "preview", "endpoints", name, "-o", "json")
	cmd.Env = os.Environ()
	o, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var endpoints []Endpoint
	err = json.Unmarshal(o, &endpoints)
	if err != nil {
		return nil, err
	}
	endpointURLs := make([]*url.URL, 0, len(endpoints))
	for _, e := range endpoints {
		endpointURLs = append(endpointURLs, e.URL)
	}
	return endpointURLs, nil
}

func parseTemplate(templateText string, vars map[string]interface{}) (string, error) {
	var output bytes.Buffer

	tmpl, err := template.New("template").Funcs(map[string]interface{}{
		"Contains":  strings.Contains,
		"HasPrefix": strings.HasPrefix,
		"HasSuffix": strings.HasSuffix,
		"Title":     cases.Title(language.English).String,
		"Trim":      strings.TrimSpace,
	}).Parse(templateText)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&output, vars)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
