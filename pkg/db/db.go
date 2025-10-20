package db

import (
	"database/sql"
	"os"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    repeat VARCHAR(128) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := false
	if err != nil {
		install = true
	}
	
	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	
	if install {
		_, err = database.Exec(schema)
		if err != nil {
			return err
		}
	}
	
	DB = database
	return nil
}

// GetDB возвращает указатель на объект базы данных
func GetDB() *sql.DB {
	return DB
}

// Close закрывает соединение с базой данных
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}