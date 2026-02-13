package http

import (
	"net/http"

	"flash-sale-reservation/internal/product"
	"flash-sale-reservation/internal/reservation"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	productService *product.Service,
	reservationService *reservation.Service,
) http.Handler {

	r := chi.NewRouter()

	// ---------- Health ----------
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// ---------- Products ----------
	productHandler := NewProductHandler(productService)
	r.Route("/products", func(r chi.Router) {
		r.Post("/", productHandler.Create)
		r.Get("/", productHandler.List)
	})

	// ---------- Reservations ----------
	reservationHandler := NewReservationHandler(reservationService)
	r.Route("/reservations", func(r chi.Router) {
		r.Post("/", reservationHandler.Create)     // создать резерв
		r.Get("/{id}", reservationHandler.GetByID) // получить по id
		r.Post("/{id}/confirm", reservationHandler.Confirm)
		r.Post("/{id}/cancel", reservationHandler.Cancel)
		r.Get("/", reservationHandler.List) // фильтры + пагинация
	})

	// ---------- Admin ----------
	r.Route("/admin", func(r chi.Router) {
		r.Route("/reservations", func(r chi.Router) {
			r.Post("/sync-expired", reservationHandler.SyncExpired)
		})
	})

	return r
}
