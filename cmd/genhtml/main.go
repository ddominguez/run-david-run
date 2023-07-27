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

	var stravaAuth db.StravaAuth

	stravaAuth, err = db.SelectStravaAuth(pgxDB)
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
			log.Fatalln(err)
		}
		stravaAuth = db.StravaAuth{
			AccessToken:  tkResp.AccessToken,
			ExpiresAt:    tkResp.ExpiresAt,
			RefreshToken: tkResp.RefreshToken,
			AthleteId:    stravaAuth.AthleteId,
		}
		if err := db.UpdateStravaAuth(pgxDB, stravaAuth); err != nil {
			log.Fatalln(err)
		}
	}

	raceTmplFiles := []string{
		path.Join(workingDir, "templates/base.html"),
		path.Join(workingDir, "templates/race.html"),
	}
	raceTmpl := template.Must(template.New("base").ParseFiles(raceTmplFiles...))

	indexTmplFiles := []string{
		path.Join(workingDir, "templates/base.html"),
		path.Join(workingDir, "templates/index.html"),
	}
	indexTmpl := template.Must(template.New("base").ParseFiles(indexTmplFiles...))

	var races []strava.Activity

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

			// gather activities into slice
			a.NameSlug = a.NameSlugified()
			races = append(races, a)

			// build race activity template
			activityFilePath := path.Join(fmt.Sprintf("%d", a.StartDateLocal.Year()), a.NameSlugified(), "index.html")

			if err := os.MkdirAll(path.Join(workingDir, distDir, path.Dir(activityFilePath)), 0770); err != nil {
				log.Fatalln(err)
			}

			file, err := os.Create(path.Join(workingDir, distDir, activityFilePath))
			if err != nil {
				log.Fatalln(err)
			}

			data := struct {
				Activity  strava.Activity
				DateTime  string
				Distance  string
				MapboxURL string
				Pace      string
				Time      string
			}{
				Activity:  a,
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

		if page == 1 {
			break
		}
	}

	// build index file
	data := struct {
		Activities []strava.Activity
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
