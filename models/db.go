package models

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

type DbSettings struct {
	driver     string
	user       string
	password   string
	database   string
	connString string
}

var dbSettings DbSettings

func InitDB() error {
	dbSettings = DbSettings{
		driver:   env("DB_DRIVER", "mysql"),
		user:     env("DB_USER", "root"),
		password: env("DB_PASSWORD", ""),
		database: env("DB_NAME", "hkdb"),
	}
	dbSettings.connString = fmt.Sprintf("%s:%s@/%s?parseTime=true", dbSettings.user, dbSettings.password, dbSettings.database)
	return CreateDefaultUsers()
}

func DbConnStringSafe() string {
	return fmt.Sprintf("%s:%s@/%s", dbSettings.user, "***", dbSettings.database)
}

func connectDB() (*sql.DB, error) {
	return sql.Open(dbSettings.driver, dbSettings.connString)
}

func env(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}

// Returns UTC Now in a format that is recognized by MySQL
// MySQL doesn't recognize the RFC3339 standard (T between date and time
// and timezone offset at the end https://golang.org/pkg/time/#pkg-constants)
func dbUtcNow() string {
	t := time.Now().UTC()
	s := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return s
}

func dateValue(t mysql.NullTime) string {
	if t.Valid {
		return t.Time.String()[0:10]
	}
	return ""
}

func timeValue(t mysql.NullTime) string {
	if t.Valid {
		return t.Time.String()
	}
	return ""
}

func stringValue(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func intValue(v sql.NullInt64) int {
	if v.Valid {
		return int(v.Int64)
	}
	return 0
}
