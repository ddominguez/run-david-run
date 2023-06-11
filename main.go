package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmplFiles := []string{
		"./templates/base.html",
		"./templates/index.html",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles(tmplFiles...))
	if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
		log.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func handleRaces(w http.ResponseWriter, r *http.Request) {
	tmplFiles := []string{
		"./templates/base.html",
		"./templates/races.html",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles(tmplFiles...))
	if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
		log.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/races", handleRaces)
	http.HandleFunc("/", handleIndex)
	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
