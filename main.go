package main

import (
	"flag"

	"hectorcorrea.com/hk/tasks"
	"hectorcorrea.com/hk/web"
)

func main() {
	var address = flag.String("address", "localhost:9001", "Address where server will listen for connections.")
	var resave = flag.String("resave", "", "Pass \"yes\" to resave all blog posts and recalculate the HTML content.")
	var scan = flag.String("scan", "", "Pass full path to folder to scan for photos that need to be added to the database.")
	var addUser = flag.String("addUser", "", "Adds a new guest user/password")
	flag.Parse()

	if *resave == "yes" {
		tasks.ResaveAll()
		return
	} else if *scan != "" {
		tasks.ScanPhotos(*scan)
		return
	} else if *addUser != "" {
		tasks.AddUser(*addUser)
		return
	}

	web.StartWebServer(*address)
}
