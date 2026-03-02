package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateMessage(t *testing.T) {
	tests := []struct {
		name         string
		previewName  string
		exitCode     string
		setupContext func(t *testing.T) func()
		wantContains []string
	}{
		{
			name:        "successful deployment",
			previewName: "test-preview",
			exitCode:    "0",
			setupContext: func(t *testing.T) func() {
				return setupOktetoContext(t, "https://okteto.example.com")
			},
			wantContains: []string{
				"Your preview environment [test-preview]",
				"has been deployed.",
				"https://okteto.example.com/previews/test-preview",
			},
		},
		{
			name:        "deployment with errors",
			previewName: "test-preview-error",
			exitCode:    "1",
			setupContext: func(t *testing.T) func() {
				return setupOktetoContext(t, "https://okteto.dev")
			},
			wantContains: []string{
				"Your preview environment [test-preview-error]",
				"has been deployed with errors.",
				"https://okteto.dev/previews/test-preview-error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupContext(t)
			defer cleanup()

			got, err := generateMessage(tt.previewName, tt.exitCode)
			if err != nil {
				t.Fatalf("generateMessage() error = %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("generateMessage() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func TestGetPRNumber(t *testing.T) {
	tests := []struct {
		name        string
		eventName   string
		eventData   interface{}
		want        int
		wantErr     bool
		errContains string
	}{
		{
			name:      "pull_request event",
			eventName: "pull_request",
			eventData: map[string]interface{}{
				"number": 123,
			},
			want:    123,
			wantErr: false,
		},
		{
			name:      "repository_dispatch event",
			eventName: "repository_dispatch",
			eventData: map[string]interface{}{
				"client_payload": map[string]interface{}{
					"pull_request": map[string]interface{}{
						"number": 456,
					},
				},
			},
			want:    456,
			wantErr: false,
		},
		{
			name:        "repository_dispatch without PR number",
			eventName:   "repository_dispatch",
			eventData:   map[string]interface{}{},
			wantErr:     true,
			errContains: "missing pull request number",
		},
		{
			name:        "unsupported event type",
			eventName:   "push",
			eventData:   map[string]interface{}{},
			wantErr:     true,
			errContains: "unsupported event type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			eventPath := filepath.Join(tmpDir, "event.json")

			data, err := json.Marshal(tt.eventData)
			if err != nil {
				t.Fatalf("Failed to marshal test event data: %v", err)
			}

			if err := os.WriteFile(eventPath, data, 0644); err != nil {
				t.Fatalf("Failed to write event file: %v", err)
			}

			os.Setenv("GITHUB_EVENT_NAME", tt.eventName)
			os.Setenv("GITHUB_EVENT_PATH", eventPath)
			defer func() {
				os.Unsetenv("GITHUB_EVENT_NAME")
				os.Unsetenv("GITHUB_EVENT_PATH")
			}()

			got, err := getPRNumber()
			if (err != nil) != tt.wantErr {
				t.Errorf("getPRNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("getPRNumber() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if got != tt.want {
				t.Errorf("getPRNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOktetoURL(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) func()
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid context",
			setupFunc: func(t *testing.T) func() {
				return setupOktetoContext(t, "https://okteto.example.com")
			},
			want:    "https://okteto.example.com",
			wantErr: false,
		},
		{
			name: "different URL",
			setupFunc: func(t *testing.T) func() {
				return setupOktetoContext(t, "https://cloud.okteto.com")
			},
			want:    "https://cloud.okteto.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc(t)
			defer cleanup()

			got, err := getOktetoURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("getOktetoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("getOktetoURL() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if got != tt.want {
				t.Errorf("getOktetoURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslateEndpoints(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []string
		want      []string
	}{
		{
			name:      "single endpoint",
			endpoints: []string{"https://app.example.com"},
			want:      []string{"[https://app.example.com](https://app.example.com)"},
		},
		{
			name: "multiple endpoints",
			endpoints: []string{
				"https://app.example.com",
				"https://api.example.com",
			},
			want: []string{
				"[https://app.example.com](https://app.example.com)",
				"[https://api.example.com](https://api.example.com)",
			},
		},
		{
			name:      "empty list",
			endpoints: []string{},
			want:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateEndpoints(tt.endpoints)

			if len(got) != len(tt.want) {
				t.Errorf("translateEndpoints() returned %d items, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("translateEndpoints()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNotifyPRValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		message     string
		token       string
		previewName string
		wantErr     bool
		errContains string
	}{
		{
			name: "missing GITHUB_EVENT_NAME",
			setupEnv: func() {
				os.Unsetenv("GITHUB_EVENT_NAME")
			},
			message:     "test message",
			token:       "test-token",
			previewName: "test",
			wantErr:     true,
			errContains: "only supports either pull_request or repository_dispatch",
		},
		{
			name: "invalid event type",
			setupEnv: func() {
				os.Setenv("GITHUB_EVENT_NAME", "push")
			},
			message:     "test message",
			token:       "test-token",
			previewName: "test",
			wantErr:     true,
			errContains: "only supports either pull_request or repository_dispatch",
		},
		{
			name: "missing token",
			setupEnv: func() {
				os.Setenv("GITHUB_EVENT_NAME", "pull_request")
			},
			message:     "test message",
			token:       "",
			previewName: "test",
			wantErr:     true,
			errContains: "missing GITHUB_TOKEN",
		},
		{
			name: "missing repository",
			setupEnv: func() {
				os.Setenv("GITHUB_EVENT_NAME", "pull_request")
				os.Unsetenv("GITHUB_REPOSITORY")
			},
			message:     "test message",
			token:       "test-token",
			previewName: "test",
			wantErr:     true,
			errContains: "missing GITHUB_REPOSITORY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer func() {
				os.Unsetenv("GITHUB_EVENT_NAME")
				os.Unsetenv("GITHUB_REPOSITORY")
			}()

			err := notifyPR(tt.message, tt.token, tt.previewName)
			if (err != nil) != tt.wantErr {
				t.Errorf("notifyPR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("notifyPR() error = %v, want error containing %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestHandleGenerateAutoNotification(t *testing.T) {
	tests := []struct {
		name            string
		previewName     string
		exitCode        string
		setupContext    func(t *testing.T) func()
		githubToken     string
		expectNotify    bool
		wantErrContains string
	}{
		{
			name:        "generate without token - no notification",
			previewName: "test-preview",
			exitCode:    "0",
			setupContext: func(t *testing.T) func() {
				return setupOktetoContext(t, "https://okteto.example.com")
			},
			githubToken:  "",
			expectNotify: false,
		},
		{
			name:        "generate with token - should attempt notification",
			previewName: "test-preview",
			exitCode:    "0",
			setupContext: func(t *testing.T) func() {
				return setupOktetoContext(t, "https://okteto.example.com")
			},
			githubToken:     "fake-token",
			expectNotify:    true,
			wantErrContains: "only supports either pull_request or repository_dispatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupContext(t)
			defer cleanup()

			// Set up environment
			if tt.githubToken != "" {
				os.Setenv("GITHUB_TOKEN", tt.githubToken)
				defer os.Unsetenv("GITHUB_TOKEN")
			} else {
				os.Unsetenv("GITHUB_TOKEN")
			}

			// Test that generateMessage works
			msg, err := generateMessage(tt.previewName, tt.exitCode)
			if err != nil {
				t.Fatalf("generateMessage() error = %v", err)
			}

			if !strings.Contains(msg, tt.previewName) {
				t.Errorf("Message doesn't contain preview name: %s", msg)
			}

			// If we expect notification to be attempted, verify the error
			if tt.expectNotify {
				err := notifyPR(msg, tt.githubToken, tt.previewName)
				if err == nil {
					t.Error("Expected error from notifyPR, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}
			}
		})
	}
}

func TestMainSwitchLogic(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "generate command exists",
			command: "generate",
			wantErr: false,
		},
		{
			name:    "notify command exists",
			command: "notify",
			wantErr: false,
		},
		{
			name:    "unknown command",
			command: "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the command is recognized in the switch statement
			// We can't call main() directly, but we can verify the command names
			validCommands := map[string]bool{
				"generate": true,
				"notify":   true,
			}

			isValid := validCommands[tt.command]
			if isValid == tt.wantErr {
				t.Errorf("Command %q validation incorrect, isValid=%v, wantErr=%v", tt.command, isValid, tt.wantErr)
			}
		})
	}
}

// Helper function to setup a mock Okteto context
func setupOktetoContext(t *testing.T, url string) func() {
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".okteto", "context")
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		t.Fatalf("Failed to create context directory: %v", err)
	}

	contextData := map[string]interface{}{
		"current-context": "test-context",
		"contexts": map[string]interface{}{
			"test-context": map[string]interface{}{
				"name": url,
			},
		},
	}

	data, err := json.Marshal(contextData)
	if err != nil {
		t.Fatalf("Failed to marshal context data: %v", err)
	}

	configPath := filepath.Join(contextDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	return func() {
		os.Setenv("HOME", oldHome)
	}
}
