package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/machinebox/graphql"
)

type Token struct {
	URL       string `json:"URL"`
	Buildkit  string `json:"Buildkit"`
	Registry  string `json:"Registry"`
	ID        string `json:"ID"`
	Username  string `json:"Username"`
	Token     string `json:"Token"`
	MachineID string `json:"MachineID"`
}

type PreviewBody struct {
	Preview Preview `json:"preview"`
}

// Preview represents the contents of an Okteto Cloud space
type Preview struct {
	GitDeploys   []PipelineRun `json:"gitDeploys"`
	Statefulsets []Statefulset `json:"statefulsets"`
	Deployments  []Deployment  `json:"deployments"`
}

//PipelineRun represents an Okteto pipeline status
type PipelineRun struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Status     string `json:"status"`
}

//Statefulset represents an Okteto statefulset
type Statefulset struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Endpoints []Endpoint `json:"endpoints"`
	Status    string     `json:"status"`
}

//Deployment represents an Okteto statefulset
type Deployment struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Endpoints []Endpoint `json:"endpoints"`
	Status    string     `json:"status"`
}

//Endpoint represents an Okteto statefulset
type Endpoint struct {
	URL string `json:"url"`
}

func main() {
	previewName := os.Args[1]

	oktetoURL := getOktetoURL()
	previewURL := fmt.Sprintf("%s/#/previews/%s", oktetoURL, previewName)
	waitForResourcesRunning(previewName)

	endpoints := getEndpoints(previewName)

	var firstLine string

	firstLine = fmt.Sprintf("Your preview environment [%s](%s) has been deployed.", previewName, previewURL)
	fmt.Println(firstLine)

	if len(endpoints) == 1 {
		fmt.Printf("\n  Preview environment endpoint is available [here](%s)", endpoints[0])
	} else if len(endpoints) > 1 {
		endpoints = translateEndpoints(endpoints)
		fmt.Printf("\n  Preview environment endpoints are available at:\n    - %s", strings.Join(endpoints, "    - "))
	}

}
func getOktetoURL() string {
	if t := getToken(); t != nil {
		return t.URL
	}
	return ""
}

func getEndpoints(previewName string) []string {
	endpoints := make([]string, 0)

	q := fmt.Sprintf(`query{
		preview(id: "%s"){
			deployments{
				endpoints{
					url
				} 
			},
			statefulsets{
				endpoints{
					url
				}
			}
		}
	}`, previewName)
	var body PreviewBody
	if err := query(q, &body); err != nil {
		return []string{}
	}

	for _, d := range body.Preview.Deployments {
		for _, endpoint := range d.Endpoints {
			endpoints = append(endpoints, endpoint.URL)
		}
	}

	for _, sfs := range body.Preview.Statefulsets {
		for _, endpoint := range sfs.Endpoints {
			endpoints = append(endpoints, endpoint.URL)
		}
	}
	return endpoints
}

func query(query string, result interface{}) error {
	var token string
	if t := getToken(); t != nil {
		token = t.Token
	}
	ctx := context.Background()
	url := fmt.Sprintf("%s/graphql", getOktetoURL())

	c := graphql.NewClient(url)

	req := graphql.NewRequest(query)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	if err := c.Run(ctx, req, result); err != nil {
		fmt.Print(err)
	}
	return nil
}

func getToken() *Token {
	tokenPath := filepath.Join(os.Getenv("HOME"), ".okteto", ".token.json")
	b, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return nil
	}

	token := &Token{}
	if err := json.Unmarshal(b, token); err != nil {
		return nil
	}
	return token
}

func translateEndpoints(endpoints []string) []string {
	result := make([]string, 0)
	for _, endpoint := range endpoints {
		result = append(result, fmt.Sprintf("[%s](%s)", endpoint, endpoint))
	}
	return result
}

func areAllServicesRunning(servicesRunningStatus map[string]bool) bool {
	for _, isRunning := range servicesRunningStatus {
		if !isRunning {
			return false
		}
	}
	return true
}

func getServicesNotRunning(servicesRunningStatus map[string]bool) []string {
	servicesNotRunnig := make([]string, 0)
	for svcName, isRunning := range servicesRunningStatus {
		if !isRunning {
			servicesNotRunnig = append(servicesNotRunnig, svcName)
		}
	}
	return servicesNotRunnig
}

func waitForResourcesRunning(previewName string) error {
	exit := false

	ticker := time.NewTicker(5 * time.Second)
	timeout := time.Now().Add(300 * time.Second)
	for time.Now().Before(timeout) {
		<-ticker.C
		resourceStatus, err := getResourceStatus(previewName)
		if err != nil {
			return err
		}
		for _, status := range resourceStatus {
			if status != "running" {
				exit = true
			}
		}
		if !exit {
			break
		}
	}
	return nil
}

func getResourceStatus(previewName string) (map[string]string, error) {
	status := make(map[string]string)
	q := fmt.Sprintf(`query{
		preview(id: "%s"){
			deployments{
				name, status
			},
			statefulsets{
				name, status
			}
		}
	}`, previewName)
	var body PreviewBody
	if err := query(q, &body); err != nil {
		return status, err
	}

	for _, d := range body.Preview.Deployments {
		status[d.Name] = d.Status
	}

	for _, sfs := range body.Preview.Statefulsets {
		status[sfs.Name] = sfs.Status
	}
	return status, nil
}
