BEGIN;

CREATE TABLE lsif_last_retention_scan (
    repository_id int NOT NULL,
    last_retention_scan_at timestamp with time zone NOT NULL,

    PRIMARY KEY(repository_id)
);

COMMIT;
