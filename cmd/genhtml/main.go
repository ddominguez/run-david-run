package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"time"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/strava"
)

type raceDetails struct {
	Name      string
	DateTime  string
	Distance  string
	MapboxURL string
	Pace      string
	Time      string
}

type indexDetails struct {
	Name     string
	NameSlug string
	RaceYear int
}

var projectName = "run-david-run"
var distDir = "dist"

func currWd() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if path.Base(wd) != projectName {
		return "", fmt.Errorf("must be called from project root")
	}

	return wd, nil
}

func getStravaAuth(pgxDB *db.PgxConn) (db.StravaAuth, error) {
	stravaAuth, err := db.SelectStravaAuth(pgxDB)
	if err != nil {
		log.Println("strava auth data not found.", err)
	}

	oauth := strava.Authorization{
		ClientId:     os.Getenv("STRAVA_CLIENT_ID"),
		ClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
		RedirectUri:  "http://localhost:8080/callback",
		Scope:        "activity:read_all",
	}

	if stravaAuth.IsExpired() {
		log.Println("Access token is expired. Requesting a new one.")
		tkResp, err := oauth.RefreshToken(stravaAuth.RefreshToken)
		if err != nil {
			return stravaAuth, err
		}
		stravaAuth = db.StravaAuth{
			AccessToken:  tkResp.AccessToken,
			ExpiresAt:    tkResp.ExpiresAt,
			RefreshToken: tkResp.RefreshToken,
			AthleteId:    stravaAuth.AthleteId,
		}
		if err := db.UpdateStravaAuth(pgxDB, stravaAuth); err != nil {
			return stravaAuth, err
		}
	}

	return stravaAuth, nil
}

func parsedTmpls(wd string, tmplFiles []string) *template.Template {
	var files []string
	for i := range tmplFiles {
		files = append(files, path.Join(wd, tmplFiles[i]))
	}
	return template.Must(template.New("base").ParseFiles(files...))
}

func main() {
	started := time.Now()
	workingDir, err := currWd()
	if err != nil {
		log.Fatalln(err)
	}

	dbUrl := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	pgxDB, err := db.NewPgxConn(dbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer pgxDB.Pool.Close()

	if err := pgxDB.Pool.Ping(context.Background()); err != nil {
		log.Fatalln(err)
	}

	stravaAuth, err := getStravaAuth(pgxDB)
	if err != nil {
		log.Fatalln(err)
	}

	raceTmpl := parsedTmpls(workingDir, []string{"templates/base.html", "templates/race.html"})
	indexTmpl := parsedTmpls(workingDir, []string{"templates/base.html", "templates/index.html"})

	var races []indexDetails

	stravaClient := strava.NewClient(stravaAuth.AccessToken)
	for page := 1; true; page++ {
		activities, err := strava.GetActivities(stravaClient, uint16(page), 200)
		if err != nil {
			log.Println(err)
		}
		if len(activities) == 0 {
			log.Println("No more activities")
			break
		}
		for _, a := range activities {
			if !a.IsRace() {
				continue
			}

			// save race activities for index
			races = append(races, indexDetails{
				Name:     a.Name,
				NameSlug: a.NameSlugified(),
				RaceYear: a.StartDateLocal.Year(),
			})

			// build race activity template
			activityFilePath := path.Join(fmt.Sprintf("%d", a.StartDateLocal.Year()), a.NameSlugified(), "index.html")

			if err := os.MkdirAll(path.Join(workingDir, distDir, path.Dir(activityFilePath)), 0770); err != nil {
				log.Fatalln(err)
			}

			file, err := os.Create(path.Join(workingDir, distDir, activityFilePath))
			if err != nil {
				log.Fatalln(err)
			}

			data := raceDetails{
				Name:      a.Name,
				DateTime:  a.StartDateLocal.Format(time.RFC1123),
				Distance:  a.DistanceInMiles(),
				MapboxURL: a.MapboxURL(),
				Pace:      a.Pace(),
				Time:      a.TimeFormatted(),
			}

			if err := raceTmpl.Execute(file, data); err != nil {
				log.Fatalln(err)
			}
			file.Close()
			log.Printf("created %s", activityFilePath)
		}

	}

	// build index file
	data := struct {
		Activities []indexDetails
	}{
		Activities: races,
	}
	file, err := os.Create(path.Join(workingDir, distDir, "index.html"))
	if err != nil {
		log.Fatalln(err)
	}
	if err := indexTmpl.Execute(file, data); err != nil {
		log.Fatalln(err)
	}
	file.Close()
	log.Println("created index.html")

	fmt.Println("Process took: ", time.Since(started))
}
