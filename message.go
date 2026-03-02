package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

type contexts struct {
	Current  string                  `json:"current-context"`
	Contexts map[string]oktetoContext `json:"contexts"`
}

type oktetoContext struct {
	Name string `json:"name"`
}

// Endpoint represents an Okteto endpoint
type Endpoint struct {
	URL     string `json:"url"`
	Divert  bool   `json:"divert"`
	Private bool   `json:"private"`
}

// GitHubEvent represents the GitHub event payload
type GitHubEvent struct {
	Number        int                    `json:"number"`
	ClientPayload *ClientPayload         `json:"client_payload,omitempty"`
}

// ClientPayload represents the client_payload in repository_dispatch events
type ClientPayload struct {
	PullRequest *PullRequestPayload `json:"pull_request,omitempty"`
}

// PullRequestPayload represents the pull request info in client_payload
type PullRequestPayload struct {
	Number int `json:"number"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		handleGenerate()
	case "notify":
		handleNotify()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  message generate <preview-name> <exit-code>")
	fmt.Println("  message notify <message> <github-token> <preview-name>")
}

func handleGenerate() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: message generate <preview-name> <exit-code>")
		os.Exit(1)
	}

	previewName := os.Args[2]
	exitCode := os.Args[3]

	message, err := generateMessage(previewName, exitCode)
	if err != nil {
		fmt.Printf("Error generating message: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(message)

	// If GITHUB_TOKEN is set, automatically notify the PR
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken != "" {
		if err := notifyPR(message, githubToken, previewName); err != nil {
			fmt.Printf("Error notifying PR: %v\n", err)
			os.Exit(1)
		}
	}
}

func handleNotify() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: message notify <message> <github-token> <preview-name>")
		os.Exit(1)
	}

	message := os.Args[2]
	githubToken := os.Args[3]
	previewName := os.Args[4]

	if err := notifyPR(message, githubToken, previewName); err != nil {
		fmt.Printf("Error notifying PR: %v\n", err)
		os.Exit(1)
	}
}

func generateMessage(previewName, exitCode string) (string, error) {
	oktetoURL, err := getOktetoURL()
	if err != nil {
		return "", err
	}

	previewURL := fmt.Sprintf("%s/previews/%s", oktetoURL, previewName)

	var message strings.Builder
	if exitCode == "0" {
		message.WriteString(fmt.Sprintf("Your preview environment [%s](%s) has been deployed.", previewName, previewURL))
	} else {
		message.WriteString(fmt.Sprintf("Your preview environment [%s](%s) has been deployed with errors.", previewName, previewURL))
	}

	endpoints, err := getEndpoints(previewName)
	if err != nil {
		return message.String(), nil
	}

	if len(endpoints) == 1 {
		message.WriteString(fmt.Sprintf("\n  Preview environment endpoint is available [here](%s)", endpoints[0]))
	} else if len(endpoints) > 1 {
		endpoints = translateEndpoints(endpoints)
		message.WriteString("\n  Preview environment endpoints are available at:")
		for _, endpoint := range endpoints {
			message.WriteString(fmt.Sprintf("\n  * %s", endpoint))
		}
	}

	return message.String(), nil
}

func notifyPR(message, githubToken, previewName string) error {
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	if eventName != "pull_request" && eventName != "repository_dispatch" {
		return fmt.Errorf("this action only supports either pull_request or repository_dispatch events")
	}

	if githubToken == "" {
		return fmt.Errorf("missing GITHUB_TOKEN")
	}

	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return fmt.Errorf("missing GITHUB_REPOSITORY")
	}

	prNumber, err := getPRNumber()
	if err != nil {
		return err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	repoParts := strings.Split(repo, "/")
	if len(repoParts) != 2 {
		return fmt.Errorf("invalid GITHUB_REPOSITORY format: %s", repo)
	}
	owner := repoParts[0]
	repoName := repoParts[1]

	// List all comments on the PR
	comments, _, err := client.Issues.ListComments(ctx, owner, repoName, prNumber, nil)
	if err != nil {
		return fmt.Errorf("error listing comments: %w", err)
	}

	// Find existing comment
	var existingComment *github.IssueComment
	markerText := fmt.Sprintf("Your preview environment")
	previewMarker := fmt.Sprintf("[%s]", previewName)

	for _, comment := range comments {
		if comment.Body != nil &&
			strings.HasPrefix(*comment.Body, markerText) &&
			strings.Contains(*comment.Body, previewMarker) {
			existingComment = comment
			break
		}
	}

	if existingComment != nil {
		fmt.Println("Message already exists in the PR. Updating")
		_, _, err = client.Issues.EditComment(ctx, owner, repoName, *existingComment.ID, &github.IssueComment{
			Body: &message,
		})
		return err
	}

	// Create new comment
	_, _, err = client.Issues.CreateComment(ctx, owner, repoName, prNumber, &github.IssueComment{
		Body: &message,
	})
	return err
}

func getPRNumber() (int, error) {
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return 0, fmt.Errorf("missing GITHUB_EVENT_PATH")
	}

	data, err := os.ReadFile(eventPath)
	if err != nil {
		return 0, fmt.Errorf("error reading event file: %w", err)
	}

	var event GitHubEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return 0, fmt.Errorf("error parsing event JSON: %w", err)
	}

	eventName := os.Getenv("GITHUB_EVENT_NAME")
	if eventName == "pull_request" {
		return event.Number, nil
	} else if eventName == "repository_dispatch" {
		if event.ClientPayload != nil && event.ClientPayload.PullRequest != nil {
			return event.ClientPayload.PullRequest.Number, nil
		}
		return 0, fmt.Errorf("missing pull request number in repository_dispatch event")
	}

	return 0, fmt.Errorf("unsupported event type: %s", eventName)
}

func getOktetoURL() (string, error) {
	contextsPath := filepath.Join(os.Getenv("HOME"), ".okteto", "context", "config.json")
	b, err := os.ReadFile(contextsPath)
	if err != nil {
		return "", err
	}

	contexts := &contexts{}
	if err := json.Unmarshal(b, contexts); err != nil {
		return "", err
	}

	if val, ok := contexts.Contexts[contexts.Current]; ok {
		return val.Name, nil
	}

	return "", fmt.Errorf("context %s is missing", contexts.Current)
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
