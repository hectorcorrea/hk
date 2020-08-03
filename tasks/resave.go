package tasks

import (
	"log"
	"os"

	"hectorcorrea.com/hk/models"
)

// Re-saves all blogs. Used to populate the information as we
// update the site (e.g. add photos to DB or handling of legacy HTML
// already on the topics).
func ResaveAll() {
	log.SetOutput(os.Stdout) // so we can redirect it
	if err := models.InitDB(); err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	log.Printf("Database: %s", models.DbConnStringSafe())
	blogs, _ := models.BlogGetAll()
	for _, b := range blogs {
		log.Printf("re-saving %d - %s", b.Id, b.Title)
		blog, _ := models.BlogGetById(b.Id)
		blog.Save()
	}
}
