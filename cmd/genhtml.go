package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/ddominguez/run-david-run/db"
	"github.com/ddominguez/run-david-run/page"
	"github.com/ddominguez/run-david-run/utils"
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

		indexTmpl := page.New([]string{"templates/base.html", "templates/index.html"})
		raceTmpl := page.New([]string{"templates/base.html", "templates/race.html"})

		// generate race files
		for _, a := range activities {
			raceYear, err := a.RaceYear()
			if err != nil {
				fmt.Println(err)
				return
			}

			racefile := path.Join("./dist", fmt.Sprintf("%d", raceYear), a.NameSlugified(), "index.html")
			if err := os.MkdirAll(path.Dir(racefile), 0770); err != nil {
				fmt.Printf("failed to create path %s\n", err)
				return
			}

			startDate, err := a.StartDateFormatted()
			if err != nil {
				fmt.Println(err)
				return
			}

			data := page.RaceData{
				Name:      a.Name,
				StartDate: startDate,
				Distance:  utils.ActivityDistance(a.Distance),
				Pace:      utils.ActivityPace(a.Distance, a.ElapsedTime),
				Time:      utils.TimeFormatted(a.ElapsedTime),
				MapboxUrl: utils.MapboxURL(a.Polyline),
			}
			err = raceTmpl.Generate(racefile, "base", data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		// generate index file
		data := struct {
			Activities  []db.RaceActivity
			IsGenerated bool
		}{
			Activities:  activities,
			IsGenerated: true,
		}
		indexFile := path.Join("./dist", "index.html")
		err = indexTmpl.Generate(indexFile, "base", data)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}
