USE hkdb;

CREATE TABLE blog_sections (
  id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
  blogId INT NOT NUll,
  sectionType VARCHAR(255) NULL,
  content MEDIUMTEXT NULL,
  sequence int
);

CREATE INDEX blog_sections_sequence ON blog_sections(blogId, sequence);

ALTER TABLE blogs ADD COLUMN contentRaw MEDIUMTEXT NULL;
UPDATE blogs SET contentRaw = content;