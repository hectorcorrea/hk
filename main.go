package main

import (
	"flag"
	"hk/models"
	"hk/web"
	"log"
)

func main() {
	var address = flag.String("address", "localhost:9001", "Address where server will listen for connections")
	var resave = flag.String("resave", "", "Pass \"yes\" to resave all blog posts and recalculate the HTML content.")
	flag.Parse()

	if *resave == "yes" {
		resaveAll()
		return
	}

	web.StartWebServer(*address)
}

func resaveAll() {
	if err := models.InitDB(); err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	log.Printf("Database: %s", models.DbConnStringSafe())
	blogs, _ := models.BlogGetAll(true)
	for _, b := range blogs {
		log.Printf("re-saving %d - %s", b.Id, b.Title)
		blog, _ := models.BlogGetById(b.Id)
		blog.Save()
	}
}
