package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/strava"
	"github.com/spf13/cobra"
)

var oauthResp strava.AuthTokenResp

// getAccessToken will make an oauth request to strava and return a formatted response.
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

func getStravaClientCreds() (string, string, error) {
	clientId := os.Getenv("STRAVA_CLIENT_ID")
	if clientId == "" {
		return "", "", fmt.Errorf("missing strava client id")
	}

	clientSecret := os.Getenv("STRAVA_CLIENT_SECRET")
	if clientSecret == "" {
		return "", "", fmt.Errorf("missing strava client secret")
	}

	return clientId, clientSecret, nil
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

func newDBStravaAuth(authResp strava.AuthTokenResp) db.StravaAuth {
	return db.StravaAuth{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    authResp.ExpiresAt,
		AthleteId:    authResp.Athlete.Id,
	}
}

var newTokenCmd = &cobra.Command{
	Use:   "newtoken",
	Short: "Get new access and refresh tokens from Strava",
	Long: "newtoken will request new access and refresh tokens from Strava.\n" +
		"The access token is needed for Strava API requests.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		clientId, clientSecret, err := getStravaClientCreds()
		if err != nil {
			fmt.Println(err)
			return
		}
		oauth := strava.Authorization{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			RedirectUri:  "http://localhost:8080/callback",
			Scope:        "activity:read_all",
		}

		oauthUser, err := db.SelectStravaAuth()
		if err != nil {
			if db.IsEmptyResultSet(err.Error()) {
				fmt.Println("-- strava auth user not found")
			} else {
				fmt.Println(err)
				return
			}
		}

		if oauthUser.AthleteId == 0 {
			fmt.Println("-- inserting strava auth")
			runServer(oauth)
			oauthData := newDBStravaAuth(oauthResp)
			err = db.InsertStravaAuth(oauthData)
		} else {
			fmt.Println("-- refreshing access token")
			oauthResp, err = oauth.RefreshToken(oauthUser.RefreshToken)
			oauthData := newDBStravaAuth(oauthResp)
			err = db.UpdateStravaAuth(oauthData)
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = db.SelectStravaAthleteById(oauthUser.AthleteId)
		athleteExists := true
		if err != nil {
			if db.IsEmptyResultSet(err.Error()) {
				athleteExists = false
				fmt.Println("-- strava athlete not found")
			} else {
				fmt.Println(err)
				return
			}
		}

		if !athleteExists && oauthResp.Athlete.Id > 0 {
			fmt.Println("-- inserting strava athlete")
			_ = db.InsertStravaAthelete(db.StravaAthlete{
				StravaId:      oauthResp.Athlete.Id,
				FirstName:     oauthResp.Athlete.FirstName,
				LastName:      oauthResp.Athlete.LastName,
				Profile:       oauthResp.Athlete.Profile,
				ProfileMedium: oauthResp.Athlete.ProfileMedium,
			})
		}

		fmt.Println("-- new strava access token acquired")
	},
}
