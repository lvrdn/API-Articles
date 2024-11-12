package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func GetApp() http.Handler {

	dsn := "postgresql://localhost:5432/realworld?password=1234&sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("cant open db by db driver, err: ", err)
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("cant connect to db, err: ", err)
	}

	list := map[string][]string{
		"/api/users":       {"POST"},
		"/api/users/login": {"POST"},
		"/api/articles":    {"GET"},
		"/":                {},
	}

	userManager := &UserHandler{
		St: NewStorageDB(db),
		SM: NewSessionDB(db, list),
	}

	articleManager := &ArticleHandler{
		St: NewStorageDBA(db),
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/users", userManager.Register)
	router.HandleFunc("/api/users/login", userManager.Login)
	router.HandleFunc("/api/user/logout", userManager.Logout)
	router.HandleFunc("/api/user", userManager.GetUserInfo).Methods(http.MethodGet)
	router.HandleFunc("/api/user", userManager.UpdateUserInfo).Methods(http.MethodPut)

	router.HandleFunc("/api/articles", articleManager.Create).Methods(http.MethodPost)
	router.HandleFunc("/api/articles", articleManager.Show).Methods(http.MethodGet)
	router.HandleFunc("/api/articles", articleManager.Update).Methods(http.MethodPut)
	router.HandleFunc("/api/articles", articleManager.Delete).Methods(http.MethodDelete)

	router.Use(userManager.SM.AuthMiddleware)

	return router
}
