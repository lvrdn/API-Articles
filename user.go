package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/crypto/argon2"
)

type UserHandler struct {
	St Storage
	SM SessionManager
}

type Storage interface {
	Add(*UserProfile) error
	Update(*UserProfile) error
	GetUserWithEmail(string) (*UserProfile, error)
	GetUserWithID(string) (*UserProfile, error)
	CheckUniqueUsername(string) (bool, error)
	CheckUniqueEmail(string) (bool, error)
	GetErrNoUpdate() error
}

type SessionManager interface {
	Check(*http.Request) (*Session, error)
	Create(string) (string, error)
	DeleteSession(*http.Request) error
	DeleteAllSession(*http.Request) error
	AuthMiddleware(http.Handler) http.Handler
}

type Response map[string]interface{}

type UserProfile struct {
	ID             string `json:"id" testdiff:"ignore"`
	Email          string `json:"email"`
	password       string
	passwordHashed []byte
	CreatedAt      interface{} `json:"createdAt"`
	UpdatedAt      interface{} `json:"updatedAt"`
	Username       string      `json:"username"`
	Bio            string      `json:"bio"`
	Image          string      `json:"image"`
	Token          string      `json:"token"`
	Following      bool
}

func (uh *UserHandler) hashPass(plainPassword, salt string) []byte {
	hashedPass := argon2.IDKey([]byte(plainPassword), []byte(salt), 1, 64*1024, 4, 32)
	res := make([]byte, len(salt))
	copy(res, salt)
	return append(res, hashedPass...)
}

func (uh *UserHandler) checkUniqueEmail(w http.ResponseWriter, r *http.Request, email string) bool {

	uniqueEmail, err := uh.St.CheckUniqueEmail(email)

	if err != nil {
		fmt.Println("error with checking unique email", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}
	if !uniqueEmail {
		http.Error(w, "user with this email exist", http.StatusBadRequest)
		return true
	}
	return false
}

func (uh *UserHandler) checkUniqueUsername(w http.ResponseWriter, r *http.Request, username string) bool {

	uniqueUsername, err := uh.St.CheckUniqueUsername(username)

	if err != nil {
		fmt.Println("error with checking unique username", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}
	if !uniqueUsername {
		http.Error(w, "user with this username exist", http.StatusBadRequest)
		return true
	}
	return false
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error with read r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataFromBody := make(map[string]*UserProfile)
	err = json.Unmarshal(body, &dataFromBody)
	if err != nil {
		fmt.Println("error with unmarshal json from r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newUser, ok := dataFromBody["user"]
	if !ok {
		http.Error(w, "bad request, no user data", http.StatusBadRequest)
		return
	}

	newUser.CreatedAt = time.Now()
	newUser.UpdatedAt = time.Now()

	newUser.passwordHashed = uh.hashPass(newUser.password, RandStringRunes(8))

	if uh.checkUniqueEmail(w, r, newUser.Email) {
		return
	}

	if uh.checkUniqueUsername(w, r, newUser.Username) {
		return
	}

	err = uh.St.Add(newUser)
	if err != nil {
		fmt.Println("error with add new user to storage", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := Response{"user": &UserProfile{
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Username:  newUser.Username,
	}}

	dataToSend, err := json.Marshal(response)
	if err != nil {
		fmt.Println("error with marshal json with error", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(dataToSend)

}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error with read r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataFromBody := make(map[string]*UserProfile)
	err = json.Unmarshal(body, &dataFromBody)
	if err != nil {
		fmt.Println("error with unmarshal json from r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, ok := dataFromBody["user"]
	if !ok {
		http.Error(w, "bad request, no user data", http.StatusBadRequest)
		return
	}

	userFromDB, err := uh.St.GetUserWithEmail(user.Email)
	user.passwordHashed = uh.hashPass(user.password, string(userFromDB.passwordHashed[0:8]))

	if err == sql.ErrNoRows || !bytes.Equal(userFromDB.passwordHashed, user.passwordHashed) {
		response := Response{"error": "invalid login or password"}
		dataResponse, err := json.Marshal(response)
		if err != nil {
			fmt.Println("error with marshal json with error", r.URL.Path, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(dataResponse)
		return
	}

	session, err := uh.SM.Create(userFromDB.ID)
	if err != nil {
		fmt.Println("error with create session for user", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := Response{"user": &UserProfile{
		Email:     userFromDB.Email,
		CreatedAt: userFromDB.CreatedAt,
		UpdatedAt: userFromDB.UpdatedAt,
		Username:  userFromDB.Username,
		Token:     session,
	}}
	dataToSend, err := json.Marshal(response)
	if err != nil {
		fmt.Println("error with marshal json with error", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(dataToSend)

}

func (uh *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {

	err := uh.SM.DeleteSession(r)
	if err != nil {
		fmt.Println("error with marshal json with error", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (uh *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {

	sess, err := SessionFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := uh.St.GetUserWithID(sess.UserID)
	if err != nil {
		fmt.Println("error with get user with id from db", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := Response{"user": &UserProfile{
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
	}}
	dataToSend, err := json.Marshal(response)
	if err != nil {
		fmt.Println("error with marshal json with error", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(dataToSend)

}

func (uh *UserHandler) UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error with read r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataFromBody := make(map[string]*UserProfile)
	err = json.Unmarshal(body, &dataFromBody)
	if err != nil {
		fmt.Println("error with unmarshal json from r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userFromReq, ok := dataFromBody["user"]
	if !ok {
		http.Error(w, "bad request, no user data", http.StatusBadRequest)
		return
	}

	sess, err := SessionFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if userFromReq.Email != "" {
		if uh.checkUniqueEmail(w, r, userFromReq.Email) {
			return
		}
	}

	if userFromReq.Username != "" {
		if uh.checkUniqueUsername(w, r, userFromReq.Username) {
			return
		}
	}

	userFromReq.ID = sess.UserID
	if userFromReq.password != "" {
		userFromReq.passwordHashed = uh.hashPass(userFromReq.password, RandStringRunes(8))
	}

	err = uh.St.Update(userFromReq)
	if err != nil {
		if err == uh.St.GetErrNoUpdate() {
			http.Error(w, "bad request, no user data to update", http.StatusBadRequest)
			return
		}
		fmt.Println("error with update user data", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	user, err := uh.St.GetUserWithID(sess.UserID)
	if err != nil {
		fmt.Println("error with get user with id from db", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := Response{"user": &UserProfile{
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
		Token:     sess.SessionKey,
	}}
	dataToSend, err := json.Marshal(response)
	if err != nil {
		fmt.Println("error with marshal json with error", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(dataToSend)

}
