package middleware

import (
	"github.com/sunkit02/filete/logging"
	"net/http"
	"strings"
)

func SessionKeyAuthMiddleware(key string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqId := ExtractRequestId(r)
		authToken := r.Header.Get("Authorization")
		if authToken == "" {
			http.Error(w, "Missing bearer token", http.StatusUnauthorized)
			logging.Info.Printf("reqId(%s) Missing bearer token", reqId)
			return
		}

		requestKey, err := parseBearerToken(authToken)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logging.Info.Printf("reqId(%s) %s", reqId, err.Error())
			return
		}

		if requestKey != key {
			http.Error(w, "Invalid bearer token", http.StatusForbidden)
			logging.Info.Printf("reqId(%s) Invalid bearer token", reqId)
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
