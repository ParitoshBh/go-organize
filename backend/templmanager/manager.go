package templmanager

import (
	"errors"
	"fmt"
	"go-organizer/backend/logger"
	"html/template"
	"net/http"
	"path/filepath"
)

var templates map[string]*template.Template
var mainTmpl = `{{define "main" }} {{ template "base" . }} {{ end }}`

func LoadTemplates() (err error) {
	_logger := logger.Logger

	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	layoutFiles, err := filepath.Glob("templates/base/*.html")
	if err != nil {
		return err
	}

	includeFiles, err := filepath.Glob("templates/*.html")
	if err != nil {
		return err
	}

	mainTemplate := template.New("main").Funcs(template.FuncMap{
		"IsLastInRange":  isLastInRange,
		"GenerateAvatar": generateAvatar,
	})

	mainTemplate, err = mainTemplate.Parse(mainTmpl)
	if err != nil {
		_logger.Fatal(err.Error())
	}

	for _, file := range includeFiles {
		fileName := filepath.Base(file)
		files := append(layoutFiles, file)

		templates[fileName], err = mainTemplate.Clone()
		if err != nil {
			return err
		}

		templates[fileName] = template.Must(templates[fileName].ParseFiles(files...))
	}

	_logger.Info("Templates loaded successfully")
	return nil
}

func RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("The template %s does not exist.", name), http.StatusInternalServerError)
		return errors.New("Template doesn't exist")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl.Execute(w, data)

	return nil
}
