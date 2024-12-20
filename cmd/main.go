package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rwa/config"
	"rwa/pkg/article"
	"rwa/pkg/session"
	"rwa/pkg/user"
	"syscall"

	articleST "rwa/pkg/article/storage"
	sessionST "rwa/pkg/session/storage"
	userST "rwa/pkg/user/storage"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("get config error: [%s]\n", err.Error())
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.DBusername,
		cfg.DBpassword,
		cfg.DBhost,
		cfg.DBname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open sql connection failed, error: [%s]\n", err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("db  ping failed, error: [%s]\n", err.Error())
	}

	whiteList := map[string]map[string]struct{}{
		"/api/users": {
			"POST": struct{}{},
		},
		"/api/users/login": {
			"POST": struct{}{},
		},
		"/api/articles": {
			"GET": struct{}{},
		},
	}

	sessionHandler := session.NewSessionHandler(
		sessionST.NewStorage(db),
		whiteList,
	)

	userManager := user.NewUserHandler(
		userST.NewStorage(db),
		sessionHandler,
	)

	articleManager := article.NewArticleHandler(
		articleST.NewStorage(db),
		sessionHandler,
	)

	router := mux.NewRouter()

	//user
	//white list
	router.HandleFunc("/api/users", userManager.Register).Methods(http.MethodPost)
	router.HandleFunc("/api/users/login", userManager.Login).Methods(http.MethodPost)
	//other
	router.HandleFunc("/api/user/logout", userManager.Logout)
	router.HandleFunc("/api/user", userManager.GetUserInfo).Methods(http.MethodGet)
	router.HandleFunc("/api/user", userManager.UpdateUserInfo).Methods(http.MethodPut)
	router.HandleFunc("/api/user", userManager.DeleteUser).Methods(http.MethodDelete)

	//article
	//white list
	router.HandleFunc("/api/articles", articleManager.ShowAll).Methods(http.MethodGet)
	router.HandleFunc("/api/articles/{id:[0-9]+}", articleManager.ShowArticle).Methods(http.MethodGet)
	//other
	router.HandleFunc("/api/articles", articleManager.Create).Methods(http.MethodPost)
	router.HandleFunc("/api/articles", articleManager.Update).Methods(http.MethodPut)
	router.HandleFunc("/api/articles", articleManager.Delete).Methods(http.MethodDelete)

	//middleware
	router.Use(userManager.SessionManager.AuthMiddleware)

	server := http.Server{
		Addr:    ":" + cfg.HTTPport,
		Handler: router,
	}

	go func() {
		log.Println("start server")
		server.ListenAndServe()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	server.Shutdown(context.Background())
	log.Println("server stopped")
}
