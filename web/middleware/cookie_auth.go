package middleware

import (
	"net/http"

	"github.com/sunkit02/filete/logging"
	"github.com/sunkit02/filete/services"
	"github.com/sunkit02/filete/web/types"
	"github.com/sunkit02/filete/web/utils"
)

func CookieAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := ExtractRequestId(r)
		logging.Info.Println(utils.WithId(id, "CookieAuthMiddleware"))

		authCookie, err := r.Cookie(types.SessionIdCookieName)
		if err != nil {
			logging.Debug.Println(utils.WithId(id, "CookieAuthMiddleware: No auth cookie found"))
			http.Error(w, utils.WithId(id, "No auth cookie found"), http.StatusUnauthorized)
			return
		}

		if !services.ValidateSessionCookie(*authCookie) {
			logging.Debug.Println(utils.WithId(id, "CookieAuthMiddleware: invalid cookie"))
			http.Error(w, utils.WithId(id, "Invalid auth cookie"), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
