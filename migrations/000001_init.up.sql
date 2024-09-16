CREATE TABLE short_urls (
    id SERIAL PRIMARY KEY,
    unique_id VARCHAR(255) UNIQUE NOT NULL,
    short_url VARCHAR(255) UNIQUE NOT NULL,
    long_url TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS users(
   id serial PRIMARY KEY,
   nickname VARCHAR (50) UNIQUE NOT NULL,
   password_hash VARCHAR (100) NOT NULL,
   refresh_token VARCHAR (100),
   expires_at TIMESTAMP
);

CREATE INDEX url_idx on short_urls (short_url);