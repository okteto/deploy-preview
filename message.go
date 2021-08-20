package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

//Endpoint represents an Okteto statefulset
type Endpoint struct {
	URL     string `json:"url"`
	Divert  bool   `json:"divert"`
	Private bool   `json:"private"`
}

func main() {
	previewName := os.Args[1]
	previewCommandExitCode := os.Args[2]

	oktetoURL := getOktetoURL()
	previewURL := fmt.Sprintf("%s/#/previews/%s", oktetoURL, previewName)

	var firstLine string
	if previewCommandExitCode == "0" {
		firstLine = fmt.Sprintf("Your preview environment [%s](%s) has been deployed.", previewName, previewURL)
	} else {
		firstLine = fmt.Sprintf("Your preview environment [%s](%s) has been deployed with errors.", previewName, previewURL)
	}
	fmt.Println(firstLine)

	endpoints, err := getEndpoints(previewName)
	if err != nil {
		return
	}
	if len(endpoints) == 1 {
		fmt.Printf("\n  Preview environment endpoint is available [here](%s)", endpoints[0])
	} else if len(endpoints) > 1 {
		endpoints = translateEndpoints(endpoints)
		fmt.Printf("\n  Preview environment endpoints are available at:")
		for _, endpoint := range endpoints {
			fmt.Printf("\n  * %s", endpoint)
		}
	}

}

func getOktetoURL() string {
	if t := getToken(); t != nil {
		return t.URL
	}
	return ""
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

func getEndpoints(name string) ([]string, error) {
	cmd := exec.Command("okteto", "preview", "endpoints", name, "-o", "json")
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
	endpointURLs := make([]string, 0)
	for _, e := range endpoints {
		endpointURLs = append(endpointURLs, e.URL)
	}
	return endpointURLs, nil
}

func translateEndpoints(endpoints []string) []string {
	result := make([]string, 0)
	for _, endpoint := range endpoints {
		result = append(result, fmt.Sprintf("[%s](%s)", endpoint, endpoint))
	}
	return result
}
