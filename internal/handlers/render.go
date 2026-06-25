package handlers

import (
	"html/template"
	"log"
	"net/http"
)

// RenderTemplate parses and executes templates, simplifying handler boilerplate.
func RenderTemplate(w http.ResponseWriter, data interface{}, files ...string) {
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("Template parsing error: %v (files: %v)", err, files)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// RenderTemplateName parses and executes a specific template by name.
func RenderTemplateName(w http.ResponseWriter, templateName string, data interface{}, files ...string) {
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("Template parsing error: %v (files: %v)", err, files)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
