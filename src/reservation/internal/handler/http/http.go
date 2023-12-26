package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"go.example/saga/reservation/internal/controller/reservation"
	"go.example/saga/reservation/internal/repository"
	"go.example/saga/reservation/pkg/model"
	"net/http"
	"net/url"
)

type Response struct {
	Status     string                 `json:"status"`
	StatusCode int                    `json:"status_code"`
	Error      string                 `json:"error"`
	Data       map[string]interface{} `json:"data"`
}

// Handler defines a HTTP rating handler.
type Handler struct {
	ctrl *reservation.Controller
}

// New creates a new reservation service Router with HTTP handler.
func New(ctrl *reservation.Controller) *httprouter.Router {

	h := Handler{ctrl}

	router := httprouter.New()
	router.POST("/api/v1/reservations", h.Create)
	router.GET("/api/v1/reservations/:id", h.Read)

	return router
}

// Create POST new reservation
func (h *Handler) Create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	var cmd model.ReservationCmd
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Malformed JSON", http.StatusBadRequest)
		return
	}

	v, err := h.ctrl.PostReservation(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	locationURL := &url.URL{
		Scheme: "http", // FIXME
		Host:   r.Host,
		Path:   fmt.Sprintf("%s/%s", r.URL.Path, v.ID),
	}

	w.Header().Set("Location", locationURL.String())
	w.Header().Set("Retry-After", "0.5") // sec
	w.WriteHeader(http.StatusAccepted)
}

// Read
func (h *Handler) Read(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	re, err := h.ctrl.GetReservation(r.Context(), ps.ByName("id"))

	w.Header().Set("Content-Type", "application/json")
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "Reservation not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(re); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
}
