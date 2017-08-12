package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"hk/models"
	"hk/viewModels"
)

var blogRouter Router

func init() {
	blogRouter.Add("GET", "/blogs/:title_id", blogViewOneLegacy)
	blogRouter.Add("GET", "/:year/:title/:id", blogViewOne)
	blogRouter.Add("GET", "/archive", blogViewAll)
	blogRouter.Add("GET", "/", blogViewRecent)
	blogRouter.Add("POST", "/:year/:title/:id/edit", blogEdit)
	blogRouter.Add("POST", "/:year/:title/:id/save", blogSave)
	blogRouter.Add("POST", "/new", blogNew)
}

func blogPages(resp http.ResponseWriter, req *http.Request) {
	// blogRouter.PrintRoutes()
	session := newSession(resp, req)
	found, route := blogRouter.FindRoute(req.Method, req.URL.Path)
	if found {
		values := route.UrlValues(req.URL.Path)
		route.handler(session, values)
	} else {
		renderNotFound(session)
	}
}

func blogViewOne(s session, values map[string]string) {
	log.Print("blogViewOne")

	id := idFromString(values["id"])
	// log.Println(values)
	if id == 0 {
		log.Printf("blogViewOne id %s", "no id")
		renderError(s, "No Blog ID was received", nil)
		return
	}
	log.Printf("blogViewOne id %d", id)

	log.Printf("Loading %d", id)
	blog, err := models.BlogGetById(id)
	if err != nil {
		renderError(s, "Fetching by ID", err)
		return
	}

	year := values["year"]
	slug := values["title"]
	if (year != strconv.Itoa(blog.Year)) || (slug != blog.Slug) {
		newUrl := fmt.Sprintf("/%d/%s/%d", blog.Year, blog.Slug, blog.Id)
		log.Printf("Redirected to %s", newUrl)
		http.Redirect(s.resp, s.req, newUrl, http.StatusMovedPermanently)
		return
	}

	log.Print("blogViewOne")
	vm := viewModels.FromBlog(blog, s.toViewModel(), false)
	renderTemplate(s, "views/blogView.html", vm)
}

func blogViewOneLegacy(s session, values map[string]string) {
	log.Print("blogViewOneLegacy")

	id := idFromLegacyUrl(values["title_id"])
	if id == 0 {
		log.Printf("Legacy post without an ID. Redirected to home page.")
		http.Redirect(s.resp, s.req, "/", http.StatusMovedPermanently)
		return
	}

	blog, err := models.BlogGetById(id)
	if err != nil {
		log.Printf("Legacy post %d not found. Redirected to home page.", id)
		http.Redirect(s.resp, s.req, "/", http.StatusMovedPermanently)
		return
	}

	newUrl := fmt.Sprintf("/%d/%s/%d", blog.Year, blog.Slug, blog.Id)
	log.Printf("Legacy %d redirected to %s", id, newUrl)
	http.Redirect(s.resp, s.req, newUrl, http.StatusMovedPermanently)
}

func blogViewRecent(s session, values map[string]string) {
	showDrafts := s.isAuth()
	log.Printf("Loading recent...")
	if blogs, err := models.BlogGetRecent(showDrafts); err != nil {
		renderError(s, "Error fetching recent", err)
	} else {
		vm := viewModels.FromBlogs(blogs, s.toViewModel(), true)
		renderTemplate(s, "views/blogList.html", vm)
	}
}

func blogViewAll(s session, values map[string]string) {
	showDrafts := s.isAuth()
	log.Printf("Loading all...")
	if blogs, err := models.BlogGetAll(showDrafts); err != nil {
		renderError(s, "Error fetching all", err)
	} else {
		vm := viewModels.FromBlogs(blogs, s.toViewModel(), false)
		renderTemplate(s, "views/blogList.html", vm)
	}
}

func blogSave(s session, values map[string]string) {
	if !s.isAuth() {
		renderNotAuthorized(s)
		return
	}

	id := idFromString(values["id"])
	blog := blogFromForm(id, s)
	if err := blog.Save(); err != nil {
		renderError(s, fmt.Sprintf("Saving blog ID: %d"), err)
	} else {
		url := fmt.Sprintf("/%d/%s/%d", blog.Year, blog.Slug, id)
		log.Printf("Redirect to %s", url)
		http.Redirect(s.resp, s.req, url, 301)
	}
}

func blogNew(s session, values map[string]string) {
	if !s.isAuth() {
		renderNotAuthorized(s)
		return
	}
	newId, err := models.SaveNew()
	if err != nil {
		renderError(s, fmt.Sprintf("Error creating new blog"), err)
		return
	}
	log.Printf("Redirect to (edit for new) %d", newId)
	values["id"] = fmt.Sprintf("%d", newId)
	blogEdit(s, values)
}

func blogEdit(s session, values map[string]string) {
	if !s.isAuth() {
		renderNotAuthorized(s)
		return
	}
	id := idFromString(values["id"])
	if id == 0 {
		renderError(s, "No blog ID was received", nil)
		return
	}

	log.Printf("Loading %d", id)
	blog, err := models.BlogGetById(id)
	if err != nil {
		renderError(s, fmt.Sprintf("Loading ID: %d", id), err)
		return
	}

	vm := viewModels.FromBlog(blog, s.toViewModel(), true)
	renderTemplate(s, "views/blogEdit.html", vm)
}

func idFromString(str string) int64 {
	id, _ := strconv.ParseInt(str, 10, 64)
	return id
}

func idFromLegacyUrl(url string) int64 {
	// url is expected as 2017-something-something-123
	// where 123 is the ID.
	index := strings.LastIndex(url, "-")
	if index == -1 {
		return 0
	}

	idString := url[index+1 : len(url)]
	return idFromString(idString)
}

func blogFromForm(id int64, s session) models.Blog {
	var blog models.Blog
	blog.Id = id
	blog.Title = s.req.FormValue("title")
	blog.Summary = s.req.FormValue("summary")
	blog.ContentHtml = s.req.FormValue("content")
	blog.Thumbnail = s.req.FormValue("thumbnail")
	blog.BlogDate = s.req.FormValue("blogdate")
	return blog
}
