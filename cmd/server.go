package cmd

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/ddominguez/run-david-run/db"
	"github.com/spf13/cobra"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	activities, err := db.AllRacesForIndex()
	if err != nil {
		fmt.Println(err)
		http.Error(w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	data := struct {
		Activities []db.RaceActivity
	}{
		Activities: activities,
	}
	tmplFiles := []string{
		"templates/base.html",
		"templates/index.html",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles(tmplFiles...))
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		fmt.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func startServer() {
	http.HandleFunc("/", handleIndex)
	port := "8080"
	fmt.Printf("Listening on http://localhost:%s\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		fmt.Println(err)
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "http server for saved race activities",
	Long:  "server will start an http server for saved race activities.",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}
