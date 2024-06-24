package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *API) routes() http.Handler {
	r := chi.NewRouter()

	r.Route("/api/kv", func(r chi.Router) {
		r.Post("/signup", api.signUp)
		r.Post("/signin", api.signIn)

		r.Route("/{key}", func(r chi.Router) {
			r.Use(keyCtx)
			r.Get("/", api.AuthRequired(api.getByKey))
			r.Delete("/", api.AuthRequired(api.deleteByKey))

		})

		r.Post("/", api.AuthRequired(api.setByKey))
		r.Get("/list", api.AuthRequired(api.listKV))

		r.Get("/test", api.test)
		r.Post("/show", api.showRootKey)
	})

	return r
}

func keyCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		ctx := context.WithValue(r.Context(), "key", key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
