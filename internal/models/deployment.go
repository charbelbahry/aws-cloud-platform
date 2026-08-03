package models

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrServiceIDRequired = errors.New("service_id is required")
	ErrVersionRequired   = errors.New("version is required")
	ErrInvalidStatus     = errors.New("invalid deployment status")
)

const (
	StatusPending    = "pending"
	StatusBuilding   = "building"
	StatusRunning    = "running"
	StatusFailed     = "failed"
	StatusRolledBack = "rolled_back"
)

type Deployment struct {
	ID        string    `json:"id"`
	ServiceID string    `json:"service_id"`
	Version   string    `json:"version"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (d *Deployment) Validate() error {
	d.ServiceID = strings.TrimSpace(d.ServiceID)
	if d.ServiceID == "" {
		return ErrServiceIDRequired
	}

	d.Version = strings.TrimSpace(d.Version)
	if d.Version == "" {
		return ErrVersionRequired
	}

	d.Status = strings.TrimSpace(d.Status)
	if d.Status == "" {
		d.Status = StatusPending
	}

	switch d.Status {
	case StatusPending, StatusBuilding, StatusFailed, StatusRunning, StatusRolledBack:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidStatus, d.Status)
	}
}
