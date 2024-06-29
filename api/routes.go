package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type pathPartsInterface struct{}

func (api *API) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Post("/signup", api.shamirRequired(api.signUp))
		r.Post("/signin", api.shamirRequired(api.signIn))

		r.Route(`/secrets`, func(r chi.Router) {
			r.Use(keyCtx)
			r.Get("/*", api.AuthRequired(api.getByKey))
			r.Delete("/*", api.AuthRequired(api.deleteByKey))
			r.Post("/*", api.AuthRequired(api.setByKey))
		})
		r.Get("/list/*", api.AuthRequired(api.listKV))

		r.Post("/unseal", api.unseal)
		r.Get("/master", api.master)

		r.Get("/test", api.test)
		r.Post("/show", api.showRootKey)
	})

	return r
}

func keyCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, _ := strings.CutPrefix(r.URL.Path, "/api/secrets/")
		pathParts := strings.Split(path, "/")

		fmt.Println(pathParts)

		ctx := context.WithValue(r.Context(), pathPartsInterface{}, pathParts)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
