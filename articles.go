package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ArticleHandler struct {
	St StorageA
}

type StorageA interface {
	Add(*Article) error
	Update(*Article, string) error
	Delete(string, string) error
	Get(map[string]string) ([]*Article, error)
	GetAuthorWithID(string) (string, string, error)
	GetErrNoUpdate() error
}

type Article struct {
	ID             string
	Author         *UserProfile `json:"author"`
	Body           string       `json:"body"`
	CreatedAt      interface{}  `json:"createdAt"`
	Description    string       `json:"description"`
	Favorited      bool         `json:"favorited"`
	FavoritesCount int          `json:"favoritesCount"`
	Slug           string       `json:"slug"`
	TagList        []string     `json:"tagList"`
	Title          string       `json:"title"`
	UpdatedAt      interface{}  `json:"updatedAt"`
}

func (ah *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {

	sess, err := SessionFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error with read r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataFromBody := make(map[string]*Article)
	err = json.Unmarshal(body, &dataFromBody)
	if err != nil {
		fmt.Println("error with unmarshal json from r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newArticle, ok := dataFromBody["article"]
	if !ok {
		http.Error(w, "bad request, no article data", http.StatusBadRequest)
		return
	}

	newArticle.CreatedAt = time.Now()
	newArticle.UpdatedAt = time.Now()

	author, bio, err := ah.St.GetAuthorWithID(sess.UserID)
	if err != nil {
		fmt.Println("error with get author about author from db", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = ah.St.Add(&Article{
		Author: &UserProfile{
			ID: sess.UserID,
		},
		Body:        newArticle.Body,
		CreatedAt:   newArticle.CreatedAt,
		Description: newArticle.Description,
		TagList:     newArticle.TagList,
		Title:       newArticle.Title,
		UpdatedAt:   newArticle.CreatedAt,
	},
	)
	if err != nil {
		fmt.Println("error with add new article to db", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := Response{"article": &Article{
		Author: &UserProfile{
			Bio:       bio,
			Username:  author,
			CreatedAt: nil,
			UpdatedAt: nil,
		},
		Body:        newArticle.Body,
		Title:       newArticle.Title,
		Description: newArticle.Description,
		CreatedAt:   newArticle.CreatedAt,
		UpdatedAt:   newArticle.UpdatedAt,
		TagList:     newArticle.TagList,
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

func (ah *ArticleHandler) Show(w http.ResponseWriter, r *http.Request) {

	params := make(map[string]string)
	if author := r.URL.Query().Get("author"); author != "" {
		params["author"] = author
	}
	if tag := r.URL.Query().Get("tag"); tag != "" {
		params["tag"] = tag
	}

	articles, err := ah.St.Get(params)
	if err != nil {
		fmt.Println("error with get articles from db", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := Response{
		"articles":      articles,
		"articlesCount": len(articles),
	}
	dataToSend, err := json.Marshal(response)
	if err != nil {
		fmt.Println("error with marshal json with error", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(dataToSend)
}

func (ah *ArticleHandler) Update(w http.ResponseWriter, r *http.Request) {

	sess, err := SessionFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error with read r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataFromBody := make(map[string]*Article)
	err = json.Unmarshal(body, &dataFromBody)
	if err != nil {
		fmt.Println("error with unmarshal json from r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	articleFromReq, ok := dataFromBody["article"]
	if !ok {
		http.Error(w, "bad request, no article data", http.StatusBadRequest)
		return
	}

	err = ah.St.Update(articleFromReq, sess.UserID)
	if err != nil {
		if err == ah.St.GetErrNoUpdate() {
			http.Error(w, "bad request, no article data to update", http.StatusBadRequest)
			return
		}
		fmt.Println("error with update article data", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (ah *ArticleHandler) Delete(w http.ResponseWriter, r *http.Request) {

	sess, err := SessionFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error with read r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataFromBody := make(map[string]*Article)
	err = json.Unmarshal(body, &dataFromBody)
	if err != nil {
		fmt.Println("error with unmarshal json from r.body", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	articleFromReq, ok := dataFromBody["article"]
	if !ok {
		http.Error(w, "bad request, no article data", http.StatusBadRequest)
		return
	}

	err = ah.St.Delete(articleFromReq.Slug, sess.UserID)
	if err != nil {
		fmt.Println("error with delete article from db", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
