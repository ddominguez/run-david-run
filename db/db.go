package db

import (
	"context"

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
func InsertStravaAuth(pgxConn *PgxConn, accessToken, refreshToken string, expiresAt, athleteId uint64) error {
	q := `INSERT INTO strava_auth(access_token, access_token_expires_at, refresh_token, athlete_id) VALUES ($1, $2, $3, $4)`
	_, err := pgxConn.Pool.Exec(context.Background(), q, accessToken, expiresAt, refreshToken, athleteId)
	if err != nil {
		return err
	}
	return nil
}

// StravaAthlete represent db table `athlete`
type StravaAthlete struct {
	Id            uint64
	FirstName     string
	LastName      string
	Profile       string
	ProfileMedium string
}

func (s *StravaAthlete) Exists() bool {
	return s.Id > 0
}

// SelectStravaAthleteById selects and returns a single strava athlete record
func SelectStravaAthleteById(pgxConn *PgxConn, athleteId uint64) (StravaAthlete, error) {
	q := `SELECT strava_id, first_name, last_name, profile, profile_medium WHERE strava_id=$1`
	var res StravaAthlete
	err := pgxConn.Pool.
		QueryRow(context.Background(), q, athleteId).
		Scan(&res.Id, &res.FirstName, &res.LastName, &res.Profile, &res.ProfileMedium)
	if err != nil {
		return res, nil
	}
	return res, nil
}

// InsertStravaAthelete inserts a new strava athlete record
func InsertStravaAthelete(pgxConn *PgxConn, id uint64, fname, lname, profile, profileMed string) error {
	q := `INSERT INTO athlete(strava_id, first_name, last_name, profile, profile_medium) VALUES($1, $2, $3, $4, $5)`
	_, err := pgxConn.Pool.Exec(context.Background(), q, id, fname, lname, profile, profileMed)
	if err != nil {
		return err
	}
	return nil
}
