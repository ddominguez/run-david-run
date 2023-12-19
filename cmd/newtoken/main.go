package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/strava"
)

var oauthResp strava.AuthTokenResp

func getAccessToken(code string, oauth strava.Authorization) (strava.AuthTokenResp, error) {
	if code == "" {
		return strava.AuthTokenResp{}, fmt.Errorf("missing code param")
	}

	resp, err := oauth.ReqAccessToken(code)
	if err != nil {
		return strava.AuthTokenResp{}, err
	}

	return resp, nil
}

func runServer(oauth strava.Authorization) strava.AuthTokenResp {
	fmt.Println("-- waiting for strava authorization --")
	fmt.Println(oauth.Url())

	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		var err error
		oauthResp, err = getAccessToken(r.URL.Query().Get("code"), oauth)
		if err != nil {
			fmt.Println(err)
			http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
			cancel()
			return
		}
		w.Write([]byte("success"))
		cancel()
	})

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	select {
	case <-ctx.Done():
		srv.Shutdown(ctx)
		fmt.Println("-- strava authorization complete --")
	}

	return oauthResp
}

func main() {
	stravaClientId := os.Getenv("STRAVA_CLIENT_ID")
	if stravaClientId == "" {
		fmt.Println("missing strava client id")
		return
	}
	stravaClientSecret := os.Getenv("STRAVA_CLIENT_SECRET")
	if stravaClientSecret == "" {
		fmt.Println("missing strava secret")
		return
	}
	oauth := strava.Authorization{
		ClientId:     stravaClientId,
		ClientSecret: stravaClientSecret,
		RedirectUri:  "http://localhost:8080/callback",
		Scope:        "activity:read_all",
	}

	runServer(oauth)

	if oauthResp.AccessToken == "" {
		fmt.Println("unable to get access token")
		return
	}

	var err error
	oauthData := db.StravaAuth{
		AccessToken:  oauthResp.AccessToken,
		RefreshToken: oauthResp.RefreshToken,
		ExpiresAt:    oauthResp.ExpiresAt,
		AthleteId:    oauthResp.Athlete.Id,
	}

	oauthUser, err := db.SelectStravaAuth()
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			fmt.Println("-- strava auth user not found")
		} else {
			fmt.Println(err)
			return
		}
	}

	if oauthUser.AthleteId == 0 {
		fmt.Println("-- inserting strava auth")
		err = db.InsertStravaAuth(oauthData)
	} else {
		fmt.Println("-- updating strava auth")
		err = db.UpdateStravaAuth(oauthData)
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("-- strava access token acquired")
}
