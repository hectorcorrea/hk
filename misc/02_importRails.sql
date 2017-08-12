# Imports data from the temp rails database
# and formats it to the new table structure
USE hkdb;

INSERT INTO blogs (id, title, slug, content, thumbnail,
  blogDate, year, createdOn, updatedOn)
SELECT id, name, url, content, thumbnail,
  date,  year, created_at, updated_at
FROM hkrails.blogs;

update blogs set postedOn = createdOn;
