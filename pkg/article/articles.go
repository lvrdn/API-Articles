package article

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"rwa/pkg/utils"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mdigger/translit"
)

type ArticleHandler struct {
	Storage        Storage
	SessionManager SessionManager
}

func NewArticleHandler(storage Storage, sessionManager SessionManager) *ArticleHandler {
	return &ArticleHandler{
		Storage:        storage,
		SessionManager: sessionManager,
	}
}

type Storage interface {
	Add(new *Article) (int, error)
	Update(article *Article, userID int) error
	Delete(articleID, userID int) error
	GetArticles(filters map[string]string) ([]*Article, error)
	GetArticleWithID(id int) (*Article, error)
	GetErrNoUpdate() error
}

type SessionManager interface {
	IdFromSessionContext(r *http.Request) (int, error)
}

type AuthorManager interface{}

type Article struct {
	ID          int       `json:"id"`
	Author      *Author   `json:"author"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	Body        *string   `json:"body"`
	TagList     []string  `json:"tagList"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Author struct {
	ID       int
	Username string
	Image    string
}

func (ah *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {

	author := &Author{}
	var err error
	author.ID, err = ah.SessionManager.IdFromSessionContext(r)
	if err != nil {
		log.Printf("get user id error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := utils.ReadBody(w, r)
	if body == nil {
		return
	}

	newArticle := unmarshalBody(w, r, body)
	if newArticle == nil {
		return
	}

	if newArticle.Title == "" {
		utils.SendErrMessage(w, r, "title must be not empty", http.StatusBadRequest)
		return
	}

	now := time.Now()
	newArticle.CreatedAt = now
	newArticle.UpdatedAt = now
	newArticle.Slug = translit.Ru(newArticle.Title)
	newArticle.Author = author

	id, err := ah.Storage.Add(newArticle)
	if err != nil {
		log.Printf("add new article error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := utils.Response{
		"article": utils.Response{
			"id": id,
		},
	}

	w.WriteHeader(http.StatusCreated)
	utils.SendResponse(w, r, response)
}

func (ah *ArticleHandler) ShowAll(w http.ResponseWriter, r *http.Request) {

	params := make(map[string]string)
	if author := r.URL.Query().Get("author"); author != "" {
		params["author"] = author
	} else if tag := r.URL.Query().Get("tag"); tag != "" {
		params["tag"] = tag
	}

	articles, err := ah.Storage.GetArticles(params)
	if err != nil {
		fmt.Println("error with get articles from db", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := utils.Response{
		"articles":      articles,
		"articlesCount": len(articles),
	}

	utils.SendResponse(w, r, response)
}

func (ah *ArticleHandler) ShowArticle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	article, err := ah.Storage.GetArticleWithID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.SendErrMessage(w, r, "bad id, no data", http.StatusBadRequest)
			return
		}
		log.Printf("get article with id error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := utils.Response{
		"article": article,
	}

	utils.SendResponse(w, r, response)
}

func (ah *ArticleHandler) Update(w http.ResponseWriter, r *http.Request) {

	userID, err := ah.SessionManager.IdFromSessionContext(r)
	if err != nil {
		log.Printf("get user id error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := utils.ReadBody(w, r)
	if body == nil {
		return
	}

	articleFromReq := unmarshalBody(w, r, body)
	if articleFromReq == nil {
		return
	}

	if articleFromReq.ID == 0 {
		utils.SendErrMessage(w, r, "article id must be not 0", http.StatusBadRequest)
		return
	}

	articleFromReq.Slug = ""

	if articleFromReq.Title != "" {
		articleFromReq.Slug = translit.Ru(articleFromReq.Title)
	}

	err = ah.Storage.Update(articleFromReq, userID)
	if err != nil {
		if err == ah.Storage.GetErrNoUpdate() {
			utils.SendErrMessage(w, r, "no article data to update", http.StatusBadRequest)
			return
		}
		log.Printf("update article data error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	article, err := ah.Storage.GetArticleWithID(articleFromReq.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.SendErrMessage(w, r, "bad id, nothing to update", http.StatusBadRequest)
			return
		}
		log.Printf("get updated article error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := utils.Response{
		"article": article,
	}

	utils.SendResponse(w, r, response)

}

func (ah *ArticleHandler) Delete(w http.ResponseWriter, r *http.Request) {

	userID, err := ah.SessionManager.IdFromSessionContext(r)
	if err != nil {
		log.Printf("get user id error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := utils.ReadBody(w, r)
	if body == nil {
		return
	}

	articleFromReq := unmarshalBody(w, r, body)
	if articleFromReq == nil {
		return
	}

	if articleFromReq.ID == 0 {
		utils.SendErrMessage(w, r, "article id must be not empty", http.StatusBadRequest)
		return
	}

	err = ah.Storage.Delete(articleFromReq.ID, userID)
	if err != nil {
		log.Printf("delete article error: [%s], path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
