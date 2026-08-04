// Package handlers implements HTTP request handlers for the API
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/charbelbahry/aws-cloud-platform/internal/database"
	"github.com/charbelbahry/aws-cloud-platform/internal/models"
)

type ServiceHandler struct {
	db *database.DB
}

func NewServiceHandler(db *database.DB) *ServiceHandler {
	return &ServiceHandler{db: db}
}

func (h *ServiceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /services", h.Create)
	mux.HandleFunc("GET /services", h.List)
	mux.HandleFunc("GET /services/{id}", h.GetByID)
	mux.HandleFunc("PUT /services/{id}", h.Update)
	mux.HandleFunc("DELETE /services/{id}", h.Delete)
}

func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.Service
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	query := `
				INSERT INTO services (name, repository, description)
				VALUES ($1, $2, $3)
				RETURNING id, created_at, updated_at;
	`
	ctx := r.Context()
	err := h.db.QueryRowContext(ctx, query, req.Name, req.Repository, req.Description).Scan(
		&req.ID,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create service")
		return
	}

	respondJSON(w, http.StatusCreated, req)
}

func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	query := `
				SELECT id, name, COALESCE(repository, ''), COALESCE(description, ''), created_at, updated_at
				FROM services
				ORDER BY created_at DESC;
	`

	ctx := r.Context()
	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch services")
		return
	}
	defer rows.Close()

	services := make([]models.Service, 0)
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Repository, &s.Description, &s.CreatedAt, &s.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan services")
			return
		}
		services = append(services, s)
	}

	if err := rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "error reading service rows")
		return
	}

	respondJSON(w, http.StatusOK, services)
}

func (h *ServiceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "service ID is required")
		return
	}

	query := `
				SELECT id, name, COALESCE(repository, ''), COALESCE(description, ''), created_at, updated_at
				FROM services
				WHERE id = $1;
	`
	var s models.Service
	ctx := r.Context()
	err := h.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID,
		&s.Name,
		&s.Repository,
		&s.Description,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		respondError(w, http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch service")
		return
	}

	respondJSON(w, http.StatusOK, s)
}

func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "service ID is required")
		return
	}

	var req models.Service
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	query := `
				UPDATE services
				SET name = $1, repository = $2, description = $3, updated_at = now()
				WHERE ID = $4
				RETURNING id, created_at, updated_at;
	`

	ctx := r.Context()
	err := h.db.QueryRowContext(ctx, query, req.Name, req.Repository, req.Description, id).Scan(
		&req.ID,
		&req.CreatedAt,
		&req.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		respondError(w, http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update service")
		return
	}

	respondJSON(w, http.StatusOK, req)
}

func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "service ID is required")
		return
	}

	query := `DELETE FROM services WHERE id = $1`

	ctx := r.Context()
	res, err := h.db.ExecContext(ctx, query, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete service")
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to check deletion status")
		return
	}

	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "service not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
