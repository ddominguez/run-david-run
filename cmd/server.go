package cmd

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/page"
	"github.com/ddominguez/run-david-run/utils"
	"github.com/spf13/cobra"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	activities, err := db.AllRaceActivities()
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

	tmpl := page.New([]string{"templates/base.html", "templates/index.html"})
	err = tmpl.Execute(w, "base", data)
	if err != nil {
		fmt.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
func handleActivity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 0)
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

	startDate, err := activity.StartDateFormatted()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := page.RaceData{
		Name:      activity.Name,
		StartDate: startDate,
		Distance:  utils.ActivityDistance(activity.Distance),
		Pace:      utils.ActivityPace(activity.Distance, activity.ElapsedTime),
		Time:      utils.TimeFormatted(activity.ElapsedTime),
		MapboxUrl: utils.MapboxURL(activity.Polyline),
	}

	tmpl := page.New([]string{"templates/base.html", "templates/race.html"})
	err = tmpl.Execute(w, "base", data)
	if err != nil {
		fmt.Println("failed to execute to templates", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func startServer() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/activity/{id}", handleActivity)

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
