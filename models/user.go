package models

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type User struct {
	Id   int64
	Type string // guest or admin
}

func CreateDefaultUsers() error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow("SELECT count(*) FROM users")
	var count int
	err = row.Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		err = createDefaultAdmin(db)
		if err == nil {
			err = createDefaultGuest(db)
		}
	}
	return err
}

func AddGuestUser(login, password string) error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return createUser(db, login, password, "guest")
}

func SetPassword(login, newPassword string) error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	hashedPassword := hashPassword(newPassword)
	sqlUpdate := "UPDATE users SET password = ? WHERE login = ?"
	_, err = db.Exec(sqlUpdate, hashedPassword, login)
	return err
}

func createDefaultAdmin(db *sql.DB) error {
	login := env("BLOG_USR", "user1")
	password := env("BLOG_PASS", "welcome1")
	log.Printf(fmt.Sprintf("Creating initial admin user: %s", login))
	return createUser(db, login, password, "admin")
}

func createDefaultGuest(db *sql.DB) error {
	login := env("BLOG_GUEST_USR", "user2")
	password := env("BLOG_GUEST_PASS", "welcome2")
	log.Printf(fmt.Sprintf("Creating initial guest user: %s", login))
	return createUser(db, login, password, "guest")
}

func createUser(db *sql.DB, login, password, userType string) error {
	hashedPwd := hashPassword(password)
	sqlInsert := `INSERT INTO users(login, name, password, type) VALUES(?, ?, ?, ?)`
	_, err := db.Exec(sqlInsert, login, login, hashedPwd, userType)
	return err
}

func hashPassword(password string) string {
	salt := env("BLOG_SALT", "")
	salted := password + salt
	data := []byte(salted)
	hashed := sha256.Sum256(data)
	return fmt.Sprintf("%x", hashed)
}

func LoginUser(login, password string) (bool, error) {
	db, err := connectDB()
	if err != nil {
		return false, err
	}
	defer db.Close()

	hashedPassword := hashPassword(password)
	row := db.QueryRow("SELECT id FROM users WHERE login = ? and password = ?", login, hashedPassword)
	var id int64
	err = row.Scan(&id)
	if err != nil {
		log.Printf("Login/password not found in database: %s/***", login)
		return false, err
	} else if id == 0 {
		return false, errors.New("User ID was zero")
	}
	return true, nil
}

func GetUserId(login string) (int64, error) {
	user, err := GetUserInfo(login)
	return user.Id, err
}

func GetUserInfo(login string) (User, error) {
	db, err := connectDB()
	if err != nil {
		return User{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, type FROM users WHERE login = ?", login)
	var userType sql.NullString
	var id int64
	err = row.Scan(&id, &userType)
	if err != nil {
		log.Printf("Error fetching id for user: %s", login)
	}

	user := User{Id: id, Type: stringValue(userType)}
	return user, err
}
