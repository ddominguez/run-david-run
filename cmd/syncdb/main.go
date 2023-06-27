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
// once the strava access token is requested and saved
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
		if err := db.InsertStravaAuth(pgxConn, resp.AccessToken, resp.RefreshToken, resp.ExpiresAt, resp.Athlete.Id); err != nil {
			log.Println(err)
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

	// If access token does not exist get new access token data and save to db
	// Save athlete data if not exists
	if !stravaAuth.Exists() {
		auth := strava.Authorization{
			ClientId:     os.Getenv("STRAVA_CLIENT_ID"),
			ClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
			RedirectUri:  "http://localhost:8080/callback",
			Scope:        "activity:read_all",
		}
		log.Println("Click to authorize on strava:", auth.Url())
		runServer(pgxDB, auth)

		stravaAuth, err = db.SelectStravaAuth(pgxDB)
		if err != nil {
			log.Fatalln("strava auth data not found.", err)
		}
	}

	// if expired use refresh token to new access token and save to db

	// use valid access token to fetch race activities from stava and save to db
	log.Println(stravaAuth)
}
