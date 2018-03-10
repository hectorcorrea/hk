package tasks

import (
	"fmt"
	"hk/models"
	"log"
	"os"
	"path/filepath"
)

// Scan files on a folder (and its subfolders) and adds them
// to the photos table if they are not there already.
func ScanPhotos(folder string) {
	log.Printf("Scanning for photos. Folder: %s", folder)

	if err := models.InitDB(); err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}

	log.Printf("Database: %s", models.DbConnStringSafe())

	// https://stackoverflow.com/a/6612243/446681
	err := filepath.Walk(folder, func(path string, f os.FileInfo, err error) error {
		isFile := !f.IsDir()
		if isFile {
			exist, err := models.PhotoExists(path)
			if err != nil {
				fmt.Printf("Error looking for photo: %s\n", path)
			} else if exist {
				// fmt.Printf("Photo already in DB: %s\n", path)
			} else {
				id, err := models.PhotoAdd(path, true)
				if err != nil {
					fmt.Printf("Error adding photo : %s\n", path)
				} else {
					fmt.Printf("Photo added: %s (%d)\n", path, id)
				}
			}
		} else {
			fmt.Printf("Skipped folder %s\n", path)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
