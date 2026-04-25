package subprocess

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestExecCommandResponseStdoutContents(t *testing.T) {
	tests := []struct {
		name   string
		stdout string
		want   string
	}{
		{"trims whitespace", "  hello world  \n", "hello world"},
		{"empty string", "", ""},
		{"only whitespace", "   \n\t  ", ""},
		{"no trimming needed", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ExecCommandResponse{Stdout: tt.stdout}
			if got := resp.StdoutContents(); got != tt.want {
				t.Errorf("StdoutContents() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecCommandResponseStderrContents(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   string
	}{
		{"trims whitespace", "  error message  \n", "error message"},
		{"empty string", "", ""},
		{"only whitespace", "   \n\t  ", ""},
		{"no trimming needed", "error", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ExecCommandResponse{Stderr: tt.stderr}
			if got := resp.StderrContents(); got != tt.want {
				t.Errorf("StderrContents() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecCommandResponseStdoutBytes(t *testing.T) {
	resp := ExecCommandResponse{Stdout: "  hello world  \n"}
	got := resp.StdoutBytes()
	want := []byte("hello world")
	if !bytes.Equal(got, want) {
		t.Errorf("StdoutBytes() = %v, want %v", got, want)
	}

	empty := ExecCommandResponse{Stdout: ""}
	if got := empty.StdoutBytes(); len(got) != 0 {
		t.Errorf("StdoutBytes() for empty = %v, want empty", got)
	}
}

func TestExecCommandResponseStderrBytes(t *testing.T) {
	resp := ExecCommandResponse{Stderr: "  error msg  \n"}
	got := resp.StderrBytes()
	want := []byte("error msg")
	if !bytes.Equal(got, want) {
		t.Errorf("StderrBytes() = %v, want %v", got, want)
	}

	empty := ExecCommandResponse{Stderr: ""}
	if got := empty.StderrBytes(); len(got) != 0 {
		t.Errorf("StderrBytes() for empty = %v, want empty", got)
	}
}

func TestCallExecCommandSuccess(t *testing.T) {
	resp, err := CallExecCommand(ExecCommandInput{
		Command: "echo",
		Args:    []string{"hello"},
	})
	if err != nil {
		t.Fatalf("CallExecCommand failed: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", resp.ExitCode)
	}
	if !strings.Contains(resp.StdoutContents(), "hello") {
		t.Errorf("stdout = %q, want it to contain 'hello'", resp.StdoutContents())
	}
}

func TestCallExecCommandFailure(t *testing.T) {
	resp, err := CallExecCommand(ExecCommandInput{
		Command: "false",
	})
	if err == nil {
		t.Fatal("expected error for failing command")
	}
	if resp.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
}

func TestCallExecCommandNotFound(t *testing.T) {
	_, err := CallExecCommand(ExecCommandInput{
		Command: "nonexistent-binary-docket-test-12345",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}

func TestCallExecCommandWithEnv(t *testing.T) {
	resp, err := CallExecCommand(ExecCommandInput{
		Command: "env",
		Env:     map[string]string{"DOCKET_TEST_VAR": "test123"},
	})
	if err != nil {
		t.Fatalf("CallExecCommand failed: %v", err)
	}
	if !strings.Contains(resp.StdoutContents(), "DOCKET_TEST_VAR=test123") {
		t.Errorf("stdout = %q, want it to contain 'DOCKET_TEST_VAR=test123'", resp.StdoutContents())
	}
}

func TestCallExecCommandWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := CallExecCommandWithContext(ctx, ExecCommandInput{
		Command: "sleep",
		Args:    []string{"10"},
	})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
