package viewModels

import (
	"fmt"
	"hk/models"
	"html/template"
	"regexp"
	"strings"
	"time"
)

type Blog struct {
	Id        int64
	Title     string
	Summary   string
	Slug      string
	Url       string
	CreatedOn string
	PostedOn  string
	UpdatedOn string
	Thumbnail string
	BlogDate  string
	Year      int
	IsNewYear bool
	IsDraft   bool
	Html      template.HTML
	Message   string
	Session
}

type BlogRow struct {
	Blogs []Blog
}

type BlogList struct {
	BlogMatrix []BlogRow
	Title      string
	Session
	ShowMoreUrl bool
	MoreUrl     string
}

func FromBlog(blog models.Blog, session Session, raw bool) Blog {
	var vm Blog
	vm.Id = blog.Id
	vm.Title = blog.Title
	vm.Summary = blog.Summary
	vm.Slug = blog.Slug
	vm.Url = blog.URL("")

	if strings.Contains(strings.ToLower(blog.Thumbnail), "_thumb.jpg") {
		vm.Thumbnail = blog.Thumbnail
	} else {
		vm.Message = fmt.Sprintf("Thumbnail ignored because it does not end in _thumb.jpg (%s)", blog.Thumbnail)
	}

	if raw {
		// For editing we show the data as-is (no masking)
		vm.Html = template.HTML(blog.ContentHtml)
	} else {
		// For viewing we mask the path to the photos
		vm.Thumbnail = models.MaskPhotoPaths(blog.Thumbnail)
		html := models.MaskPhotoPaths(blog.ContentHtml)
		vm.Html = template.HTML(addGalleryTags(html))
	}

	vm.CreatedOn = blog.CreatedOn
	vm.PostedOn = blog.PostedOn
	vm.UpdatedOn = blog.UpdatedOn
	vm.Year = blog.Year
	vm.BlogDate = blog.BlogDate
	vm.IsDraft = (vm.PostedOn == "")
	vm.IsNewYear = false
	vm.Session = session
	return vm
}

func FromBlogs(blogs []models.Blog, session Session, showMore bool) BlogList {
	matrix := []BlogRow{}
	r := -1
	c := -1
	lastYear := -1
	for _, blog := range blogs {
		vm := FromBlog(blog, session, false)
		if blog.Year != lastYear {
			vm.IsNewYear = true
			lastYear = int(blog.Year)
		}
		if vm.IsNewYear || c == 3 {
			// a new row is needed
			matrix = append(matrix, BlogRow{})
			r += 1
			c = 0
		}
		matrix[r].Blogs = append(matrix[r].Blogs, vm)
		c += 1
	}

	blogList := BlogList{
		BlogMatrix:  matrix,
		Session:     session,
		ShowMoreUrl: showMore,
		Title:       "",
		MoreUrl:     fmt.Sprintf("/archive/%d", time.Now().Year()-2),
	}
	return blogList
}

func addGalleryTags(html string) string {
	reImg := regexp.MustCompile("<img(.*?) src=\"(.*?)\"(.*?)/>")
	imgTags := reImg.FindAllString(html, -1)
	for _, img := range imgTags {
		html = strings.Replace(html, img, wrapImgTag(img), 1)
	}
	return html
}

// Converts
//		<img src="somefile_thumb.jpg" alt="y" />
// into
// 		<a href="somefile.jpg">
//			<img="somefile_thumb.jpg" alt="y" />
// 		</a>
func wrapImgTag(img string) string {
	reSrc := regexp.MustCompile("src=\"(.*?)\"")
	reAlt := regexp.MustCompile("alt=\"(.*?)\"")
	src := reSrc.FindString(img) // src="hello.jpg"
	srcValue := TextInQuotes(src)
	alt := reAlt.FindString(img) // alt="hello"
	file := strings.Replace(srcValue, "_thumb.jpg", ".jpg", 1)
	newImg := "<img src=\"" + srcValue + "\" " + alt + " />"

	newTag := "<a href=\"" + file + "\" >\r\n"
	newTag += "  " + newImg + "\r\n"
	newTag += "</a>\r\n"
	return newTag
}

func TextInQuotes(text string) string {
	q1 := strings.Index(text, "\"") + 1
	q2 := strings.LastIndex(text, "\"")
	if q2 < q1 {
		return ""
	}
	return text[q1:q2]
}
