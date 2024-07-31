package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type GitHubEvent struct {
	Number        int `json:"number"`
	ClientPayload struct {
		PullRequest struct {
			Number int `json:"number"`
		} `json:"pull_request"`
	} `json:"client_payload"`
}

type GitHubComment struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type github struct {
	repository    string
	sourceURL     string
	prNumber      int
	defaultBranch string
}

type ciInfo interface {
	Repository() string
	SourceURL() string
	DefaultBranch() string
	Notify(message string) error
}

var _ ciInfo = (*github)(nil)

func newGitHub() (ciInfo, error) {
	event, err := os.Open(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		log.Printf("failed to read GITHUB_EVENT_PATH: %s", os.Getenv("GITHUB_EVENT_PATH"))
		return nil, err
	}
	defer event.Close()

	var payload GitHubEvent
	err = json.NewDecoder(event).Decode(&payload)
	if err != nil {
		log.Printf("failed to parse GITHUB_EVENT_PATH: %s", os.Getenv("GITHUB_EVENT_PATH"))
		return nil, err
	}

	gh := &github{}

	switch os.Getenv("GITHUB_EVENT_NAME") {
	case "pull_request":

		gh.prNumber = payload.Number
		gh.defaultBranch = os.Getenv("GITHUB_HEAD_REF")
	case "repository_dispatch":
		gh.prNumber = payload.ClientPayload.PullRequest.Number
		gh.defaultBranch = strings.TrimPrefix(os.Getenv("GITHUB_REF"), "refs/heads/")
	}

	gh.repository = fmt.Sprintf("%s/%s", os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY"))
	gh.sourceURL = fmt.Sprintf("%s/%s", gh.repository, gh.prNumber)

	return gh, nil
}

func (gh *github) Repository() string {
	return gh.repository
}

func (gh *github) SourceURL() string {
	return gh.sourceURL
}

func (gh *github) DefaultBranch() string {
	return gh.defaultBranch
}

func (gh *github) PRNumber() int {
	return gh.prNumber
}

func (gh *github) Notify(message string) error {
	githubToken := os.Getenv("GITHUB_TOKEN")
	githubRepository := os.Getenv("GITHUB_REPOSITORY")

	if githubToken == "" {
		log.Println("failed to set message as no GITHUB_TOKEN found")
		return errors.New("missing GITHUB_TOKEN")
	}

	resp, err := gh.callGitHub(githubToken, "GET", githubRepository, nil, "issues", fmt.Sprintf("%d", gh.prNumber), "comments")
	if err != nil {
		return fmt.Errorf("failed to retrieve PR comments: %s", err)
	}

	defer resp.Body.Close()

	var comments []GitHubComment
	json.NewDecoder(resp.Body).Decode(&comments)

	identifier := gh.previewIdentifier()
	var comment *GitHubComment
	for _, c := range comments {
		if strings.Contains(c.Body, identifier) {
			comment = &c
			break
		}
	}

	msgBodyBuf, err := gh.getGitHubCommentMessage(message)
	if err != nil {
		return err
	}
	if comment == nil {
		_, err = gh.callGitHub(githubToken, "POST", githubRepository, msgBodyBuf, "issues", fmt.Sprintf("%d", gh.prNumber), "comments")
		if err != nil {
			return fmt.Errorf("failed to create new comment: %s", err)
		}
		return nil
	}

	fmt.Println("Message already exists in the PR. Updating")
	_, err = gh.callGitHub(githubToken, "PATCH", githubRepository, msgBodyBuf, "issues", "comments", fmt.Sprintf("%d", comment.ID))
	if err != nil {
		return fmt.Errorf("failed to update comment: %s", err)
	}
	return nil
}

func (gh *github) getGitHubCommentMessage(message string) (*bytes.Buffer, error) {
	msgBody := struct {
		Body string `json:"body"`
	}{
		Body: message + gh.previewIdentifier(),
	}

	msgBodyStr, err := json.Marshal(msgBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode message body: %v", msgBody)
	}

	return bytes.NewBuffer(msgBodyStr), nil
}

func (gh *github) callGitHub(token string, method string, repo string, body io.Reader, path ...string) (*http.Response, error) {
	path = append([]string{repo}, path...)
	uri, _ := url.JoinPath("https://api.github.com/repos/", path...)
	req, _ := http.NewRequest(method, uri, body)
	req.Header.Set("Authorization", "token "+token)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode < http.StatusOK || r.StatusCode >= http.StatusBadRequest {
		var errResp = struct {
			Message string `json:"message"`
		}{}
		err = json.NewDecoder(r.Body).Decode(&errResp)
		if err != nil {
			return nil, fmt.Errorf("failed to decode body: %s", err)
		}

		return nil, errors.New(errResp.Message)
	}

	return r, nil
}

func (gh *github) previewIdentifier() string {
	return fmt.Sprintf("<!-- okteto-preview %d -->", gh.prNumber)
}
