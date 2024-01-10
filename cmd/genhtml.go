package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/page"
	"github.com/spf13/cobra"
)

var genHtmlCmd = &cobra.Command{
	Use:   "genhtml",
	Short: "Generate html for saved race activities",
	Long:  "genhtml will generate static html for saved race activities.",
	Run: func(cmd *cobra.Command, args []string) {
		activities, err := db.AllRaceActivities()
		if err != nil {
			fmt.Println(err)
			return
		}

		// build race files
		for _, a := range activities {
			raceYear, _ := a.RaceYear()
			fp := path.Join("./dist", fmt.Sprintf("%d", raceYear), a.NameSlugified(), "index.html")
			if err := os.MkdirAll(path.Dir(fp), 0770); err != nil {
				fmt.Println(err)
				return
			}

			file, err := os.Create(fp)
			if err != nil {
				fmt.Println(err)
				return
			}
			racedt, err := a.StartDateFormatted()
			if err != nil {
				fmt.Println(err)
				racedt = a.StartDate
			}

			data := page.RaceData{
				Name:      a.Name,
				StartDate: racedt,
				Distance:  a.DistanceInMiles(),
				Pace:      a.Pace(),
				Time:      a.TimeFormatted(),
				MapboxUrl: mapboxURL(a.Polyline),
			}

			tmpl := page.New([]string{"templates/base.html", "templates/race.html"})
			err = tmpl.Render(file, "base", data)
			if err != nil {
				fmt.Println("failed to execute race template", err)
			}

			file.Close()
			fmt.Printf("created %s\n", fp)
		}

		// build index file
		data := struct {
			Activities  []db.RaceActivity
			IsGenerated bool
		}{
			Activities:  activities,
			IsGenerated: true,
		}
		file, err := os.Create(path.Join("./dist", "index.html"))
		if err != nil {
			fmt.Println(err)
			return
		}

		page := page.New([]string{"templates/base.html", "templates/index.html"})
		err = page.Render(file, "base", data)
		if err != nil {
			fmt.Println("failed to execute index template", err)
		}

		file.Close()
		fmt.Printf("created %s\n", file.Name())
	},
}
