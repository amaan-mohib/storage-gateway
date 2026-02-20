package http

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/storage-gateway/src/config"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Access-Token")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		decodedToken, decodeErr := base64.StdEncoding.DecodeString(token)
		if decodeErr != nil {
			http.Error(w, "Invalid token", http.StatusBadRequest)
			return
		}
		if string(decodedToken) != config.GetSafeEnv(config.AdminAccessToken) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "token", string(decodedToken))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
