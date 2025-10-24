package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() *chi.Mux {

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Testing chi router"))
	})

	return r
}
