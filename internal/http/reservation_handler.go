package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"

	"flash-sale-reservation/internal/reservation"
)

type ReservationHandler struct {
	service *reservation.Service
}

func NewReservationHandler(service *reservation.Service) *ReservationHandler {
	return &ReservationHandler{service: service}
}

// POST /reservations
func (h *ReservationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID int64 `json:"product_id"`
		UserID    int64 `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.service.Create(r.Context(), req.ProductID, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// GET /reservations/{id}
func (h *ReservationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	res, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(res)
}

// POST /reservations/{id}/confirm
func (h *ReservationHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err := h.service.Confirm(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// POST /reservations/{id}/cancel
func (h *ReservationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err := h.service.Cancel(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /reservations?user_id=&status=&limit=&offset=
func (h *ReservationHandler) List(w http.ResponseWriter, r *http.Request) {
	var (
		userID *int64
		status *string
		limit  = 20
		offset = 0
	)

	if v := r.URL.Query().Get("user_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		userID = &id
	}

	if v := r.URL.Query().Get("status"); v != "" {
		status = &v
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		limit, _ = strconv.Atoi(v)
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		offset, _ = strconv.Atoi(v)
	}

	res, err := h.service.List(r.Context(), userID, status, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
