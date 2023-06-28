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
		if err := db.InsertStravaAuth(
			pgxConn,
			resp.AccessToken,
			resp.RefreshToken,
			resp.ExpiresAt,
			resp.Athlete.Id); err != nil {
			log.Println(err)
		}

		// check if athlete record exists
		athlete, err := db.SelectStravaAthleteById(pgxConn, resp.Athlete.Id)
		if err != nil {
			log.Println(err)
		}

		// create athlete record if not exist
		if !athlete.Exists() {
			if err := db.InsertStravaAthelete(
				pgxConn,
				resp.Athlete.Id,
				resp.Athlete.FirstName,
				resp.Athlete.LastName,
				resp.Athlete.Profile,
				resp.Athlete.ProfileMedium); err != nil {
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
	dbUrl := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	pgxDB, err := db.NewPgxConn(dbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer pgxDB.Pool.Close()

	if err := pgxDB.Pool.Ping(context.Background()); err != nil {
		log.Fatalln(err)
	}

	// Get access token from db
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

	// If access token does not exist get new access token data and save to db
	// Save athlete data if not exists
	if !stravaAuth.Exists() {
		log.Println("Click to authorize on strava:", oauth.Url())
		runServer(pgxDB, oauth)

		stravaAuth, err = db.SelectStravaAuth(pgxDB)
		if err != nil {
			log.Fatalln("strava auth data not found.", err)
		}
	}
	// use valid access token to fetch race activities from stava and save to db
	h := map[string]string{"Authorization": "Bearer " + stravaAuth.AccessToken}
	c := strava.NewClient(h)

	// if expired use refresh token to new access token and save to db
	if stravaAuth.IsExpired() {
		log.Println("Access token is expired. Get a new one.")
		tkResp, err := oauth.RefreshToken(stravaAuth.RefreshToken)
		if err != nil {
			log.Println(err)
			return
		}
		stravaAuth = db.StravaAuth{
			AccessToken:  tkResp.AccessToken,
			ExpiresAt:    tkResp.ExpiresAt,
			RefreshToken: tkResp.RefreshToken,
			AthleteId:    stravaAuth.AthleteId,
		}
		if err := db.UpdateStravaAuth(pgxDB, stravaAuth); err != nil {
			log.Fatalln("Failed to update db table `strava_auth`. " + err.Error())
		}
	}

	page := uint16(1)
	activities, err := strava.GetActivities(c, page, 30)
	if err != nil {
		log.Println(err)
	}
	if len(activities) == 0 {
		log.Println("No more activities")
	}
	for _, a := range activities {
		if a.SportType != "Run" || a.WorkoutType != 1 {
			continue
		}
		log.Println(a.Id, a.Name, a.SportType, a.WorkoutType)
	}
}
