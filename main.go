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
	"strconv"
	"strings"
	"time"

	"github.com/ddominguez/run-david-run/db"
)

var pgxDB *db.PgxConn

//go:embed dist/*.css
var distFiles embed.FS

func handleRaces(w http.ResponseWriter, r *http.Request) {
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
func handleRaceDetails(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	fmt.Fprint(w, fmt.Sprintf("Year: %s, Slug: %s", params[0], params[1]))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	valid := validReqPath(r.URL.Path)
	if !valid {
		http.NotFound(w, r)
		return
	}

	if r.URL.Path == "/" {
		handleRaces(w, r)
		return
	}

	handleRaceDetails(w, r)
}

func validReqPath(path string) bool {
	if path == "/" {
		return true
	}

	params := strings.Split(strings.Trim(path, "/"), "/")
	if len(params) != 2 {
		return false
	}

	y, err := strconv.Atoi(params[0])
	if err != nil {
		return false
	}

	// First race was in 2014
	minY := 2014
	t := time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC)
	if t.Year() < minY || t.Year() > time.Now().Year() {
		return false
	}

	return true
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
