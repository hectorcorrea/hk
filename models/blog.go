package models

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Blog struct {
	Id          int64
	Title       string
	Summary     string
	Slug        string
	BlogDate    string
	Year        int
	Thumbnail   string
	ContentHtml string
	CreatedOn   string
	UpdatedOn   string
	PostedOn    string
	Photos      []string
}

var photoPath = "UNSET"

func (b Blog) DebugString() string {
	str := fmt.Sprintf("Id: %d\nTitle: %s\nSummary: %s\n",
		b.Id, b.Title, b.Summary)
	return str
}

func (b Blog) URL(base string) string {
	return fmt.Sprintf("%s/%d/%s/%d", base, b.Year, b.Slug, b.Id)
}

// RFC 1123Z looks like "Mon, 02 Jan 2006 15:04:05 -0700"
// https://golang.org/pkg/time/
func (b Blog) PostedOnRFC1123Z() string {
	layout := "2006-01-02 15:04:05 -0700 MST"
	t, err := time.Parse(layout, b.PostedOn)
	if err != nil {
		return ""
	}
	return t.Format(time.RFC1123Z)
}

func BlogGetAll() ([]Blog, error) {
	blogs, err := getBlogsWhere("")
	return blogs, err
}

func BlogGetYear(year int) ([]Blog, error) {
	if year < 2000 || year > time.Now().Year() {
		return []Blog{}, errors.New(fmt.Sprintf("Invalid year received (%d)", year))
	}

	where := fmt.Sprintf("WHERE year = %d", year)
	blogs, err := getBlogsWhere(where)
	return blogs, err
}

func BlogGetRecent() ([]Blog, error) {
	where := fmt.Sprintf("WHERE year >= %d", time.Now().Year()-1)
	blogs, err := getBlogsWhere(where)
	return blogs, err
}

func BlogGetById(id int64) (Blog, error) {
	blog, err := getOne(id)
	return blog, err
}

func BlogGetBySlug(slug string) (Blog, error) {
	id, err := getIdBySlug(slug)
	if err != nil {
		return Blog{}, err
	}
	return getOne(id)
}

func (b *Blog) beforeSave() error {
	b.Slug = getSlug(b.Title)
	b.UpdatedOn = dbUtcNow()
	b.Thumbnail = strings.Replace(b.Thumbnail, "http://", "https://", -1)
	b.ContentHtml = strings.Replace(b.ContentHtml, "http://www.hectorykarla.com", "https://www.hectorykarla.com", -1)
	b.ContentHtml = strings.Replace(b.ContentHtml, "http://hectorykarla.com", "https://hectorykarla.com", -1)
	if b.BlogDate == "" {
		b.BlogDate = b.UpdatedOn
	}
	b.BlogDate = b.BlogDate[0:10]
	b.Year = yearFromDbDate(b.BlogDate)
	b.Photos = b.getPhotos()
	return nil
}

func (b *Blog) AddPhoto(photoName string) {
	b.ContentHtml += fmt.Sprintf("<img src=\"%s\" />", photoName)
}

func (b *Blog) getPhotos() []string {
	// Gets the SRC path from all IMG tags. For example, for
	// 	<img x=xxx src="http://some/path" y=yyy />
	// it returns ["http://some/path"]
	var photos []string
	reImg := regexp.MustCompile("<img(.*?)>")
	reSrc := regexp.MustCompile("src=\"(.*?)\"")
	imgTags := reImg.FindAllString(b.ContentHtml, -1)
	for _, img := range imgTags {
		src := reSrc.FindString(img)
		if src != "" {
			path := src[5 : len(src)-1]
			photos = append(photos, path)
		}
	}
	return photos
}

func yearFromDbDate(dbDate string) int {
	if len(dbDate) < 4 {
		return 0
	}
	id, _ := strconv.ParseInt(dbDate[0:4], 10, 64)
	return int(id)
}

func getSlug(title string) string {
	slug := strings.Trim(title, " ")
	slug = strings.ToLower(slug)
	slug = strings.Replace(slug, "c#", "c-sharp", -1)
	var chars []rune
	for _, c := range slug {
		isAlpha := c >= 'a' && c <= 'z'
		isDigit := c >= '0' && c <= '9'
		if isAlpha || isDigit {
			chars = append(chars, c)
		} else {
			chars = append(chars, '-')
		}
	}
	slug = string(chars)

	// remove double dashes
	for strings.Index(slug, "--") > -1 {
		slug = strings.Replace(slug, "--", "-", -1)
	}

	if len(slug) == 0 || slug == "-" {
		return ""
	}

	// make sure we don't end with a dash
	if slug[len(slug)-1] == '-' {
		return slug[0 : len(slug)-1]
	}

	return slug
}

func SaveNew() (int64, error) {
	db, err := connectDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	dbNow := dbUtcNow()
	sqlInsert := `
		INSERT INTO blogs(title, summary, slug, content, blogDate, year, createdOn)
		VALUES(?, ?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(sqlInsert, "new blog", "", "new-blog", "",
		dbNow, yearFromDbDate(dbNow), dbNow)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (b *Blog) Save() error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	b.beforeSave()

	sqlUpdate := `
		UPDATE blogs
		SET title = ?, slug = ?, content = ?,
			blogDate = ?, year = ?, updatedOn = ?,
			thumbnail = ?
		WHERE id = ?`
	_, err = db.Exec(sqlUpdate, b.Title, b.Slug, b.ContentHtml,
		b.BlogDate, b.Year, dbUtcNow(), b.Thumbnail, b.Id)
	if err != nil {
		return err
	}

	// delete previous photos attached to this blog
	sqlDelete := `DELETE FROM blogs_photos WHERE blog_id = ?`
	_, err = db.Exec(sqlDelete, b.Id)
	if err != nil {
		return err
	}

	// re-add photos
	for _, path := range b.Photos {
		sqlInsert := `INSERT INTO blogs_photos(blog_id, path) VALUES(?, ?)`
		_, err = db.Exec(sqlInsert, b.Id, path)
		if err != nil {
			return err
		}
	}

	return nil
}

func getOne(id int64) (Blog, error) {
	db, err := connectDB()
	if err != nil {
		return Blog{}, err
	}
	defer db.Close()

	sqlSelect := `
		SELECT title, slug, blogDate, year, content, thumbnail,
			createdOn, updatedOn, postedOn
		FROM blogs
		WHERE id = ?`
	row := db.QueryRow(sqlSelect, id)

	var year sql.NullInt64
	var title, slug, content, thumbnail sql.NullString
	var blogDate, createdOn, updatedOn, postedOn mysql.NullTime
	err = row.Scan(&title, &slug, &blogDate, &year, &content, &thumbnail,
		&createdOn, &updatedOn, &postedOn)
	if err != nil {
		return Blog{}, err
	}

	var blog Blog
	blog.Id = id
	blog.Title = stringValue(title)
	blog.Slug = stringValue(slug)
	blog.Year = intValue(year)
	blog.BlogDate = dateValue(blogDate)
	blog.Thumbnail = stringValue(thumbnail)
	blog.CreatedOn = timeValue(createdOn)
	blog.UpdatedOn = timeValue(updatedOn)
	blog.PostedOn = timeValue(postedOn)
	blog.ContentHtml = stringValue(content)
	return blog, nil
}

func PhotoPath() (string, error) {
	if photoPath != "UNSET" {
		return photoPath, nil
	}

	// Fetch if from the database...
	db, err := connectDB()
	if err != nil {
		// Bail out but don't cache a value for it, so it will be retried
		// in the next call to getPhotoPath()
		return "", err
	}
	defer db.Close()

	var path sql.NullString
	sqlSelect := "SELECT photoPath FROM settings LIMIT 1;"
	row := db.QueryRow(sqlSelect)
	err = row.Scan(&path)
	if err != nil {
		photoPath = "" // cache a value to prevent retries
		return photoPath, err
	}

	// ...and cache it
	photoPath = stringValue(path)
	return photoPath, nil
}

func MaskPhotoPaths(text string) string {
	path, err := PhotoPath()
	if err != nil {
		return text
	}
	return strings.Replace(text, "/photos/", "/"+path+"/", -1)
}

func getIdBySlug(slug string) (int64, error) {
	db, err := connectDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var id int64
	sqlSelect := "SELECT id FROM blogs WHERE slug = ? LIMIT 1"
	row := db.QueryRow(sqlSelect, slug)
	err = row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func getBlogsWhere(where string) ([]Blog, error) {
	db, err := connectDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sqlSelect := "SELECT id, title, summary, slug, year, postedOn, thumbnail FROM blogs "
	sqlSelect += where + " "
	sqlSelect += "ORDER BY blogDate DESC"
	rows, err := db.Query(sqlSelect)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []Blog
	var id int64
	var year sql.NullInt64
	var title, summary, slug, thumbnail sql.NullString
	var postedOn mysql.NullTime
	for rows.Next() {
		err := rows.Scan(&id, &title, &summary, &slug, &year, &postedOn, &thumbnail)
		if err != nil {
			return nil, err
		}
		blog := Blog{
			Id:        id,
			Title:     stringValue(title),
			Summary:   stringValue(summary),
			Slug:      stringValue(slug),
			Thumbnail: stringValue(thumbnail),
			Year:      intValue(year),
			PostedOn:  timeValue(postedOn),
		}
		blogs = append(blogs, blog)
	}
	return blogs, nil
}
