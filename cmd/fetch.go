package cmd

import (
	"fmt"
	"time"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/strava"
	"github.com/spf13/cobra"
)

func dateTimeToEpoch(dt string) (int64, error) {
	t, err := time.Parse(time.RFC3339, dt)
	if err != nil {
		fmt.Println("unable to parse latest activity datetime: ", err)
		return 0, err
	}
	return t.Unix(), nil
}

func getLatestActivityEpoch(athleteId uint64) (int64, error) {
	res, err := db.SelectLatestActivityDateTime(athleteId)
	if err != nil {
		fmt.Println("unable to select latest activity datetime: ", err)
		return 0, err
	}

	if res == "" {
		return 0, nil
	}

	return dateTimeToEpoch(res)
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

		latestActivityEpoch, err := getLatestActivityEpoch(stravaAuth.AthleteId)
		if err != nil {
			fmt.Println(err)
			return
		}

		client := strava.NewClient(stravaAuth.AccessToken)
		var page uint16
		var perPage uint8 = 200
		var latestActivityDateTime string

		params := strava.ReqParams{
			Page:    page,
			PerPage: perPage,
			After:   latestActivityEpoch,
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
				// TODO: figure out a better way to do this
				err = db.InsertRaceActivity(db.RaceActivity{
					StravaId:    a.Id,
					AthleteId:   stravaAuth.AthleteId,
					Name:        a.Name,
					StartDate:   a.StartDateLocal,
					Distance:    a.Distance,
					MovingTime:  a.MovingTime,
					ElapsedTime: a.ElapsedTime,
					Polyline:    a.Map.SummaryPolyline,
				})
				if err != nil {
					fmt.Println("unable to insert new race activity", err)
					return
				}
				fmt.Println(a.Name)
			}
			latestActivityDateTime = activities[activitiesLen-1].StartDateLocal
		}

		if latestActivityDateTime != "" {
			currEpoch, err := dateTimeToEpoch(latestActivityDateTime)
			if err != nil {
				fmt.Println(err)
			}
			if currEpoch != latestActivityEpoch {
				err = db.UpdateLatestActivityDateTime(stravaAuth.AthleteId, latestActivityDateTime)
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		fmt.Println("-- done ---")
	},
}
