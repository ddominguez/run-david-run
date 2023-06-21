package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ddominguez/run-david-run/db"
)

func main() {
	// get access token from db
	// -- if access token exists check if expired
	// -- if expired use refresh token to new access token and save to db

	// if access token does not exist get new access token data and save to db

	// use valid access token to fetch race activities from stava and save to db

	// save athlete data if not exists

	dbUrl := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	pgxDB, err := db.NewPgxConn(dbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer pgxDB.Pool.Close()

	sa, err := db.SelectStravaAuth(pgxDB)
	if err != nil {
		log.Println("strava auth data not found.", err)
	}
	if !sa.Exists() {
		log.Println("get new strava auth data")
	}

}
