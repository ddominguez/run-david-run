-- +goose Up
-- +goose StatementBegin
CREATE TABLE strava_auth (
    athlete_id INTEGER PRIMARY KEY,
    access_token TEXT NOT NULL,
    access_token_expires_at INTEGER NOT NULL,
    refresh_token TEXT NOT NULL
);

CREATE TABLE race_activity (
    strava_id INTEGER NOT NULL PRIMARY KEY,
    strava_athlete_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    distance REAL DEFAULT 0.0,
    moving_time INTEGER DEFAULT 0,
    elapsed_time INTEGER DEFAULT 0,
    start_date_local TEXT NOT NULL,
    polyline TEXT NOT NULL
);

CREATE TABLE athlete (
    strava_id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    profile TEXT,
    profile_medium TEXT,
    latest_activity_datetime TEXT DEFAULT ''
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE strava_auth;
DROP TABLE race_activity;
DROP TABLE athlete;
-- +goose StatementEnd
