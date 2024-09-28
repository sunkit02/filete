package web

import (
	"net/http"
	"strings"
)

func SessionKeyAuthMiddleware(key string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("Authorization")
		if authToken == "" {
			http.Error(w, "Missing bearer token", http.StatusUnauthorized)
			return
		}

		requestKey, err := parseBearerToken(authToken)

		if err != nil {
			// TODO: Get requestId in here using a middleware
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if requestKey != key {
			http.Error(w, "Invalid bearer token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseBearerToken(token string) (string, error) {
	if !strings.HasPrefix(token, "Bearer ") {
		return "", BadTokenError{message: "Bad bearer token"}
	}

	return strings.SplitN(token, " ", 2)[1], nil
}

type BadTokenError struct {
	message string
}

func (e BadTokenError) Error() string {
	return e.message
}
