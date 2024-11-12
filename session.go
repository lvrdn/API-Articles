package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
)

type ctxKey int

const sessionKey ctxKey = 1

func RandStringRunes(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type SessionsDB struct {
	DB        *sql.DB
	WhiteList map[string][]string
}

func NewSessionDB(db *sql.DB, list map[string][]string) *SessionsDB {
	return &SessionsDB{
		DB:        db,
		WhiteList: list,
	}
}

type Session struct {
	UserID     string
	SessionKey string
}

func SessionFromContext(r *http.Request) (*Session, error) {
	sess, ok := r.Context().Value(sessionKey).(*Session)
	if !ok {
		return nil, fmt.Errorf("no auth")
	}
	return sess, nil // добавить обработку ошибки
}

func (sm *SessionsDB) Create(userID string) (string, error) {
	sessKey := RandStringRunes(32)
	_, err := sm.DB.Exec("INSERT INTO sessions(user_id, session_key) VALUES($1,$2)", userID, sessKey)
	if err != nil {
		return "", err
	}
	return sessKey, nil
}

func (sm *SessionsDB) Check(r *http.Request) (*Session, error) {
	sessKeyFromRec := r.Header.Get("Authorization")
	if sessKeyFromRec == "" {
		return nil, fmt.Errorf("no auth token in request")
	}
	sessKeyFromRec = strings.TrimPrefix(sessKeyFromRec, "Token ")
	var userID string
	err := sm.DB.QueryRow("SELECT user_id FROM sessions WHERE session_key = $1", sessKeyFromRec).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("check: no rows from sessions db")
	}

	return &Session{
		UserID:     userID,
		SessionKey: sessKeyFromRec,
	}, nil
}

func (sm *SessionsDB) DeleteSession(r *http.Request) error {
	session, err := SessionFromContext(r)
	if err != nil {
		return err
	}
	_, err = sm.DB.Exec("DELETE FROM sessions WHERE session_key = $1", session.SessionKey)

	if err != nil {
		return err
	}

	return nil
}

func (sm *SessionsDB) DeleteAllSession(r *http.Request) error {
	session, err := SessionFromContext(r)
	if err != nil {
		return err
	}
	_, err = sm.DB.Exec("DELETE FROM sessions WHERE user_id = $1", session.UserID)

	if err != nil {
		return err
	}

	return nil
}

func (sm *SessionsDB) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if methods, ok := sm.WhiteList[r.URL.Path]; ok {
			for _, method := range methods {
				if method == r.Method {
					next.ServeHTTP(w, r)
					return
				}
			}

		}
		sess, err := sm.Check(r)
		if err != nil {
			http.Error(w, "no auth", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), sessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
