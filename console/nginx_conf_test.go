package console

import (
	"os"
	"strings"
	"testing"
)

func TestGetNginxWorkerProcesses(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		setEnv   bool
		want     string
	}{
		{name: "default when unset", setEnv: false, want: DefaultNginxWorkerProcesses},
		{name: "default when empty", setEnv: true, envValue: "", want: DefaultNginxWorkerProcesses},
		{name: "positive integer", setEnv: true, envValue: "4", want: "4"},
		{name: "auto", setEnv: true, envValue: "auto", want: "auto"},
		{name: "AUTO case insensitive", setEnv: true, envValue: "AUTO", want: "auto"},
		{name: "invalid falls back", setEnv: true, envValue: "not-a-number", want: DefaultNginxWorkerProcesses},
		{name: "zero falls back", setEnv: true, envValue: "0", want: DefaultNginxWorkerProcesses},
		{name: "negative falls back", setEnv: true, envValue: "-1", want: DefaultNginxWorkerProcesses},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(NginxWorkerProcessesEnvVar, tt.envValue)
			} else {
				t.Setenv(NginxWorkerProcessesEnvVar, "")
				_ = os.Unsetenv(NginxWorkerProcessesEnvVar)
			}
			if got := GetNginxWorkerProcesses(); got != tt.want {
				t.Errorf("GetNginxWorkerProcesses() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateNginxConf_DefaultWorkerProcesses(t *testing.T) {
	t.Setenv(NginxWorkerProcessesEnvVar, "")
	_ = os.Unsetenv(NginxWorkerProcessesEnvVar)

	conf := GenerateNginxConf()
	if !strings.Contains(conf, "worker_processes 8;") {
		t.Errorf("expected default worker_processes 8, got conf without it")
	}
	if strings.Contains(conf, "worker_processes auto;") {
		t.Errorf("expected no worker_processes auto")
	}
}

func TestGenerateNginxConf_WorkerProcessesFromEnv(t *testing.T) {
	t.Setenv(NginxWorkerProcessesEnvVar, "2")
	conf := GenerateNginxConf()
	if !strings.Contains(conf, "worker_processes 2;") {
		t.Errorf("expected worker_processes 2 from env")
	}
	if strings.Contains(conf, "worker_processes auto;") {
		t.Errorf("expected no worker_processes auto")
	}
}
