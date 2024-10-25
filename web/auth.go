package web

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sunkit02/filete/logging"
	"github.com/sunkit02/filete/services"
	"github.com/sunkit02/filete/web/middleware"
	"github.com/sunkit02/filete/web/types"
	"github.com/sunkit02/filete/web/utils"
)

func AuthRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /authenticate", handleAuthenticate)
	mux.HandleFunc("POST /invalidate-token", handleInvalidateToken)

	return mux
}

type authRequest struct {
	sessionKey string `json:"sessionKey"`
}

func handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	id := middleware.ExtractRequestId(r)
	logging.Info.Println(utils.WithId(id, "handleAuthenticate"))

	body, err := io.ReadAll(r.Body)
	if err != nil {
		msg := utils.WithId(id, "Failed to read request body")
		logging.Debug.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	authReq := &authRequest{}
	err = json.Unmarshal(body, &authReq)
	if err != nil {
		logging.Debug.Println(utils.WithId(id, "Failed to decode request body: %v", err))
		http.Error(w, utils.WithId(id, "Invalid request body"), http.StatusBadRequest)
		return
	}

	logging.Trace.Println("authRequest", authReq)

	session, err := services.AuthenticateWithSessionKey("123")
	if err != nil {
		msg := utils.WithId(id, "Invalid sessionId")
		logging.Debug.Println(msg, http.StatusUnauthorized)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     types.SessionIdCookieName,
		Value:    session.Id,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   3600, // one hour
	})

	w.WriteHeader(http.StatusNoContent)
}

func handleInvalidateToken(w http.ResponseWriter, r *http.Request) {
	id := middleware.ExtractRequestId(r)
	logging.Info.Println(utils.WithId(id, "handleInvalidateToken"))

	sessionCookie, err := r.Cookie(types.SessionIdCookieName)
	// NOTE: This should not happen, this means the cookie has been tampered with or we
	// issued an empty cookie
	if err != nil {
		msg := utils.WithId(id, "No session cookie found")
		logging.Debug.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	services.InvalidateSession(sessionCookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:     types.SessionIdCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1, // expire the cookie
	})

	w.WriteHeader(http.StatusNoContent)
}
