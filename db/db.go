package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxConn struct {
	Pool *pgxpool.Pool
}

// NewPgxConn creates and returns a new pool connection
func NewPgxConn(dbUrl string) (*PgxConn, error) {
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return nil, err
	}
	return &PgxConn{Pool: pool}, nil
}

// StravaAuth represents db table `strava_auth`
type StravaAuth struct {
	AccessToken  string
	ExpiresAt    uint64
	RefreshToken string
	AthleteId    uint64
}

func (s *StravaAuth) Exists() bool {
	return s.AthleteId > 0
}

func (s *StravaAuth) IsExpired() bool {
	now := time.Now()
	return now.Unix() > int64(s.ExpiresAt)
}

// SelectStravaAuth selects and returns a single strava_auth record
func SelectStravaAuth(pgxConn *PgxConn) (StravaAuth, error) {
	q := `SELECT access_token, access_token_expires_at, refresh_token, athlete_id FROM strava_auth LIMIT 1`
	var res StravaAuth
	err := pgxConn.Pool.
		QueryRow(context.Background(), q).
		Scan(&res.AccessToken, &res.ExpiresAt, &res.RefreshToken, &res.AthleteId)
	if err != nil {
		return res, err
	}
	return res, nil
}

// InsertStravaAuth inserts a new strava_auth record
func InsertStravaAuth(pgxConn *PgxConn, a StravaAuth) error {
	q := `INSERT INTO strava_auth(access_token, access_token_expires_at, refresh_token, athlete_id) VALUES ($1, $2, $3, $4)`
	_, err := pgxConn.Pool.Exec(context.Background(), q, a.AccessToken, a.ExpiresAt, a.RefreshToken, a.AthleteId)
	if err != nil {
		return err
	}
	return nil
}

func UpdateStravaAuth(pgxConn *PgxConn, sa StravaAuth) error {
	q := `UPDATE strava_auth SET access_token=$1, access_token_expires_at=$2, refresh_token=$3 WHERE athlete_id=$4`
	_, err := pgxConn.Pool.Exec(context.Background(), q, sa.AccessToken, sa.ExpiresAt, sa.RefreshToken, sa.AthleteId)
	if err != nil {
		return err
	}
	return nil
}

// StravaAthlete represent db table `athlete`
type StravaAthlete struct {
	StravaId      uint64
	FirstName     string
	LastName      string
	Profile       string
	ProfileMedium string
}

func (s *StravaAthlete) Exists() bool {
	return s.StravaId > 0
}

// SelectStravaAthleteById selects and returns a single strava athlete record
func SelectStravaAthleteById(pgxConn *PgxConn, athleteId uint64) (StravaAthlete, error) {
	q := `SELECT strava_id, first_name, last_name, profile, profile_medium FROM athlete WHERE strava_id=$1`
	var res StravaAthlete
	err := pgxConn.Pool.
		QueryRow(context.Background(), q, athleteId).
		Scan(&res.StravaId, &res.FirstName, &res.LastName, &res.Profile, &res.ProfileMedium)
	if err != nil {
		return res, nil
	}
	return res, nil
}

// InsertStravaAthelete inserts a new strava athlete record
func InsertStravaAthelete(pgxConn *PgxConn, a StravaAthlete) error {
	q := `INSERT INTO athlete(strava_id, first_name, last_name, profile, profile_medium) VALUES($1, $2, $3, $4, $5)`
	_, err := pgxConn.Pool.Exec(context.Background(), q, a.StravaId, a.FirstName, a.LastName, a.Profile, a.ProfileMedium)
	if err != nil {
		return err
	}
	return nil
}

type RaceActivity struct {
	StravaId  uint64
	AthleteId uint64
	Name      string
	NameSlug  string
	Distance  float64
	StartDate time.Time
}

func (r *RaceActivity) Exists() bool {
	return r.StravaId > 0
}

// InsertRaceActivity inserts a new race_activity record
func InsertRaceActivity(pgxConn *PgxConn, r RaceActivity) error {
	q := `INSERT INTO race_activity(strava_id, strava_athlete_id, name, name_slug, distance, start_date_local) VALUES($1, $2, $3, $4, $5, $6)`

	_, err := pgxConn.Pool.Exec(context.Background(), q, r.StravaId, r.AthleteId, r.Name, r.NameSlug, r.Distance, r.StartDate)
	if err != nil {
		return err
	}

	return nil
}

// SelectRaceActivityById selects and returns a single race_activity record by strava_id
func SelectRaceActivityById(pgxConn *PgxConn, stravaId uint64) (RaceActivity, error) {
	q := `SELECT strava_id, strava_athlete_id, name, name_slug, distance, start_date_local FROM race_activity WHERE strava_id=$1`
	var res RaceActivity

	err := pgxConn.Pool.
		QueryRow(context.Background(), q, stravaId).
		Scan(&res.StravaId, &res.AthleteId, &res.Name, &res.NameSlug, &res.Distance, &res.StartDate)
	if err != nil {
		return res, nil
	}

	return res, nil
}

func SelectAllRaces(pgxConn *PgxConn) ([]RaceActivity, error) {
	q := `SELECT strava_id, strava_athlete_id, name, name_slug, distance, start_date_local FROM race_activity ORDER BY start_date_local DESC`
	var res []RaceActivity

	rows, err := pgxConn.Pool.Query(context.Background(), q)
	if err != nil {
		return res, fmt.Errorf("Failed to execute SelectAllRaces query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r RaceActivity
		err := rows.Scan(&r.StravaId, &r.AthleteId, &r.Name, &r.NameSlug, &r.Distance, &r.StartDate)
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

func SelectRaceByYearAndSlug(pgxConn *PgxConn, year int, slug string) (RaceActivity, error) {
	q := `SELECT strava_id, strava_athlete_id, name, name_slug, distance, start_date_local
    FROM race_activity
    WHERE name_slug=$1 and EXTRACT(year FROM "start_date_local")=$2`
	var res RaceActivity

	err := pgxConn.Pool.
		QueryRow(context.Background(), q, slug, year).
		Scan(&res.StravaId, &res.AthleteId, &res.Name, &res.NameSlug, &res.Distance, &res.StartDate)
	if err != nil {
		return res, fmt.Errorf("Failed to select race activity by year and slug: %s", err)
	}

	return res, nil
}
