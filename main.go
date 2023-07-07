package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/ddominguez/run-david-run/db"
)

var pgxDB *db.PgxConn

//go:embed dist/*.css
var distFiles embed.FS

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	activities, err := db.SelectAllRaces(pgxDB)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data := struct {
		Activities []db.RaceActivity
	}{
		Activities: activities,
	}
	tmplFiles := []string{
		"./templates/base.html",
		"./templates/index.html",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles(tmplFiles...))
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func main() {
	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@localhost:5432/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	var err error
	pgxDB, err = db.NewPgxConn(dbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer pgxDB.Pool.Close()
	if err := pgxDB.Pool.Ping(context.Background()); err != nil {
		log.Fatalln(err)
	}

	static, err := fs.Sub(distFiles, "dist")
	if err != nil {
		log.Fatalln(err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))

	http.HandleFunc("/", handleIndex)
	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
