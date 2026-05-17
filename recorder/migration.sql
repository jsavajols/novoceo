-- À exécuter une seule fois en tant que superuser PostgreSQL.
-- psql "postgres://<superuser>:<pass>@51.159.109.144:30015/novoceo" -f recorder/migration.sql

CREATE TABLE IF NOT EXISTS states (
    id           SERIAL PRIMARY KEY,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    topic        TEXT        NOT NULL,
    device_state JSONB       NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_states_topic      ON states (topic);
CREATE INDEX IF NOT EXISTS idx_states_created_at ON states (created_at DESC);

-- Droits d'insertion pour l'utilisateur applicatif.
GRANT INSERT, SELECT ON TABLE states TO "rpi-novoceo";
GRANT USAGE, SELECT ON SEQUENCE states_id_seq TO "rpi-novoceo";
