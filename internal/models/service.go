// Package models defines data structures and domain validation logic
package models

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrServiceNameRequired = errors.New("service name is required")
	ErrServiceNameTooLong  = errors.New("service name must be 100 characters or less")
)

type Service struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Repository  string    `json:"repository,omitempty"`
	Description string    `json:"description.omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Service) Validate() error {
	s.Name = strings.TrimSpace(s.Name)
	if s.Name == "" {
		return ErrServiceNameRequired
	}
	if len(s.Name) > 100 {
		return ErrServiceNameTooLong
	}
	return nil
}
