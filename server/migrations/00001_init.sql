-- +goose Up
CREATE TABLE scans (
    id uuid PRIMARY KEY,
    url text NOT NULL,
    depth integer NOT NULL DEFAULT 1,
    status text NOT NULL DEFAULT 'queued',
    score integer,
    stats jsonb NOT NULL DEFAULT '{"critical":0,"warning":0,"ok":0}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz
);

CREATE TABLE pages (
    id uuid PRIMARY KEY,
    scan_id uuid NOT NULL REFERENCES scans (id) ON DELETE CASCADE,
    url text NOT NULL,
    status_code integer,
    title text,
    score integer,
    stats jsonb NOT NULL DEFAULT '{"critical":0,"warning":0,"ok":0}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX pages_scan_id_idx ON pages (scan_id);

CREATE TABLE findings (
    id uuid PRIMARY KEY,
    scan_id uuid NOT NULL REFERENCES scans (id) ON DELETE CASCADE,
    page_id uuid REFERENCES pages (id) ON DELETE CASCADE,
    kind text NOT NULL,
    severity text NOT NULL,
    code text NOT NULL,
    title text NOT NULL,
    element text NOT NULL DEFAULT '',
    recommendation text NOT NULL DEFAULT '',
    wcag text NOT NULL DEFAULT '',
    confidence real,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX findings_scan_id_idx ON findings (scan_id);
CREATE INDEX findings_page_id_idx ON findings (page_id);
CREATE INDEX findings_severity_idx ON findings (severity);

-- +goose Down
DROP TABLE findings;
DROP TABLE pages;
DROP TABLE scans;
