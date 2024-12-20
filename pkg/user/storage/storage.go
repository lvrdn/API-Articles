package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"rwa/pkg/user"
	"time"
)

var errNoUpdate = errors.New("no data to update")

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (st *Storage) GetErrNoUpdate() error {
	return errNoUpdate
}

func (st *Storage) NewUser(user *user.User) error {
	var LastInsertId int
	var bioSQL, imageSQL sql.NullString

	if user.Bio == nil {
		bioSQL.Valid = false
	} else {
		bioSQL.String = *user.Bio
		bioSQL.Valid = true
	}

	if user.Image == nil {
		imageSQL.Valid = false
	} else {
		imageSQL.String = *user.Image
		imageSQL.Valid = true
	}

	err := st.db.QueryRow("INSERT INTO users(email,username,password_hashed,bio,image,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id",
		user.Email, user.Username, user.PasswordHashed, bioSQL, imageSQL, user.CreatedAt, user.UpdatedAt,
	).Scan(&LastInsertId)

	if err != nil {
		return err
	}

	if LastInsertId == 0 {
		return fmt.Errorf("no last insert id")
	}

	return nil
}

func (st *Storage) Update(user *user.User) error {

	query := "UPDATE users SET "
	placeholderNum := 1
	args := make([]interface{}, 0)

	if user.Email != "" {
		query += fmt.Sprintf("email = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, user.Email)
	}

	if user.PasswordHashed != nil {
		query += fmt.Sprintf("password_hashed = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, user.PasswordHashed)
	}

	if user.Username != "" {
		query += fmt.Sprintf("username = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, user.Username)
	}

	if user.Bio != nil {
		query += fmt.Sprintf("bio = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, *user.Bio)
	}

	if user.Image != nil {
		query += fmt.Sprintf("image = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, *user.Image)
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

func (st *Storage) GetUserWithEmail(email string) (*user.User, error) {
	var id int
	var username string
	var createdAt, updatedAt time.Time
	var passwordHashed []byte
	var bioSQL, imageSQL sql.NullString

	err := st.db.
		QueryRow("SELECT id, username, password_hashed, bio, image, created_at, updated_at FROM users WHERE email=$1", email).
		Scan(&id, &username, &passwordHashed, &bioSQL, &imageSQL, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	bio := new(string)
	image := new(string)
	if bioSQL.Valid {
		*bio = bioSQL.String
	}
	if imageSQL.Valid {
		*image = imageSQL.String
	}

	return &user.User{
		ID:             id,
		Email:          email,
		PasswordHashed: passwordHashed,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		Username:       username,
		Bio:            bio,
		Image:          image,
	}, nil
}

func (st *Storage) GetUserWithID(id int) (*user.User, error) {

	var email, username string
	var createdAt, updatedAt time.Time
	var passwordHashed []byte
	var bioSQL, imageSQL sql.NullString

	err := st.db.
		QueryRow("SELECT email, username, password_hashed, bio, image, created_at, updated_at FROM users WHERE id=$1", id).
		Scan(&email, &username, &passwordHashed, &bioSQL, &imageSQL, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	bio := new(string)
	image := new(string)
	if bioSQL.Valid {
		*bio = bioSQL.String
	}
	if imageSQL.Valid {
		*image = imageSQL.String
	}

	return &user.User{
		ID:             id,
		Email:          email,
		PasswordHashed: passwordHashed,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		Username:       username,
		Bio:            bio,
		Image:          image,
	}, nil
}

func (st *Storage) Delete(id int) error {
	_, err := st.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (st *Storage) GetPasswordHasherWithID(id int) ([]byte, error) {
	var passwordHashed []byte
	err := st.db.QueryRow("SELECT password_hashed FROM users WHERE id=$1", id).Scan(&passwordHashed)
	if err != nil {
		return nil, err
	}
	return passwordHashed, nil
}

func (st *Storage) CheckUniqueEmail(email string) (bool, error) {
	var ok bool
	err := st.db.QueryRow("SELECT EXISTS (SELECT id FROM users WHERE email=$1)", email).Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (st *Storage) CheckUniqueUsername(username string) (bool, error) {
	var ok bool
	err := st.db.QueryRow("SELECT EXISTS (SELECT id FROM users WHERE username=$1)", username).Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}
