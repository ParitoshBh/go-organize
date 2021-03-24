package handlers

import (
	"go-organizer/backend/connections"
	"go-organizer/backend/models"
	"go-organizer/backend/utils"
	"net/http"
)

func Update(w http.ResponseWriter, r *http.Request) {
	sessionManager := utils.GetSessionManager()
	goOrmDB := connections.GetGoOrmDBConnection()
	userConfig := models.UserConfig{}

	if !sessionManager.Exists(r.Context(), "userId") {
		sessionManager.Put(r.Context(), "FlashMessage", "Unable to retrieve User ID from session")
		// TODO Redirect to error page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	err := r.ParseForm()
	if err != nil {
		sessionManager.Put(r.Context(), "FlashMessage", err.Error())
		// TODO Redirect to error page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	result := goOrmDB.Model(&userConfig).Where("user_id = ?", sessionManager.Get(r.Context(), "userId")).Update("layout_style", r.FormValue("layoutStyle"))
	if result.Error != nil {
		sessionManager.Put(r.Context(), "FlashMessage", result.Error.Error())
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
