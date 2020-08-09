package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
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
	ShareAlias  string
	ContentHtml string
	CreatedOn   string
	UpdatedOn   string
	PostedOn    string
	Photos      []string
	Sections    []BlogSection
}

type BlogSection struct {
	Id       int64
	Sequence int
	Content  string
	Type     string
	Saved    bool
}

func (s BlogSection) IsHeading() bool {
	return s.Type == "h"
}

func (s BlogSection) IsParagraph() bool {
	return s.Type == "p"
}

func (s BlogSection) IsParagraphEN() bool {
	return s.Type == "p-en"
}

func (s BlogSection) IsParagraphES() bool {
	return s.Type == "p-es"
}
func (s BlogSection) IsPhoto() bool {
	return s.Type == "i"
}

func (s BlogSection) Lines() []string {
	lines := []string{}
	for _, line := range strings.Split(s.Content, "\r") {
		image := strings.Trim(line, " \r\n")
		if image != "" {
			lines = append(lines, image)
		}
	}
	return lines
}

var photoPath = "UNSET"

func (b Blog) DebugString() string {
	str := fmt.Sprintf("Id: %d\nTitle: %s\nSummary: %s\n",
		b.Id, b.Title, b.Summary)
	return str
}

func (b Blog) SectionsNextSeq() int {
	maxSeq := 0
	for _, section := range b.Sections {
		if section.Sequence >= maxSeq {
			maxSeq = section.Sequence
		}
	}
	return maxSeq + 1
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

func BlogGetByAlias(alias string) (Blog, error) {
	id, err := getIdByAlias(alias)
	if err != nil {
		return Blog{}, err
	}
	return getOne(id)
}

func (b *Blog) beforeSave() error {
	b.Slug = getSlug(b.Title)
	b.UpdatedOn = dbUtcNow()
	b.Thumbnail = strings.Replace(b.Thumbnail, "http://", "https://", -1)

	html := b.sectionsAsHtml()
	if html != "" {
		log.Printf("Using sections rather than raw HTML for blog %d, %s", b.Id, b.Slug)
		b.ContentHtml = html
	}

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

// AddSection adds a section to the blog
func (b *Blog) AddSection(id int64, sectionType string, content string, seq int) {
	section := BlogSection{}
	section.Id = id
	section.Type = sectionType
	section.Content = content
	section.Sequence = seq
	b.Sections = append(b.Sections, section)
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

	shareAlias := sql.NullString{String: "", Valid: false}
	if b.ShareAlias != "" {
		shareAlias = sql.NullString{String: b.ShareAlias, Valid: true}
	}

	if b.ContentHtml != "" {
		// Update all fields, including contentHtml
		sqlUpdate := `
			UPDATE blogs
			SET title = ?, slug = ?, content = ?,
				blogDate = ?, year = ?, updatedOn = ?,
				thumbnail = ?, shareAlias = ?
			WHERE id = ?`
		_, err = db.Exec(sqlUpdate, b.Title, b.Slug, b.ContentHtml,
			b.BlogDate, b.Year, dbUtcNow(), b.Thumbnail, shareAlias, b.Id)
	} else {
		// No ContentHtml received, don't update the content field
		// (this is so that we don't overwrite old entries
		// accidentally while testing the new editor)
		sqlUpdate := `
			UPDATE blogs
			SET title = ?, slug = ?,
				blogDate = ?, year = ?, updatedOn = ?,
				thumbnail = ?, shareAlias = ?
			WHERE id = ?`
		_, err = db.Exec(sqlUpdate, b.Title, b.Slug,
			b.BlogDate, b.Year, dbUtcNow(), b.Thumbnail, shareAlias, b.Id)
	}
	if err != nil {
		return err
	}

	// Update the blog sections (insert, update, delete)
	for _, section := range b.Sections {
		if section.Id != 0 {
			if section.Content == "" {
				sqlDelete := `DELETE FROM blog_sections WHERE id = ?`
				_, err = db.Exec(sqlDelete, section.Id)
			} else {
				sqlUpdate := `UPDATE blog_sections 
				SET sectionType = ?, content = ?, sequence = ? 
				WHERE id = ?`
				_, err = db.Exec(sqlUpdate, section.Type, section.Content, section.Sequence, section.Id)
			}
		} else {
			if section.Content == "" {
				// ignore it
			} else {
				sqlInsert := `INSERT INTO blog_sections(blogId, sectionType, content, sequence)
				VALUES(?, ?, ?, ?)`
				_, err = db.Exec(sqlInsert, b.Id, section.Type, section.Content, section.Sequence)
			}
		}
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
		SELECT title, slug, blogDate, year, content, thumbnail, shareAlias,
			createdOn, updatedOn, postedOn
		FROM blogs
		WHERE id = ?`
	row := db.QueryRow(sqlSelect, id)

	var year sql.NullInt64
	var title, slug, content, thumbnail, shareAlias sql.NullString
	var blogDate, createdOn, updatedOn, postedOn mysql.NullTime
	err = row.Scan(&title, &slug, &blogDate, &year, &content, &thumbnail, &shareAlias,
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
	blog.ShareAlias = stringValue(shareAlias)
	blog.CreatedOn = timeValue(createdOn)
	blog.UpdatedOn = timeValue(updatedOn)
	blog.PostedOn = timeValue(postedOn)
	blog.ContentHtml = stringValue(content)
	blog.Sections, err = blog.getSections()
	return blog, err
}

func (b Blog) sectionsAsHtml() string {
	html := ""

	sorted := b.Sections
	sort.Slice(sorted, func(i, j int) bool {
		return b.Sections[i].Sequence < b.Sections[j].Sequence
	})

	for _, section := range sorted {
		content := strings.Trim(section.Content, " \r\n")
		if content == "" {
			continue
		}

		if section.Type == "h" {
			html += "<h2>" + section.Content + "</h2>"
		} else if section.Type == "i" {
			for _, line := range strings.Split(section.Content, "\r") {
				image := strings.Trim(line, " \r\n")
				if image != "" {
					html += fmt.Sprintf("<img alt=\"\" src=\"%s\" />", image)
				}
			}
		} else if section.Type == "p-en" {
			html += "<p class=\"en\">" + section.Content + "</p>"
		} else if section.Type == "p-es" {
			html += "<p class=\"es\">" + section.Content + "</p>"
		} else {
			html += "<p>" + section.Content + "</p>"
		}
	}
	return html
}

func (b *Blog) getSections() ([]BlogSection, error) {
	db, err := connectDB()
	if err != nil {
		return []BlogSection{}, err
	}
	defer db.Close()

	sqlSelect := `
		SELECT id, sectionType, content, sequence
		FROM blog_sections
		WHERE blogId = ?`
	rows, err := db.Query(sqlSelect, b.Id)
	if err != nil {
		return []BlogSection{}, err
	}
	defer rows.Close()

	var sections []BlogSection
	var id int64
	var sequence sql.NullInt64
	var content, sectionType sql.NullString
	for rows.Next() {
		err := rows.Scan(&id, &sectionType, &content, &sequence)
		if err != nil {
			return []BlogSection{}, err
		}
		section := BlogSection{
			Id:       id,
			Type:     stringValue(sectionType),
			Content:  stringValue(content),
			Sequence: intValue(sequence),
			Saved:    true,
		}
		sections = append(sections, section)
	}
	return sections, nil
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

func getIdByAlias(alias string) (int64, error) {
	db, err := connectDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var id int64
	sqlSelect := "SELECT id FROM blogs WHERE shareAlias = ? LIMIT 1"
	row := db.QueryRow(sqlSelect, alias)
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
