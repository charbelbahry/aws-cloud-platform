package models

import (
	"errors"
	"strings"
	"testing"
)

func TestServiceValidate(t *testing.T) {
	tests := []struct {
		name    string
		service Service
		wantErr error
	}{
		{
			name:    "valid service",
			service: Service{Name: "web-api"},
			wantErr: nil,
		},
		{
			name:    "empty name",
			service: Service{Name: "  "},
			wantErr: ErrServiceNameRequired,
		},
		{
			name:    "name too long",
			service: Service{Name: strings.Repeat("a", 101)},
			wantErr: ErrServiceNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
