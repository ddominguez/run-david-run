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

// StravaAuth represents db table strava_auth
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
