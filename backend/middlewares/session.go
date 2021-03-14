package middlewares

import (
	"go-organizer/backend/utils"
	"net/http"
)

// Session loads a session
func Session(next http.Handler) http.Handler {
	sessionManager := utils.GetSessionManager()

	// TODO
	// add error handler if session table isn't available
	// default behaviour is to throw internal server error

	return sessionManager.LoadAndSave(next)
}
