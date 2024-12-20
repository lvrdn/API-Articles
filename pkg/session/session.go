package session

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"rwa/pkg/utils"
	"strings"

	"github.com/google/uuid"
)

type forSession string

const ctxKey forSession = "key"

type SessionHandler struct {
	Storage   Storage
	WhiteList map[string]map[string]struct{}
}

type Session struct {
	UserID     int
	SessionKey string
}

type Storage interface {
	Create(sessionKey string, userID int) error
	CheckSession(sessionKey string) (int, error)
	Delete(sessionKey string) error
	DeleteAll(userID int) error
}

func NewSessionHandler(storage Storage, list map[string]map[string]struct{}) *SessionHandler {
	return &SessionHandler{
		Storage:   storage,
		WhiteList: list,
	}
}

func sessionFromContext(r *http.Request) (*Session, error) {
	session, ok := r.Context().Value(ctxKey).(*Session)
	if !ok {
		return nil, fmt.Errorf("no session in context")
	}
	return session, nil
}

func (sh *SessionHandler) IdFromSessionContext(r *http.Request) (int, error) {
	session, ok := r.Context().Value(ctxKey).(*Session)
	if !ok {
		return 0, fmt.Errorf("no session in context")
	}
	return session.UserID, nil
}

func (sh *SessionHandler) Create(userID int) (string, error) {

	sessionKey := uuid.New().String()
	err := sh.Storage.Create(sessionKey, userID)
	if err != nil {
		return "", err
	}

	return sessionKey, nil
}

func (sh *SessionHandler) Check(r *http.Request) (*Session, error) {
	sessionKeyFromRec := r.Header.Get("Authorization")
	if sessionKeyFromRec == "" {
		return nil, getErrNoAuth()
	}

	var userID int
	userID, err := sh.Storage.CheckSession(sessionKeyFromRec)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return nil, getErrNoAuth()
		}
		return nil, err
	}

	return &Session{
		UserID:     userID,
		SessionKey: sessionKeyFromRec,
	}, nil
}

func (sh *SessionHandler) Delete(r *http.Request) error {
	session, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	err = sh.Storage.Delete(session.SessionKey)
	if err != nil {
		return err
	}

	return nil
}

func (sh *SessionHandler) DeleteAll(r *http.Request) error {
	session, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	err = sh.Storage.DeleteAll(session.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (sh *SessionHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		url := r.URL.Path
		if strings.Count(url, "/") == 3 {
			url = url[:strings.LastIndex(url, "/")]
		}

		if methods, ok := sh.WhiteList[url]; ok {
			if _, ok := methods[r.Method]; ok {
				next.ServeHTTP(w, r)
				return
			}
		}

		session, err := sh.Check(r)
		if err != nil {
			if err.Error() == getErrNoAuth().Error() {
				utils.SendErrMessage(w, r, "no auth", http.StatusUnauthorized)
				return
			}
			log.Printf("checking session error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getErrNoAuth() error {
	return fmt.Errorf("no auth")
}
