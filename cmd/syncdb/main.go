package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/strava"
)

// runServer creates new a new http server and will shutdown
// once strava token and athlete data is requested and saved.
func runServer(pgxConn *db.PgxConn, auth strava.Authorization) {
	log.Println("Http server started")
	log.Println("Listening on http://localhost:8080")
	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			log.Printf("Request URL (%s) is missing code param", r.URL)
			http.Error(w, "Missing code param", http.StatusBadRequest)
			return
		}
		resp, err := auth.ReqAccessToken(code)
		if err != nil {
			log.Println(err)
			http.Error(w, "bad request.", http.StatusBadRequest)
			return
		}

		if err := db.InsertStravaAuth(pgxConn, db.StravaAuth{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			ExpiresAt:    resp.ExpiresAt,
			AthleteId:    resp.Athlete.Id,
		}); err != nil {
			log.Println(err)
		}

		athlete, err := db.SelectStravaAthleteById(pgxConn, resp.Athlete.Id)
		if err != nil {
			log.Println(err)
		}

		if !athlete.Exists() {
			if err := db.InsertStravaAthelete(pgxConn, db.StravaAthlete{
				StravaId:      resp.Athlete.Id,
				FirstName:     resp.Athlete.FirstName,
				LastName:      resp.Athlete.LastName,
				Profile:       resp.Athlete.Profile,
				ProfileMedium: resp.Athlete.ProfileMedium,
			}); err != nil {
				log.Println(err)
			}
		}
		w.Write([]byte("thank you."))
		cancel()
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	select {
	case <-ctx.Done():
		srv.Shutdown(ctx)
		log.Println("Http server shutdown")
	}
}

func main() {
	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@localhost:5432/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
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

	if !stravaAuth.Exists() {
		log.Println("Click to authorize on strava:", oauth.Url())
		runServer(pgxDB, oauth)

		stravaAuth, err = db.SelectStravaAuth(pgxDB)
		if err != nil {
			log.Fatalln("strava auth data not found.", err)
		}
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

	stravaClient := strava.NewClient(stravaAuth.AccessToken)

	page := uint16(1)
	activities, err := strava.GetActivities(stravaClient, page, 200)
	if err != nil {
		log.Println(err)
	}
	if len(activities) == 0 {
		log.Println("No more activities")
	}
	for _, a := range activities {
		if !a.IsRace() {
			continue
		}
		ra, err := db.SelectRaceActivityById(pgxDB, a.Id)
		if err != nil {
			log.Fatalln(err)
		}

		if ra.Exists() {
			log.Printf("strava activity id %d already exists\n", a.Id)
			continue
		}

		if err := db.InsertRaceActivity(pgxDB, db.RaceActivity{
			StravaId:  a.Id,
			Name:      a.Name,
			NameSlug:  a.NameSlugified(),
			Distance:  a.Distance,
			StartDate: a.StartDateLocal,
			AthleteId: a.Athlete.Id,
		}); err != nil {
			log.Fatalln(err)
		}
		log.Printf("strava id: %d, race name: %s \n", a.Id, a.Name)
	}
}
