CREATE TABLE photos (
  id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
  path VARCHAR(255) NOT NULL,
  on_disk INT NOT NULL
);

CREATE INDEX photos_index_id ON photos(id);
CREATE INDEX photos_index_path ON photos(path);

CREATE TABLE blogs_photos (
  id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
  blog_id INT NOT NULL,
  path VARCHAR(255) NOT NULL
);

CREATE INDEX blogs_photos_index_id ON blogs_photos(id);
CREATE INDEX blogs_photos_index_blog_id ON blogs_photos(blog_id);
/*
CREATE INDEX blogs_photos_index_path ON blogs_photos(path);
*/
