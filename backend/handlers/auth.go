package handlers

import (
	"errors"
	"go-organizer/backend/connections"
	"go-organizer/backend/logger"
	"go-organizer/backend/models"
	"go-organizer/backend/templmanager"
	"go-organizer/backend/utils"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Login handles login request
func Login(w http.ResponseWriter, r *http.Request) {
	// ViewData variables for view vars
	type ViewData struct {
		Error string
	}

	viewData := ViewData{}
	_logger := logger.Logger

	// authenticate user, if post request
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			_logger.Fatalf(err.Error())
		}

		userID, err := validateUser(r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			viewData.Error = err.Error()
		} else {
			sessionManager := utils.GetSessionManager()
			sessionManager.Put(r.Context(), "isAuthorized", true)
			sessionManager.Put(r.Context(), "userId", userID)

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}

	templmanager.RenderTemplate(w, "login.html", viewData)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	sessionManager := utils.GetSessionManager()
	sessionManager.Destroy(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func validateUser(username string, password string) (uint, error) {
	user := models.User{}

	if username == "" || password == "" {
		return 0, errors.New("Username or password fields cannot be empty")
	}

	goOrmDB := connections.GetGoOrmDBConnection()
	result := goOrmDB.Where("username = ?", username).First(&user)

	// if no result from query -> invalid username
	if result.RowsAffected == 0 {
		return 0, errors.New("Username not found")
	}

	// if password hash doesn't match -> invalid password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return 0, errors.New("Password doesn't match")
	}

	return user.ID, nil
}
