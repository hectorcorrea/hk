package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

type UserSession struct {
	SessionId string
	ExpiresOn time.Time
	Login     string
	UserType  string
}

func GetUserSession(sessionId string) (UserSession, error) {
	if sessionId == "" {
		return UserSession{}, errors.New("No ID was received")
	}

	db, err := connectDB()
	if err != nil {
		return UserSession{}, err
	}
	defer db.Close()

	sqlSelect := `
		SELECT expiresOn, users.login, users.type
		FROM sessions
		 	INNER JOIN users ON sessions.userId = users.id
		WHERE sessions.id = ?`
	row := db.QueryRow(sqlSelect, sessionId)
	var expiresOn mysql.NullTime
	var login sql.NullString
	var userType sql.NullString
	err = row.Scan(&expiresOn, &login, &userType)
	if err != nil {
		log.Printf("Error on scan: %s", err)
		return UserSession{}, err
	}

	if expiresOn.Valid && expiresOn.Time.After(time.Now().UTC()) {

		// Mark last time session has been used.
		now := time.Now().UTC()
		sqlUpdate := `UPDATE sessions SET lastSeenOn = ? WHERE id = ?`
		_, err = db.Exec(sqlUpdate, now, sessionId)
		if err != nil {
			log.Printf("Could not update session last seen on: %s", err)
		}

		s := UserSession{
			SessionId: sessionId,
			ExpiresOn: expiresOn.Time,
			Login:     stringValue(login),
			UserType:  stringValue(userType),
		}
		return s, nil
	}

	return UserSession{}, errors.New("UserSession has already expired")
}

func NewUserSession(login string) (UserSession, error) {
	return newSession(login, 365)
}

func NewTicketSession(login string) (UserSession, error) {
	return newSession(login, 60)
}

func DeleteUserSession(sessionId string) {
	db, err := connectDB()
	if err != nil {
		return
	}
	defer db.Close()

	sqlDelete := `DELETE FROM sessions WHERE id = ?`
	_, err = db.Exec(sqlDelete, sessionId)
}

func cleanSessions(db *sql.DB, userId int64, singleSessionUser bool) error {
	if singleSessionUser {
		// All sessions for this user (regardless of expiration date)
		sqlDelete := "DELETE FROM sessions WHERE userId = ?"
		_, err := db.Exec(sqlDelete, userId)
		if err != nil {
			return err
		}
	}

	// All expired sessions (regardless of the user)
	sqlDelete := "DELETE FROM sessions WHERE expiresOn < utc_timestamp()"
	_, err := db.Exec(sqlDelete)
	return err
}

// source: https://www.socketloop.com/tutorials/golang-how-to-generate-random-string
func newId() (string, error) {
	size := 32
	rb := make([]byte, size)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rb), nil
}

func newSession(login string, days int) (UserSession, error) {
	db, err := connectDB()
	if err != nil {
		return UserSession{}, err
	}
	defer db.Close()

	sessionId, err := newId()
	if err != nil {
		return UserSession{}, err
	}

	user, err := GetUserInfo(login)
	if err != nil {
		return UserSession{}, err
	}

	singleSessionUser := (user.Type == "admin")
	err = cleanSessions(db, user.Id, singleSessionUser)
	if err != nil {
		log.Printf("Error cleaning older sessions for user %s, %s", login, err)
	}

	s := UserSession{
		SessionId: sessionId,
		Login:     login,
		ExpiresOn: time.Now().UTC().AddDate(0, 0, days),
		UserType:  user.Type,
	}

	now := time.Now().UTC()
	sqlInsert := `
		INSERT INTO sessions(id, userId, expiresOn, lastSeenOn)
		VALUES(?, ?, ?, ?)`
	_, err = db.Exec(sqlInsert, s.SessionId, user.Id, s.ExpiresOn, now)
	if err != nil {
		log.Printf("Error in SQL INSERT INTO sessions: %s", err)
	}
	return s, err
}
