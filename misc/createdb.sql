CREATE DATABASE hkdb;

USE hkdb;

CREATE TABLE blogs (
  id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  summary VARCHAR(255) NULL,
  slug VARCHAR(255) NULL,
  content MEDIUMTEXT NULL,
  thumbnail varchar(255),
  blogDate date,
  year int,
  createdOn DATETIME NOT NULL,
  updatedOn DATETIME NULL,
  postedOn DATETIME NULL
);

CREATE INDEX blogs_index_title ON blogs(title);
CREATE INDEX blogs_index_blogDaten ON blogs(blogDate DESC);


CREATE TABLE users (
  id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
  login VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL
);


CREATE INDEX users_index_id ON users(id);


CREATE TABLE sessions (
  id char(64) NOT NULL PRIMARY KEY,
  userId INT NOT NULL,
  expiresOn DATETIME NOT NULL
);

CREATE INDEX sessions_index_id ON sessions(id);
