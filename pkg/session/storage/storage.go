package storage

import "database/sql"

type Storage struct {
	DB *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		DB: db,
	}
}

func (st *Storage) Create(sessionKey string, userID int) error {

	_, err := st.DB.Exec("INSERT INTO sessions(session_key, user_id) VALUES($1,$2)", sessionKey, userID)
	if err != nil {
		return err
	}
	return nil
}

func (st *Storage) CheckSession(sessionKey string) (int, error) {

	var userID int
	err := st.DB.QueryRow("SELECT user_id FROM sessions WHERE session_key = $1", sessionKey).Scan(&userID)
	if err != nil {
		return 0, err
	}

	if userID == 0 {
		return 0, nil //TODO подумать что вернуть, если id=0 ?sql.NoRows()
	}

	return userID, nil
}

func (st *Storage) Delete(sessionKey string) error {

	_, err := st.DB.Exec("DELETE FROM sessions WHERE session_key = $1", sessionKey)
	if err != nil {
		return err
	}

	return nil
}

func (st *Storage) DeleteAll(userID int) error {

	_, err := st.DB.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	return nil
}
