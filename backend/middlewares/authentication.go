package middlewares

import (
	"go-organizer/backend/logger"
	"go-organizer/backend/utils"
	"net/http"

	"github.com/gorilla/mux"
)

// Authentication authenticates session
func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_logger := logger.Logger

		sessionManager := utils.GetSessionManager()

		// skip checking for valid of assets
		currentRoute := mux.CurrentRoute(r)
		if currentRoute.GetName() != "assets" {
			if currentRoute.GetName() == "login" {
				// if already autherized, redirect to homepage
				if sessionManager.GetBool(r.Context(), "isAuthorized") {
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			} else {
				if !sessionManager.GetBool(r.Context(), "isAuthorized") {
					_logger.Warn("Unauthorized request, redirecting to login page")
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
			}
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
