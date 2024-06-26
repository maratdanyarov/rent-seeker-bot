package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) CreateTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER UNIQUE,
		state TEXT,
		property_type TEXT,
		price_range TEXT,
		bedrooms TEXT,
		furnished TEXT,
		area TEXT
	);
	`
	_, err := db.Exec(query)
	return err
}

func (db *DB) SaveUser(chatID int64, state, propertyType, priceRange, bedrooms, furnished, area string) error {
	query := `
	INSERT INTO users (chat_id, state, property_type, price_range, bedrooms, furnished, area)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(chat_id) DO UPDATE SET
		state = ?,
		property_type = ?,
		price_range = ?,
		bedrooms = ?,
		furnished = ?,
		area = ?
	`
	_, err := db.Exec(query, chatID, state, propertyType, priceRange, bedrooms, furnished, area,
		state, propertyType, priceRange, bedrooms, furnished, area)
	return err
}

func (db *DB) GetUser(chatID int64) (*UserData, error) {
	query := `SELECT state, property_type, price_range, bedrooms, furnished, area FROM users WHERE chat_id = ?`
	var userData UserData
	err := db.QueryRow(query, chatID).Scan(&userData.State, &userData.PropertyType, &userData.PriceRange,
		&userData.Bedrooms, &userData.Furnished, &userData.Area)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &userData, nil
}

type UserData struct {
	State        string
	PropertyType string
	PriceRange   string
	Bedrooms     string
	Furnished    string
	Area         string
}
