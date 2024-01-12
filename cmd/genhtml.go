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

func generateIndex(tmpl page.Tmpl, activities *[]db.RaceActivity) error {
	data := struct {
		Activities  *[]db.RaceActivity
		IsGenerated bool
	}{
		Activities:  activities,
		IsGenerated: true,
	}
	file, err := os.Create(path.Join("./dist", "index.html"))
	if err != nil {
		return fmt.Errorf("failed to create file: %s", err)
	}

	err = tmpl.Execute(file, "base", data)
	if err != nil {
		return fmt.Errorf("failed to write file: %s", err)
	}

	file.Close()
	fmt.Printf("created %s\n", file.Name())
	return nil
}

func generateRace(tmpl page.Tmpl, activity *db.RaceActivity) error {
	raceYear, _ := activity.RaceYear()
	fp := path.Join("./dist", fmt.Sprintf("%d", raceYear), activity.NameSlugified(), "index.html")
	if err := os.MkdirAll(path.Dir(fp), 0770); err != nil {
		return fmt.Errorf("failed to create path %s", err)
	}

	file, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf("failed to create file: %s", err)
	}
	racedt, err := activity.StartDateFormatted()
	if err != nil {
		fmt.Println(err)
		racedt = activity.StartDate
	}

	data := page.RaceData{
		Name:      activity.Name,
		StartDate: racedt,
		Distance:  utils.ActivityDistance(activity.Distance),
		Pace:      utils.ActivityPace(activity.Distance, activity.ElapsedTime),
		Time:      utils.TimeFormatted(activity.ElapsedTime),
		MapboxUrl: utils.MapboxURL(activity.Polyline),
	}

	err = tmpl.Execute(file, "base", data)
	if err != nil {
		return fmt.Errorf("failed to write file: %s", err)
	}

	file.Close()
	fmt.Printf("created %s\n", fp)
	return nil
}

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
            err := generateRace(raceTmpl, &a)
			if err != nil {
				fmt.Println(err)
                return
			}
		}

		// generate index file
		err = generateIndex(indexTmpl, &activities)
		if err != nil {
			fmt.Println(err)
            return
		}
	},
}
