package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var db = sqlx.MustConnect("sqlite3", "strava.db")

func IsEmptyResultSet(e string) bool {
	return strings.Contains(e, "no rows in result set")
}

// StravaAuth represents db table `strava_auth`
type StravaAuth struct {
	AccessToken  string `db:"access_token"`
	ExpiresAt    uint64 `db:"access_token_expires_at"`
	RefreshToken string `db:"refresh_token"`
	AthleteId    uint64 `db:"athlete_id"`
}

func (s *StravaAuth) Exists() bool {
	return s.AthleteId > 0
}

func (s *StravaAuth) IsExpired() bool {
	now := time.Now()
	return now.Unix() > int64(s.ExpiresAt)
}

// SelectStravaAuth selects and returns a single strava_auth record
func SelectStravaAuth() (*StravaAuth, error) {
	q := `SELECT access_token, access_token_expires_at, refresh_token, athlete_id
            FROM strava_auth
            LIMIT 1`
	var res StravaAuth
	if err := db.Get(&res, q); err != nil {
		return &res, err
	}
	return &res, nil
}

// InsertStravaAuth inserts a new strava_auth record
func InsertStravaAuth(a StravaAuth) error {
	q := `INSERT INTO strava_auth(access_token, access_token_expires_at, refresh_token, athlete_id)
            VALUES (?, ?, ?, ?)`
	res := db.MustExec(q, a.AccessToken, a.ExpiresAt, a.RefreshToken, a.AthleteId)
	_, err := res.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

func UpdateStravaAuth(sa StravaAuth) error {
	q := `UPDATE strava_auth
            SET access_token=?, access_token_expires_at=?, refresh_token=?
            WHERE athlete_id=?`
	_, err := db.Exec(q, sa.AccessToken, sa.ExpiresAt, sa.RefreshToken, sa.AthleteId)
	if err != nil {
		return err
	}
	return nil
}

// StravaAthlete represent db table `athlete`
type StravaAthlete struct {
	StravaId      uint64 `db:"strava_id"`
	FirstName     string `db:"first_name"`
	LastName      string `db:"last_name"`
	Profile       string `db:"profile"`
	ProfileMedium string `db:"profile_medium"`
}

func (s *StravaAthlete) Exists() bool {
	return s.StravaId > 0
}

// SelectStravaAthleteById selects and returns a single strava athlete record
func SelectStravaAthleteById(athleteId uint64) (*StravaAthlete, error) {
	q := `SELECT strava_id, first_name, last_name, profile, profile_medium
            FROM athlete
            WHERE strava_id=?`
	var res StravaAthlete
	err := db.Get(&res, q, athleteId)
	if err != nil {
		return &res, err
	}
	return &res, nil
}

// InsertStravaAthelete inserts a new strava athlete record
func InsertStravaAthelete(a StravaAthlete) error {
	q := `INSERT INTO athlete(strava_id, first_name, last_name, profile, profile_medium)
            VALUES(?, ?, ?, ?, ?)`
	res := db.MustExec(q, a.StravaId, a.FirstName, a.LastName, a.Profile, a.ProfileMedium)
	_, err := res.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

type RaceActivity struct {
	StravaId  uint64    `db:"strava_id"`
	AthleteId uint64    `db:"strava_athlete_id"`
	Name      string    `db:"name"`
	Distance  float64   `db:"distance"`
	StartDate time.Time `db:"start_date_local"`
	Polyline  string    `db:"polyline"`
}

func (r *RaceActivity) Exists() bool {
	return r.StravaId > 0
}

// InsertRaceActivity inserts a new race_activity record
func InsertRaceActivity(r RaceActivity) error {
	q := `INSERT INTO race_activity(strava_id, strava_athlete_id, name, distance, start_date_local)
            VALUES(?, ?, ?, ?, ?)`
	res := db.MustExec(q, r.StravaId, r.AthleteId, r.Name, r.Distance, r.StartDate)
	_, err := res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

// SelectRaceActivityById selects and returns a single race_activity record by strava_id
func SelectRaceActivityById(stravaId uint64) (*RaceActivity, error) {
	q := `SELECT strava_id, strava_athlete_id, name, distance, start_date_local
            FROM race_activity
            WHERE strava_id=?`
	var res RaceActivity

	err := db.Get(&res, q, stravaId)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

func SelectAllRaces() ([]RaceActivity, error) {
	q := `SELECT strava_id, strava_athlete_id, name, distance, start_date_local
            FROM race_activity
            ORDER BY start_date_local DESC`
	var res []RaceActivity

	rows, err := db.Query(q)
	if err != nil {
		return res, fmt.Errorf("Failed to execute SelectAllRaces query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r RaceActivity
		err := rows.Scan(&r.StravaId, &r.AthleteId, &r.Name, &r.Distance, &r.StartDate)
		if err != nil {
			return res, fmt.Errorf("Error scanning race activity rows: %w", err)
		}
		res = append(res, r)
	}

	if err := rows.Err(); err != nil {
		return res, err
	}

	return res, nil
}
