package main

import (
	"database/sql"
	"fmt"
	"time"
)

type StorageDB struct {
	db *sql.DB
}

func NewStorageDB(db *sql.DB) *StorageDB {
	return &StorageDB{
		db: db,
	}
}

func (st *StorageDB) GetErrNoUpdate() error {
	return fmt.Errorf("no data to update")
}

func (st *StorageDB) Add(new *UserProfile) error {
	var LastInsertId int
	err := st.db.QueryRow("INSERT INTO users(email,username,password_hashed,bio,image,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id",
		new.Email, new.Username, new.passwordHashed, new.Bio, new.Image, new.CreatedAt, new.UpdatedAt,
	).Scan(&LastInsertId)

	if err != nil {
		return err
	}

	if LastInsertId == 0 {
		return fmt.Errorf("no last insert id")
	}

	return nil
}

func (st *StorageDB) Update(user *UserProfile) error {

	query := "UPDATE users SET "
	placeholderNum := 1
	args := make([]interface{}, 0)

	if user.Email != "" {
		query += fmt.Sprintf("email = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, user.Email)
	}

	if user.passwordHashed != nil {
		query += fmt.Sprintf("password_hashed = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, user.passwordHashed)
	}

	if user.Username != "" {
		query += fmt.Sprintf("username = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, user.Username)
	}

	if user.Bio != "" {
		query += fmt.Sprintf("bio = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, user.Bio)
	}

	if user.Image != "" {
		query += fmt.Sprintf("image = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, user.Image)
	}

	if placeholderNum == 1 {
		return st.GetErrNoUpdate()
	}
	user.UpdatedAt = time.Now()
	query += fmt.Sprintf("updated_at = $%v WHERE id = $%v ", placeholderNum, placeholderNum+1)

	args = append(args, user.UpdatedAt, user.ID)

	_, err := st.db.Exec(query, args...)

	if err != nil {
		return err
	}
	return nil
}

func (st *StorageDB) GetUserWithEmail(email string) (*UserProfile, error) {

	var id, username, passwordHashed, bio, image string
	var createdAt, updatedAt time.Time
	err := st.db.
		QueryRow("SELECT id, username, password_hashed, bio, image, created_at, updated_at FROM users WHERE email=$1", email).
		Scan(&id, &username, &passwordHashed, &bio, &image, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &UserProfile{
		ID:             id,
		Email:          email,
		Username:       username,
		passwordHashed: []byte(passwordHashed),
		Bio:            bio,
		Image:          image,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func (st *StorageDB) GetUserWithID(id string) (*UserProfile, error) {

	var email, username, passwordHashed, bio, image string
	var createdAt, updatedAt time.Time
	err := st.db.
		QueryRow("SELECT email, username, password_hashed, bio, image, created_at, updated_at FROM users WHERE id=$1", id).
		Scan(&email, &username, &passwordHashed, &bio, &image, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &UserProfile{
		ID:             id,
		Email:          email,
		Username:       username,
		passwordHashed: []byte(passwordHashed),
		Bio:            bio,
		Image:          image,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func (st *StorageDB) CheckUniqueUsername(username string) (bool, error) {
	var id int
	err := st.db.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&id)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func (st *StorageDB) CheckUniqueEmail(email string) (bool, error) {
	var id int
	err := st.db.QueryRow("SELECT id FROM users WHERE email=$1", email).Scan(&id)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}
