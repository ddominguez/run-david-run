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
	"github.com/ddominguez/run-david-run/strava"
)

var pgxDB *db.PgxConn

//go:embed dist/*.css
var distFS embed.FS

//go:embed templates/*.html
var tmplFS embed.FS

var oauth = strava.Authorization{
	ClientId:     os.Getenv("STRAVA_CLIENT_ID"),
	ClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
	RedirectUri:  "http://localhost:8080/callback",
	Scope:        "activity:read_all",
}

func getAccessToken() (string, error) {
	stravaAuth, err := db.SelectStravaAuth(pgxDB)
	if err != nil {
		log.Println("strava auth data not found.", err)
		return "", err
	}

	if stravaAuth.IsExpired() {
		log.Println("Access token is expired. Requesting a new one.")
		tkResp, err := oauth.RefreshToken(stravaAuth.RefreshToken)
		if err != nil {
			log.Println(err)
			return "", err
		}
		stravaAuth = db.StravaAuth{
			AccessToken:  tkResp.AccessToken,
			ExpiresAt:    tkResp.ExpiresAt,
			RefreshToken: tkResp.RefreshToken,
			AthleteId:    stravaAuth.AthleteId,
		}
		if err := db.UpdateStravaAuth(pgxDB, stravaAuth); err != nil {
			log.Println(err)
			return "", err
		}
	}
	return stravaAuth.AccessToken, nil
}

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
		"templates/base.html",
		"templates/index.html",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFS(tmplFS, tmplFiles...))
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func handleRaceDetails(w http.ResponseWriter, r *http.Request) {
	pv, err := getPathValues(r.URL.Path)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	activity, err := db.SelectRaceByYearAndSlug(pgxDB, pv.year, pv.slug)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	token, err := getAccessToken()
	if err != nil {
		log.Println("failed to get access token", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	stravaClient := strava.NewClient(token)
	sa, err := strava.GetActivity(stravaClient, activity.StravaId)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := struct {
		Activity  strava.Activity
		DateTime  string
		Distance  string
		MapboxURL string
		Pace      string
		Time      string
	}{
		Activity:  sa,
		DateTime:  sa.StartDateLocal.Format(time.RFC1123),
		Distance:  sa.DistanceInMiles(),
		MapboxURL: sa.MapboxURL(),
		Pace:      sa.Pace(),
		Time:      sa.TimeFormatted(),
	}
	tmplFiles := []string{
		"templates/base.html",
		"templates/race.html",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFS(tmplFS, tmplFiles...))
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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

type racePath struct {
	year int
	slug string
}

func getPathValues(path string) (racePath, error) {
	var res racePath
	params := strings.Split(strings.Trim(path, "/"), "/")
	if len(params) != 2 {
		return res, fmt.Errorf("Invalid values in request path")
	}

	y, err := strconv.Atoi(params[0])
	if err != nil {
		return res, fmt.Errorf("Request contains invalid year")
	}

	res.year = y
	res.slug = params[1]

	return res, nil
}

func validReqPath(path string) bool {
	if path == "/" {
		return true
	}

	pv, err := getPathValues(path)
	if err != nil {
		log.Println(err)
		return false
	}

	// First race was in 2014
	minY := 2014
	t := time.Date(pv.year, time.January, 1, 0, 0, 0, 0, time.UTC)
	if t.Year() < minY || t.Year() > time.Now().Year() {
		return false
	}

	return true
}

func main() {
	dbUrl := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
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

	static, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatalln(err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))

	http.HandleFunc("/", handleIndex)
	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
