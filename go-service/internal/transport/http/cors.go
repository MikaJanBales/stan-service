package http

import (
	"net/http"

	"github.com/rs/cors"
)

func CorsSettings() *cors.Cors {
	c := cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodDelete,
			http.MethodPost,
			http.MethodPut,
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Accept",
			"Accept-Language",
		},
		AllowCredentials:   true,
		OptionsPassthrough: true,
		ExposedHeaders: []string{
			"Content-Type",
			"Accept",
			"Accept-Language",
		},
		Debug: true,
	})
	return c
}
