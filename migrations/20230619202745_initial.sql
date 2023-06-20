-- +goose Up
-- +goose StatementBegin
CREATE TABLE strava_auth (
    athlete_id integer PRIMARY KEY,
    access_token varchar(255) NOT NULL,
    access_token_expires_at integer NOT NUll,
    refresh_token varchar(255) NOT NULL
);

CREATE TABLE race_activity (
    strava_id integer NOT NULL,
    strava_athlete_id integer NOT NULL,
    name varchar(100) NOT NULL,
    name_slug varchar(150) PRIMARY KEY,
    distance float DEFAULT 0.0,
    race_type varchar(25) NOT NULL,
    race_date timestamp with time zone NOT NULL
);
CREATE INDEX strava_activity_id_idx ON race_activity (strava_id);

CREATE TABLE athlete (
    strava_id integer PRIMARY KEY,
    first_name varchar(25) NOT NULL,
    last_name varchar(25) NOT NULL,
    profile varchar(255),
    profile_medium varchar(255)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE strava_auth;
DROP TABLE race_activity;
DROP TABLE athlete;
-- +goose StatementEnd
