package app_test

import (
	"context"
	"testing"
	"time"

	"suppercommand/internal/app"
)

func TestApplication_Initialize(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful initialization",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := app.NewApplication()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := application.Initialize(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Application.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Clean shutdown
			if err == nil {
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()
				application.Shutdown(shutdownCtx)
			}
		})
	}
}

func TestApplication_ExecuteCommand(t *testing.T) {
	application := app.NewApplication()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initialize application
	if err := application.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		application.Shutdown(shutdownCtx)
	}()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty command",
			input:   "",
			wantErr: false,
		},
		{
			name:    "exit command",
			input:   "exit",
			wantErr: false,
		},
		{
			name:    "unknown command",
			input:   "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := application.ExecuteCommand(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Application.ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Errorf("Application.ExecuteCommand() returned nil result")
			}
		})
	}
}
