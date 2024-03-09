package server

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

const (
	keyPathName               = "key"
	errBadRequestResponse     = "Bad Request"
	errNotFoundResponse       = "Key Not Found"
	errInternalServerResponse = "Internal Server Error"
)

type Cache interface {
	Set(key string, value string)
	Get(key string) (string, bool)
}

func New(logger *zerolog.Logger, cache Cache) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{key}", get(cache, logger))
	mux.HandleFunc("POST /{key}", store(cache, logger))
	var handler http.Handler = mux
	return handler
}

func get(cache Cache, logger *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue(keyPathName)
		if key == "" {
			http.Error(w, errBadRequestResponse, http.StatusBadRequest)
			return
		}
		logger.Debug().Str("key", key).Msg("Received GET key request")
		value, ok := cache.Get(key)
		if !ok {
			logger.Debug().Str("key", key).Msg("Cache miss.")
			http.Error(w, errNotFoundResponse, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(value))
		if err != nil {
			log.Error().Err(err).Msg("Failed to write response")
		}
		logger.Debug().Str("key", key).Msg("Cache hit.")
	}
}

func store(cache Cache, logger *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue(keyPathName)
		if key == "" {
			http.Error(w, errBadRequestResponse, http.StatusBadRequest)
			return
		}
		logger.Debug().Str("key", key).Msg("Received Post key request")
		value, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to read request body")
			http.Error(w, errBadRequestResponse, http.StatusBadRequest)
			return
		}
		valueStr := string(value)
		if valueStr == "" {
			http.Error(w, errBadRequestResponse, http.StatusBadRequest)
			return
		}
		cache.Set(key, valueStr)
		w.WriteHeader(http.StatusCreated)
	}
}
