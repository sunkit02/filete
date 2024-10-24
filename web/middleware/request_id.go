package middleware

import (
	"github.com/sunkit02/filete/logging"
	"github.com/sunkit02/filete/web/types"
	"net/http"

	"github.com/google/uuid"
)

func RequestIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add(types.RequestIdHeaderName, uuid.New().String())
		next.ServeHTTP(w, r)
	})
}

// This function extracts the request id from an HTTP request and if there is
// none present, it sets it and returns it
func ExtractRequestId(r *http.Request) uuid.UUID {
	idString := r.Header.Get(types.RequestIdHeaderName)
	if idString == "" {
		idString = uuid.New().String()
		r.Header.Set(types.RequestIdHeaderName, idString)
		logging.Warning.Printf("Request with id %v didn't initially have one generated.", idString)
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		panic("Failed to parse newly created uuid. This should never happen.")
	}

	return id
}
