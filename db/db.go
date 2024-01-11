package db

import (
	"regexp"
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
	StravaId               uint64 `db:"strava_id"`
	FirstName              string `db:"first_name"`
	LastName               string `db:"last_name"`
	Profile                string `db:"profile"`
	ProfileMedium          string `db:"profile_medium"`
	LatestActivityDateTime string `db:"latest_activity_datetime"`
}

func (s *StravaAthlete) Exists() bool {
	return s.StravaId > 0
}

func SelectLatestActivityDateTime(athleteId uint64) (string, error) {
	q := `SELECT latest_activity_datetime FROM athlete WHERE strava_id=?`
	var dt string
	err := db.Get(&dt, q, athleteId)
	if err != nil {
		return dt, err
	}
	return dt, nil
}

func UpdateLatestActivityDateTime(athleteId uint64, dt string) error {
	q := `UPDATE athlete SET latest_activity_datetime=? WHERE strava_id=?`
	res, err := db.Exec(q, dt, athleteId)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}
	return nil
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
	q := `INSERT INTO athlete(strava_id, first_name, last_name, profile, profile_medium) VALUES(?, ?, ?, ?, ?)`
	res := db.MustExec(q, a.StravaId, a.FirstName, a.LastName, a.Profile, a.ProfileMedium)
	_, err := res.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

type RaceActivity struct {
	StravaId    uint64  `db:"strava_id"`
	AthleteId   uint64  `db:"strava_athlete_id"`
	Name        string  `db:"name"`
	Distance    float64 `db:"distance"`
	MovingTime  uint32  `db:"moving_time"`
	ElapsedTime uint32  `db:"elapsed_time"`
	StartDate   string  `db:"start_date_local"`
	Polyline    string  `db:"polyline"`
}

func (r *RaceActivity) Exists() bool {
	return r.StravaId > 0
}

func (r RaceActivity) StartDateFormatted() (string, error) {
	t, err := time.Parse(time.RFC3339, r.StartDate)
	if err != nil {
		return "", err
	}
	return t.Format(time.RFC1123), nil
}

// RaceYear parses the StartDate string and returns the year of activity
func (r *RaceActivity) RaceYear() (int, error) {
	t, err := time.Parse(time.RFC3339, r.StartDate)
	if err != nil {
		return 0, err
	}
	return t.Year(), nil
}

var re = regexp.MustCompile("[^a-z0-9]+")

func (r *RaceActivity) NameSlugified() string {
	return strings.Trim(re.ReplaceAllString(strings.ToLower(r.Name), "-"), "-")
}

// InsertRaceActivity inserts a new race_activity record
func InsertRaceActivity(r RaceActivity) error {
	q := `INSERT INTO race_activity(
            strava_id,
            strava_athlete_id,
            name,
            distance,
            moving_time,
            elapsed_time,
            start_date_local,
            polyline
        ) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`
	res := db.MustExec(
		q, r.StravaId, r.AthleteId, r.Name, r.Distance, r.MovingTime,
		r.ElapsedTime, r.StartDate, r.Polyline,
	)
	_, err := res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func SelectRaceActivityId(stravaId uint64) (uint64, error) {
	var sid uint64
	err := db.Get(&sid, "SELECT strava_id from race_activity where strava_id=?", stravaId)
	if err != nil {
		return sid, err
	}
	return sid, nil
}

func SelectRaceActivityById(stravaId uint64) (RaceActivity, error) {
	var resp RaceActivity
	err := db.Get(&resp, "SELECT * from race_activity where strava_id=?", stravaId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func AllRacesForIndex() ([]RaceActivity, error) {
	q := `SELECT strava_id, name, start_date_local FROM race_activity ORDER BY start_date_local DESC`
	var res []RaceActivity
	err := db.Select(&res, q)
	if err != nil {
		return res, err
	}
	return res, nil
}

func AllRaceActivities() ([]RaceActivity, error) {
	var res []RaceActivity
	err := db.Select(&res, "SELECT * from race_activity ORDER BY start_date_local DESC")
	if err != nil {
		return res, err
	}
	return res, nil
}
