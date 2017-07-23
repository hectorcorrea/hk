package viewModels

import (
	"hk/models"
	"html/template"
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
	Year 			int
	IsNewYear bool
	IsDraft   bool
	Html      template.HTML
	Markdown  string
	Session
}

type BlogList struct {
	Blogs []Blog
	Session
}

func FromBlog(blog models.Blog, session Session) Blog {
	var vm Blog
	vm.Id = blog.Id
	vm.Title = blog.Title
	vm.Summary = blog.Summary
	vm.Slug = blog.Slug
	vm.Url = blog.URL("")
	if strings.Contains(strings.ToLower(blog.Thumbnail), "_thumb.jpg") {
		vm.Thumbnail = blog.Thumbnail
	}
	vm.Html = template.HTML(blog.ContentHtml)
	vm.Markdown = blog.ContentMarkdown
	vm.CreatedOn = blog.CreatedOn
	vm.PostedOn = blog.PostedOn
	vm.UpdatedOn = blog.UpdatedOn
	vm.Year = blog.Year
	vm.IsDraft = (vm.PostedOn == "")
	vm.IsNewYear = false
	vm.Session = session
	return vm
}

func FromBlogs(blogs []models.Blog, session Session) BlogList {
	var list []Blog
	lastYear := time.Now().Year()
	for _, blog := range blogs {
		vm := FromBlog(blog, session)
		if blog.Year != lastYear {
			vm.IsNewYear = true
			lastYear = int(blog.Year)
		}
		list = append(list, vm)
	}
	return BlogList{Blogs: list, Session: session}
}
