package cmd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/page"
	"github.com/spf13/cobra"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	activities, err := db.AllRacesForIndex()
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := struct {
		Activities  []db.RaceActivity
		IsGenerated bool
	}{
		Activities:  activities,
		IsGenerated: false,
	}

	page := page.New([]string{"templates/base.html", "templates/index.html"})
	err = page.Render(w, "base", data)
	if err != nil {
		fmt.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func handleActivity(w http.ResponseWriter, r *http.Request) {
	pathParams := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParams) != 2 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseUint(pathParams[1], 10, 0)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	activity, err := db.SelectRaceActivityById(id)
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}

	racedt, err := activity.StartDateFormatted()
	if err != nil {
		fmt.Println(err)
		racedt = activity.StartDate
	}

	data := page.RaceData{
		Name:      activity.Name,
		StartDate: racedt,
		Distance:  activity.DistanceInMiles(),
		Pace:      activity.Pace(),
		Time:      activity.TimeFormatted(),
		MapboxUrl: mapboxURL(activity.Polyline),
	}

	page := page.New([]string{"templates/base.html", "templates/race.html"})
	err = page.Render(w, "base", data)
	if err != nil {
		fmt.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func startServer() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/activity/", handleActivity)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	port := "8080"
	fmt.Printf("Listening on http://localhost:%s\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		fmt.Println(err)
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "http server for saved race activities",
	Long:  "server will start an http server for saved race activities.",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}
