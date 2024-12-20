package user

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"rwa/pkg/utils"
	"time"
)

type UserHandler struct {
	Storage        Storage
	SessionManager SessionManager
}

func NewUserHandler(st Storage, sm SessionManager) *UserHandler {
	return &UserHandler{
		Storage:        st,
		SessionManager: sm,
	}
}

type Storage interface {
	NewUser(user *User) error
	GetUserWithEmail(email string) (*User, error)
	GetUserWithID(id int) (*User, error)
	GetPasswordHasherWithID(id int) ([]byte, error)
	Update(*User) error
	Delete(id int) error
	CheckUniqueUsername(username string) (bool, error)
	CheckUniqueEmail(email string) (bool, error)
	GetErrNoUpdate() error
}

type SessionManager interface {
	Create(userID int) (string, error)
	Delete(r *http.Request) error
	DeleteAll(r *http.Request) error
	AuthMiddleware(next http.Handler) http.Handler
	IdFromSessionContext(r *http.Request) (int, error)
}

type User struct {
	ID             int       `json:"id"`
	Email          string    `json:"email"`
	Password       string    `json:"password,omitempty"`
	PasswordHashed []byte    `json:"-"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Username       string    `json:"username"`
	Bio            *string   `json:"bio"`
	Image          *string   `json:"image"`
}

func (uh *UserHandler) checkUniqueEmail(w http.ResponseWriter, r *http.Request, email string) bool {

	if email == "" {
		utils.SendErrMessage(w, r, "email must be not empty", http.StatusBadRequest)
		return false
	}

	unique, err := uh.Storage.CheckUniqueEmail(email)
	if err != nil {
		log.Printf("checking unique email error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}

	if unique {
		utils.SendErrMessage(w, r, "user with this email exist", http.StatusBadRequest)
		return false
	}

	return true
}

func (uh *UserHandler) checkUniqueUsername(w http.ResponseWriter, r *http.Request, username string) bool {

	if username == "" {
		utils.SendErrMessage(w, r, "username must be not empty", http.StatusBadRequest)
		return false
	}

	unique, err := uh.Storage.CheckUniqueUsername(username)
	if err != nil {
		log.Printf("checking unique username error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}

	if unique {
		utils.SendErrMessage(w, r, "user with this username exist", http.StatusBadRequest)
		return false
	}
	return true
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {

	body := utils.ReadBody(w, r)
	if body == nil {
		return
	}

	newUser := unmarshalBody(w, r, body)
	if newUser == nil {
		return
	}

	if !uh.checkUniqueEmail(w, r, newUser.Email) {
		return
	}

	if !uh.checkUniqueUsername(w, r, newUser.Username) {
		return
	}

	if newUser.Password == "" {
		utils.SendErrMessage(w, r, "password must be not empty", http.StatusBadRequest)
		return
	}

	now := time.Now()
	newUser.CreatedAt = now
	newUser.UpdatedAt = now

	salt := randStringRunes(8)
	newUser.PasswordHashed = hashPassword(newUser.Password, salt)

	err := uh.Storage.NewUser(newUser)
	if err != nil {
		log.Printf("add new user to storage error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {

	body := utils.ReadBody(w, r)
	if body == nil {
		return
	}

	userFromReq := unmarshalBody(w, r, body)
	if userFromReq == nil {
		return
	}

	if userFromReq.Email == "" || userFromReq.Password == "" {
		utils.SendErrMessage(w, r, "email or password must be not empty", http.StatusBadRequest)
		return
	}

	user, err := uh.Storage.GetUserWithEmail(userFromReq.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.SendErrMessage(w, r, "invalid email or password", http.StatusBadRequest)
			return
		}
		log.Printf("get user info with email error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	salt := string(user.PasswordHashed[0:8])
	userFromReq.PasswordHashed = hashPassword(userFromReq.Password, salt)

	if err == sql.ErrNoRows || !bytes.Equal(user.PasswordHashed, userFromReq.PasswordHashed) {
		utils.SendErrMessage(w, r, "invalid email or password", http.StatusBadRequest)
		return
	}

	sessionKey, err := uh.SessionManager.Create(user.ID)
	if err != nil {
		log.Printf("create session key error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", sessionKey)

	response := utils.Response{
		"user": user,
	}

	utils.SendResponse(w, r, response)

}

func (uh *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {

	deleteAll := r.Header.Get("DeleteAll")

	switch deleteAll {
	case "":
		err := uh.SessionManager.Delete(r)
		if err != nil {
			log.Printf("delete session error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case "true":
		err := uh.SessionManager.DeleteAll(r)
		if err != nil {
			log.Printf("delete all sessions error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

}

func (uh *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {

	id, err := uh.SessionManager.IdFromSessionContext(r)
	if err != nil {
		log.Printf("get user id error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := uh.Storage.GetUserWithID(id)
	if err != nil {
		log.Printf("get user info with id error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := utils.Response{"user": user}

	utils.SendResponse(w, r, response)
}

func (uh *UserHandler) UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	body := utils.ReadBody(w, r)
	if body == nil {
		return
	}

	userFromReq := unmarshalBody(w, r, body)
	if userFromReq == nil {
		return
	}

	id, err := uh.SessionManager.IdFromSessionContext(r)
	if err != nil {
		log.Printf("getting user id error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if userFromReq.Email != "" && !uh.checkUniqueEmail(w, r, userFromReq.Email) {
		return
	}

	if userFromReq.Username != "" && !uh.checkUniqueUsername(w, r, userFromReq.Username) {
		return
	}

	if userFromReq.Password != "" {
		NewPasswordHashed, err := uh.Storage.GetPasswordHasherWithID(id)
		if err != nil {
			log.Printf("getting hashed password error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		salt := string(NewPasswordHashed[0:8])
		OLDpasswordHashed := hashPassword(userFromReq.Password, salt)

		if bytes.Equal(NewPasswordHashed, OLDpasswordHashed) {
			utils.SendErrMessage(w, r, "you write your old password, need make new password", http.StatusBadRequest)
			return
		}

		salt = randStringRunes(8)
		userFromReq.PasswordHashed = hashPassword(userFromReq.Password, salt)
	}
	userFromReq.ID = id
	err = uh.Storage.Update(userFromReq)
	if err != nil {
		if err == uh.Storage.GetErrNoUpdate() {
			utils.SendErrMessage(w, r, "no user data to update", http.StatusBadRequest)
			return
		}
		log.Printf("update user data error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := uh.Storage.GetUserWithID(id)
	if err != nil {
		fmt.Println("error with get user with id from db", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := utils.Response{"user": user}

	utils.SendResponse(w, r, response)

}

func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {

	id, err := uh.SessionManager.IdFromSessionContext(r)
	if err != nil {
		log.Printf("get user id error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = uh.Storage.Delete(id)
	if err != nil {
		log.Printf("delete user error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
