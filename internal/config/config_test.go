package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    Config
	}{
		{
			name:    "defaults",
			envVars: nil,
			want: Config{
				SigrokCLIPath: "sigrok-cli",
				Timeout:       30 * time.Second,
				WorkingDir:    "",
			},
		},
		{
			name: "custom cli path",
			envVars: map[string]string{
				"SIGROK_CLI_PATH": "/usr/local/bin/sigrok-cli",
			},
			want: Config{
				SigrokCLIPath: "/usr/local/bin/sigrok-cli",
				Timeout:       30 * time.Second,
				WorkingDir:    "",
			},
		},
		{
			name: "custom timeout",
			envVars: map[string]string{
				"SIGROK_TIMEOUT_SECONDS": "60",
			},
			want: Config{
				SigrokCLIPath: "sigrok-cli",
				Timeout:       60 * time.Second,
				WorkingDir:    "",
			},
		},
		{
			name: "custom working dir",
			envVars: map[string]string{
				"SIGROK_WORKING_DIR": "/tmp/sigrok",
			},
			want: Config{
				SigrokCLIPath: "sigrok-cli",
				Timeout:       30 * time.Second,
				WorkingDir:    "/tmp/sigrok",
			},
		},
		{
			name: "all custom",
			envVars: map[string]string{
				"SIGROK_CLI_PATH":        "/opt/sigrok/bin/sigrok-cli",
				"SIGROK_TIMEOUT_SECONDS": "120",
				"SIGROK_WORKING_DIR":     "/data",
			},
			want: Config{
				SigrokCLIPath: "/opt/sigrok/bin/sigrok-cli",
				Timeout:       120 * time.Second,
				WorkingDir:    "/data",
			},
		},
		{
			name: "invalid timeout falls back to default",
			envVars: map[string]string{
				"SIGROK_TIMEOUT_SECONDS": "not-a-number",
			},
			want: Config{
				SigrokCLIPath: "sigrok-cli",
				Timeout:       30 * time.Second,
				WorkingDir:    "",
			},
		},
		{
			name: "negative timeout falls back to default",
			envVars: map[string]string{
				"SIGROK_TIMEOUT_SECONDS": "-5",
			},
			want: Config{
				SigrokCLIPath: "sigrok-cli",
				Timeout:       30 * time.Second,
				WorkingDir:    "",
			},
		},
		{
			name: "zero timeout falls back to default",
			envVars: map[string]string{
				"SIGROK_TIMEOUT_SECONDS": "0",
			},
			want: Config{
				SigrokCLIPath: "sigrok-cli",
				Timeout:       30 * time.Second,
				WorkingDir:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			got := Load()

			if got.SigrokCLIPath != tt.want.SigrokCLIPath {
				t.Errorf("SigrokCLIPath = %q, want %q", got.SigrokCLIPath, tt.want.SigrokCLIPath)
			}
			if got.Timeout != tt.want.Timeout {
				t.Errorf("Timeout = %v, want %v", got.Timeout, tt.want.Timeout)
			}
			if got.WorkingDir != tt.want.WorkingDir {
				t.Errorf("WorkingDir = %q, want %q", got.WorkingDir, tt.want.WorkingDir)
			}
		})
	}
}
