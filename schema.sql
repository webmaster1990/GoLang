DROP SCHEMA public CASCADE;
CREATE SCHEMA public;

CREATE TABLE organizations (
  organization_id   UUID PRIMARY KEY,
  organization_name VARCHAR,
  logo_url          VARCHAR,
  mission           TEXT,
  vision            TEXT,
  program           TEXT,
  objective         TEXT,
  ts_created        TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE users (
  user_id         UUID PRIMARY KEY,
  organization_id UUID REFERENCES organizations (organization_id),
  email           VARCHAR,
  password        VARCHAR,
  full_name       VARCHAR,
  is_admin        BOOL        DEFAULT FALSE,
  is_active       BOOL,
  ts_created      TIMESTAMPTZ DEFAULT now()
);

-- index for unique emails
CREATE UNIQUE INDEX ON users (lower(email));

CREATE TABLE projects (
  project_id      UUID PRIMARY KEY,
  organization_id UUID REFERENCES organizations (organization_id),
  project_name    VARCHAR,
  logo_url        VARCHAR,
  description     TEXT,
  budget          NUMERIC(15, 2),
  timeline_from   TIMESTAMPTZ,
  timeline_to     TIMESTAMPTZ,
  mission         TEXT,
  vision          TEXT,
  donor           VARCHAR,
  ts_created      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE external_resources (
  resource_id   UUID PRIMARY KEY,
  project_id    UUID REFERENCES projects (project_id) ON DELETE CASCADE,
  resource_url  VARCHAR,
  resource_name VARCHAR,
  created_by    UUID REFERENCES users (user_id),
  ts_created    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE boundary_partners (
  boundary_partner_id UUID PRIMARY KEY,
  project_id          UUID REFERENCES projects (project_id) ON DELETE CASCADE,
  partner_name        VARCHAR,
  outcome_statement   TEXT,
  ts_created          TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE progress_markers (
  progress_marker_id  UUID PRIMARY KEY,
  boundary_partner_id UUID REFERENCES boundary_partners (boundary_partner_id) ON DELETE CASCADE,
  title               VARCHAR,
  type                SMALLINT,
  order_number        SMALLINT,
  ts_created          TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE challenges (
  challenge_id       UUID PRIMARY KEY,
  progress_marker_id UUID REFERENCES progress_markers (progress_marker_id) ON DELETE CASCADE,
  challenge_name     VARCHAR,
  ts_created         TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE strategies (
  strategy_id        UUID PRIMARY KEY,
  progress_marker_id UUID REFERENCES progress_markers (progress_marker_id) ON DELETE CASCADE,
  strategy_name      VARCHAR,
  ts_created         TIMESTAMPTZ DEFAULT now()
);