USE hkdb;

ALTER TABLE blogs ADD COLUMN shareAlias VARCHAR(255) NULL;
CREATE INDEX blogs_index_shareAlias ON blogs(shareAlias);

