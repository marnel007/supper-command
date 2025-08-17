package security_test

import (
	"testing"

	"suppercommand/internal/security"
)

func TestValidator_ValidateCommand(t *testing.T) {
	validator := security.NewValidator()

	tests := []struct {
		name    string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid command",
			cmd:     "ls",
			args:    []string{"-l", "home"},
			wantErr: false,
		},
		{
			name:    "empty command",
			cmd:     "",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "command with dangerous characters",
			cmd:     "ls; rm -rf /",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "path traversal attempt",
			cmd:     "cat",
			args:    []string{"../../../etc/passwd"},
			wantErr: true,
		},
		{
			name:    "command with null byte",
			cmd:     "ls\x00",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCommand(tt.cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_SanitizeInput(t *testing.T) {
	validator := security.NewValidator()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "clean input",
			input:    "ls -l",
			expected: "ls -l",
			wantErr:  false,
		},
		{
			name:     "input with null bytes",
			input:    "ls\x00 -l",
			expected: "ls -l",
			wantErr:  false,
		},
		{
			name:     "input with line breaks",
			input:    "ls\r\n -l",
			expected: "ls -l",
			wantErr:  false,
		},
		{
			name:     "input with extra whitespace",
			input:    "  ls -l  ",
			expected: "ls -l",
			wantErr:  false,
		},
		{
			name:    "input too long",
			input:   string(make([]byte, 2000)), // Longer than max length
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.SanitizeInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("SanitizeInput() = %v, want %v", result, tt.expected)
			}
		})
	}
}
