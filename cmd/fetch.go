package cmd

import (
	"fmt"
	"time"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/strava"
	"github.com/spf13/cobra"
)

func getLatestActivityEpochTS(athleteId uint64) (int64, error) {
	res, err := db.SelectLatestActivityTimeStamp(athleteId)
	if err != nil {
		fmt.Println("unable to select latest activity timestamp: ", err)
		return 0, err
	}

	if res == "" {
		return 0, nil
	}

	t, err := time.Parse(time.RFC3339, res)
	if err != nil {
		fmt.Println("unable to parse latest activity timestamp: ", err)
		return 0, err
	}
	return t.Unix(), nil
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch and save Strava race activities",
	Long: "fetch will request activities from Strava and \n." +
		"save the race activities.",
	Run: func(cmd *cobra.Command, args []string) {
		stravaAuth, err := db.SelectStravaAuth()
		if err != nil {
			fmt.Println(err)
		}
		if !stravaAuth.Exists() {
			fmt.Println("strava auth user does not exist")
			return
		}
		if stravaAuth.IsExpired() {
			fmt.Println("strava access token is expired")
			return
		}

		latest_ts, err := getLatestActivityEpochTS(stravaAuth.AthleteId)
		if err != nil {
			fmt.Println(err)
			return
		}

		client := strava.NewClient(stravaAuth.AccessToken)
		var page uint16
		var perPage uint8 = 200
		var last_activity_ts string

		params := strava.ReqParams{
			Page:    page,
			PerPage: perPage,
			After:   latest_ts,
		}

		for page = 1; true; page++ {
			params.Page = page
			activities, err := strava.GetActivities(client, params)
			if err != nil {
				fmt.Println(err)
			}
			activitiesLen := len(activities)
			if activitiesLen == 0 {
				fmt.Println("no more activities")
				break
			}
			for _, a := range activities {
				if !a.IsRace() {
					continue
				}
				sid, err := db.SelectRaceActivityId(a.Id)
				if err != nil && !db.IsEmptyResultSet(err.Error()) {
					fmt.Println(err)
					return
				}
				if sid > 0 {
					fmt.Printf("--- strava activity id %d already exists ---\n", a.Id)
					continue
				}
				err = db.InsertRaceActivity(db.RaceActivity{
					StravaId:  a.Id,
					AthleteId: stravaAuth.AthleteId,
					Name:      a.Name,
					StartDate: a.StartDateLocal,
					Distance:  a.Distance,
					Polyline:  a.Map.SummaryPolyline,
				})
				if err != nil {
					fmt.Println("unable to insert new race activity", err)
					return
				}
				fmt.Println(a.Name)
			}
			last_activity_ts = activities[activitiesLen-1].StartDateLocal
		}

		if last_activity_ts != "" {
			err := db.UpdateLatestActivityTimeStamp(stravaAuth.AthleteId, last_activity_ts)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("--- updated last activity timestamp ---")
		}

		fmt.Println("-- done ---")
	},
}
