package main

import (
	"fmt"
	"go-organizer/backend/connections"
	"go-organizer/backend/handlers"
	"go-organizer/backend/logger"
	"go-organizer/backend/middlewares"
	"go-organizer/backend/templmanager"
	"go-organizer/backend/utils"
	"net/http"

	"github.com/gorilla/mux"
)

// TODO
// Show or hide hidden files (starting with ".")

// NotifyWebServer acts as web server for sending notifications to Matrix server
func main() {
	port := 8080

	// Init logger
	logger.InitLogger()
	_logger := logger.Logger

	// Init database connection
	connections.InitDatabaseConnection()

	// Init session
	utils.InitSessionStore()

	// Load templates
	templmanager.LoadTemplates()

	// Init connection to S3
	connections.BuildS3Connection()

	r := mux.NewRouter()

	// serve static files (css, js)
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))).Name("assets")

	r.HandleFunc("/", handlers.Home).Methods(http.MethodGet).Name("home")
	r.HandleFunc("/login", handlers.Login).Methods(http.MethodGet, http.MethodPost).Name("login")
	r.HandleFunc("/logout", handlers.Logout).Methods(http.MethodGet).Name("logout")

	r.HandleFunc("/user/update", handlers.Update).Methods(http.MethodPost)

	r.HandleFunc("/object/{path:.*}", handlers.GetObject).Methods(http.MethodGet)
	r.HandleFunc("/object/create", handlers.CreateObject).Methods(http.MethodPost)
	r.HandleFunc("/object/delete", handlers.DeleteObject).Methods(http.MethodPost)

	r.Use(middlewares.Session, middlewares.Authentication)

	_logger.Infof("Started server http://localhost:%d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
