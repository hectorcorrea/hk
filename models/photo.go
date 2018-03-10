package models

import (
	"database/sql"
)

type Photo struct {
	Id   int64
	Path string
}

func PhotoExists(path string) (bool, error) {
	db, err := connectDB()
	if err != nil {
		return false, err
	}
	defer db.Close()

	var id int64
	sqlSelect := `SELECT id FROM photos WHERE path = ?`
	err = db.QueryRow(sqlSelect, path).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return true, err
}

func PhotoAdd(path string, onDisk bool) (int64, error) {
	db, err := connectDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var onDiskInt int
	if onDisk {
		onDiskInt = 1
	}
	sqlInsert := `INSERT INTO photos(path, on_disk) VALUES(?, ?)`
	result, err := db.Exec(sqlInsert, path, onDiskInt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
