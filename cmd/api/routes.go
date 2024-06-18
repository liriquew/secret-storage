package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *App) routes() http.Handler {
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Get("/get", app.getByKey)
		r.Post("/set", app.setValueByKey)
		r.Delete("/delete", app.deleteByKey)
		r.Get("/test", app.test)
	})

	return r
}
