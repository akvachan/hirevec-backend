CREATE SCHEMA IF NOT EXISTS v1;

-- ULID base generator
CREATE OR REPLACE FUNCTION generate_ulid() RETURNS text AS $$
DECLARE
  encoding   bytea = '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
  timestamp  bytea = E'\\000\\000\\000\\000\\000\\000';
  output     text  = '';
  unix_time  bigint;
  ulid       bytea;
BEGIN
  unix_time = (EXTRACT(EPOCH FROM clock_timestamp()) * 1000)::bigint;

  timestamp = SET_BYTE(timestamp, 0, ((unix_time >> 40) & 255)::int);
  timestamp = SET_BYTE(timestamp, 1, ((unix_time >> 32) & 255)::int);
  timestamp = SET_BYTE(timestamp, 2, ((unix_time >> 24) & 255)::int);
  timestamp = SET_BYTE(timestamp, 3, ((unix_time >> 16) & 255)::int);
  timestamp = SET_BYTE(timestamp, 4, ((unix_time >> 8)  & 255)::int);
  timestamp = SET_BYTE(timestamp, 5, (unix_time         & 255)::int);


  -- uuid_send converts UUID to raw 16 bytes; take 10 for the random component
  ulid = timestamp || substring(uuid_send(gen_random_uuid()) FROM 1 FOR 10);

  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 0) & 224) >> 5));
  output = output || CHR(GET_BYTE(encoding,  GET_BYTE(ulid, 0) & 31));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 1) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 1) & 7) << 2) | ((GET_BYTE(ulid, 2) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 2) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 2) & 1) << 4) | ((GET_BYTE(ulid, 3) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 3) & 15) << 1) | ((GET_BYTE(ulid, 4) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 4) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 4) & 3) << 3) | ((GET_BYTE(ulid, 5) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding,  GET_BYTE(ulid, 5) & 31));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 6) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 6) & 7) << 2) | ((GET_BYTE(ulid, 7) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 7) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 7) & 1) << 4) | ((GET_BYTE(ulid, 8) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 8) & 15) << 1) | ((GET_BYTE(ulid, 9) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 9) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 9) & 3) << 3) | ((GET_BYTE(ulid, 10) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding,  GET_BYTE(ulid, 10) & 31));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 11) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 11) & 7) << 2) | ((GET_BYTE(ulid, 12) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 12) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 12) & 1) << 4) | ((GET_BYTE(ulid, 13) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 13) & 15) << 1) | ((GET_BYTE(ulid, 14) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 14) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 14) & 3) << 3) | ((GET_BYTE(ulid, 15) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding,  GET_BYTE(ulid, 15) & 31));

  RETURN output;
END
$$ LANGUAGE plpgsql VOLATILE;

-- Prefixed ULID helper
CREATE OR REPLACE FUNCTION generate_ulid(prefix text) RETURNS text AS $$
BEGIN
  RETURN prefix || '_' || generate_ulid();
END
$$ LANGUAGE plpgsql VOLATILE;

-- Per-table ULID generators
CREATE OR REPLACE FUNCTION generate_ulid_usr() RETURNS text AS $$ BEGIN RETURN generate_ulid('usr'); END $$ LANGUAGE plpgsql VOLATILE;
CREATE OR REPLACE FUNCTION generate_ulid_rtk() RETURNS text AS $$ BEGIN RETURN generate_ulid('rtk'); END $$ LANGUAGE plpgsql VOLATILE;
CREATE OR REPLACE FUNCTION generate_ulid_can() RETURNS text AS $$ BEGIN RETURN generate_ulid('can'); END $$ LANGUAGE plpgsql VOLATILE;
CREATE OR REPLACE FUNCTION generate_ulid_rec() RETURNS text AS $$ BEGIN RETURN generate_ulid('rec'); END $$ LANGUAGE plpgsql VOLATILE;
CREATE OR REPLACE FUNCTION generate_ulid_pos() RETURNS text AS $$ BEGIN RETURN generate_ulid('pos'); END $$ LANGUAGE plpgsql VOLATILE;

-- USERS
CREATE TABLE IF NOT EXISTS v1.users (
    id               TEXT PRIMARY KEY DEFAULT generate_ulid_usr(),
    provider         VARCHAR(50)  NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    email            VARCHAR(255),
    full_name        VARCHAR(255),
    user_name        VARCHAR(100) UNIQUE,
    updated_at       TIMESTAMP DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

-- REFRESH TOKENS
CREATE TABLE IF NOT EXISTS v1.refresh_tokens (
    jti        TEXT PRIMARY KEY DEFAULT generate_ulid_rtk(),
    user_id    TEXT      NOT NULL REFERENCES v1.users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    revoked    BOOLEAN   DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id
    ON v1.refresh_tokens(user_id);

-- CANDIDATES
CREATE TABLE IF NOT EXISTS v1.candidates (
    id      TEXT PRIMARY KEY DEFAULT generate_ulid_can(),
    user_id TEXT NOT NULL REFERENCES v1.users(id) ON DELETE CASCADE,
    about   TEXT NOT NULL
);

-- RECRUITERS
CREATE TABLE IF NOT EXISTS v1.recruiters (
    id      TEXT PRIMARY KEY DEFAULT generate_ulid_rec(),
    user_id TEXT NOT NULL REFERENCES v1.users(id) ON DELETE CASCADE
);

-- POSITIONS
CREATE TABLE IF NOT EXISTS v1.positions (
    id          TEXT PRIMARY KEY DEFAULT generate_ulid_pos(),
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    company     TEXT
);

-- REACTIONS
DO $$ BEGIN
    CREATE TYPE v1.reaction_type AS ENUM ('positive', 'negative', 'neutral');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS v1.candidates_reactions (
    candidate_id  TEXT NOT NULL REFERENCES v1.candidates(id) ON DELETE CASCADE,
    position_id   TEXT NOT NULL REFERENCES v1.positions(id)  ON DELETE CASCADE,
    reaction_type v1.reaction_type NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (candidate_id, position_id)
);

CREATE TABLE IF NOT EXISTS v1.recruiters_reactions (
    recruiter_id  TEXT NOT NULL REFERENCES v1.recruiters(id)  ON DELETE CASCADE,
    position_id   TEXT NOT NULL REFERENCES v1.positions(id)   ON DELETE CASCADE,
    candidate_id  TEXT NOT NULL REFERENCES v1.candidates(id)  ON DELETE CASCADE,
    reaction_type v1.reaction_type NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (recruiter_id, position_id, candidate_id)
);

-- MATCHES
CREATE TABLE IF NOT EXISTS v1.matches (
    candidate_id TEXT NOT NULL REFERENCES v1.candidates(id) ON DELETE CASCADE,
    position_id  TEXT NOT NULL REFERENCES v1.positions(id)  ON DELETE CASCADE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (candidate_id, position_id)
);

-- TEST DATA
INSERT INTO v1.users (provider, provider_user_id, email, full_name, user_name)
VALUES
    ('google', 'google-123', 'candidate@test.com', 'Jane Doe',   'jane_doe'),
    ('google', 'google-456', 'recruiter@test.com', 'John Smith', 'john_smith')
ON CONFLICT (provider, provider_user_id) DO NOTHING;

INSERT INTO v1.candidates (user_id, about)
SELECT id, 'Backend developer with 5 years of experience'
FROM v1.users
WHERE email = 'candidate@test.com'
ON CONFLICT DO NOTHING;

INSERT INTO v1.recruiters (user_id)
SELECT id
FROM v1.users
WHERE email = 'recruiter@test.com'
ON CONFLICT DO NOTHING;

INSERT INTO v1.positions (title, description, company)
VALUES
    ('Backend Engineer',    'Work on APIs and databases', 'Acme Inc'),
    ('Fullstack Developer', 'React + Node.js role',       'Tech Corp')
ON CONFLICT DO NOTHING;

INSERT INTO v1.candidates_reactions (candidate_id, position_id, reaction_type)
SELECT c.id, p.id, 'positive'
FROM v1.candidates c, v1.positions p
WHERE c.id = (SELECT id FROM v1.candidates LIMIT 1)
  AND p.id = (SELECT id FROM v1.positions  LIMIT 1)
ON CONFLICT DO NOTHING;

INSERT INTO v1.recruiters_reactions (recruiter_id, position_id, candidate_id, reaction_type)
SELECT r.id, p.id, c.id, 'positive'
FROM v1.recruiters r, v1.positions p, v1.candidates c
WHERE r.id = (SELECT id FROM v1.recruiters LIMIT 1)
  AND p.id = (SELECT id FROM v1.positions  LIMIT 1)
  AND c.id = (SELECT id FROM v1.candidates LIMIT 1)
ON CONFLICT DO NOTHING;

INSERT INTO v1.matches (candidate_id, position_id)
SELECT c.id, p.id
FROM v1.candidates c, v1.positions p
WHERE c.id = (SELECT id FROM v1.candidates LIMIT 1)
  AND p.id = (SELECT id FROM v1.positions  LIMIT 1)
ON CONFLICT DO NOTHING;

INSERT INTO v1.refresh_tokens (user_id, expires_at)
SELECT id, NOW() + INTERVAL '30 days'
FROM v1.users
WHERE email = 'candidate@test.com';
