package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *API) routes() http.Handler {
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Post("/signup", app.signUp)
		r.Post("/signin", app.signIn)

		r.Get("/get", app.AuthRequired(app.getByKey))
		r.Post("/set", app.AuthRequired(app.setValueByKey))
		r.Delete("/delete", app.AuthRequired(app.deleteByKey))
		r.Get("/test", app.test)
	})

	return r
}
