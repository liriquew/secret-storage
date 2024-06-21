package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (api *API) routes() http.Handler {
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Post("/signup", api.signUp)
		r.Post("/signin", api.signIn)

		r.Get("/get", api.AuthRequired(api.getByKey))
		r.Post("/set", api.AuthRequired(api.setByKey))
		r.Delete("/delete", api.AuthRequired(api.deleteByKey))
		r.Get("/test", api.test)
		r.Post("/show", api.showRootKey)
	})

	return r
}
