package middlewares

import (
	"go-organizer/backend/utils"
	"net/http"
)

// Session loads a session
func Session(next http.Handler) http.Handler {
	sessionManager := utils.GetSessionManager()

	return sessionManager.LoadAndSave(next)
}
