package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/charbelbahry/aws-cloud-platform/internal/database"
	"github.com/charbelbahry/aws-cloud-platform/internal/models"
)

type DeploymentHandler struct {
	db *database.DB
}

func NewDeploymentHandler(db *database.DB) *DeploymentHandler {
	return &DeploymentHandler{db: db}
}

func (h *DeploymentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /services/{id}/deployments", h.Create)
	mux.HandleFunc("GET /services/{id}/deployments", h.List)
}

func (h *DeploymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")
	if serviceID == "" {
		respondError(w, http.StatusBadRequest, "service ID is required")
		return
	}

	var req models.Deployment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	req.ServiceID = serviceID
	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM services WHERE id = $1);`
	if err := h.db.QueryRowContext(ctx, checkQuery, serviceID).Scan(&exists); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify service existence")
		return
	}

	if !exists {
		respondError(w, http.StatusNotFound, "service not found")
		return
	}

	insertQuery := `
		INSERT INTO deployments (service_id, version, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`

	err := h.db.QueryRowContext(ctx, insertQuery, req.ServiceID, req.Version, req.Status).Scan(
		&req.ID,
		&req.CreatedAt,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create deployment")
		return
	}

	respondJSON(w, http.StatusCreated, req)
}

func (h *DeploymentHandler) List(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")
	if serviceID == "" {
		respondError(w, http.StatusBadRequest, "service ID is required")
		return
	}

	ctx := r.Context()

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM services WHERE id = $1);`
	if err := h.db.QueryRowContext(ctx, checkQuery, serviceID).Scan(&exists); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify service existence")
		return
	}

	if !exists {
		respondError(w, http.StatusNotFound, "service not found")
		return
	}

	query := `
			SELECT id, service_id, version, status, created_at
			FROM deployments
			WHERE service_id = $1
			ORDER BY created_at DESC;
	`

	rows, err := h.db.QueryContext(ctx, query, serviceID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed deployments")
		return
	}
	defer rows.Close()

	deployments := make([]models.Deployment, 0)
	for rows.Next() {
		var d models.Deployment
		if err := rows.Scan(&d.ID, &d.ServiceID, &d.Version, &d.Status, &d.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan deployment")
			return
		}
		deployments = append(deployments, d)
	}

	if err := rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "error reading deployment rows")
		return
	}

	respondJSON(w, http.StatusOK, deployments)
}
