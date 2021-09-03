BEGIN;

CREATE INDEX lsif_nearest_uploads_uploads ON lsif_nearest_uploads USING GIN(uploads);
CREATE INDEX lsif_nearest_uploads_links_repository_id_ancestor_commit_bytea ON lsif_nearest_uploads_links(repository_id, ancestor_commit_bytea);

COMMIT;
