package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"rwa/pkg/article"
	"time"

	"github.com/lib/pq"
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

func (st *Storage) GetAuthorWithID(id string) (string, string, error) {
	var author, bio string

	err := st.db.QueryRow("SELECT username, bio FROM users WHERE id = $1", id).Scan(&author, &bio)
	if err != nil {
		return "", "", err
	}
	fmt.Println("AuthoR:", author)
	return author, bio, nil
}

func (st *Storage) Add(new *article.Article) (int, error) {
	var lastInsertId int

	var descriptionSQL, bodySQL sql.NullString

	if new.Description == nil {
		descriptionSQL.Valid = false
	} else {
		descriptionSQL.String = *new.Description
		descriptionSQL.Valid = true
	}

	if new.Body == nil {
		bodySQL.Valid = false
	} else {
		bodySQL.String = *new.Body
		bodySQL.Valid = true
	}

	err := st.db.QueryRow(`INSERT INTO 
	articles(user_id,title,slug,description,body,tag_list,created_at,updated_at) 
	VALUES($1,$2,$3,$4,$5,$6,$7,$8) 
	RETURNING id`,
		new.Author.ID, new.Title, new.Slug, descriptionSQL, bodySQL, pq.Array(new.TagList), new.CreatedAt, new.UpdatedAt,
	).Scan(&lastInsertId)

	if err != nil {
		return 0, err
	}

	if lastInsertId == 0 {
		return 0, fmt.Errorf("no last insert id")
	}

	return lastInsertId, nil
}

func (st *Storage) Update(article *article.Article, userID int) error {

	query := "UPDATE articles SET "
	placeholderNum := 1
	args := make([]interface{}, 0)

	if article.Body != nil {
		query += fmt.Sprintf("body = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, *article.Body)
	}

	if article.Description != nil {
		query += fmt.Sprintf("description = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, *article.Description)
	}

	if article.Title != "" {
		query += fmt.Sprintf("title = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, article.Title)
	}

	if article.Slug != "" {
		query += fmt.Sprintf("slug = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, article.Slug)
	}

	if article.TagList != nil {
		query += fmt.Sprintf("tag_list = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, pq.Array(article.TagList))
	}

	if placeholderNum == 1 {
		return st.GetErrNoUpdate()
	}
	article.UpdatedAt = time.Now()
	query += fmt.Sprintf("updated_at = $%v WHERE id = $%v and user_id = $%v", placeholderNum, placeholderNum+1, placeholderNum+2)

	args = append(args, article.UpdatedAt, article.ID, userID)

	_, err := st.db.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (st *Storage) Delete(articleID, userID int) error {
	_, err := st.db.Exec("DELETE FROM articles WHERE id = $1 and user_id = $2", articleID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (st *Storage) GetArticles(filters map[string]string) ([]*article.Article, error) {
	articles := []*article.Article{}
	query := "SELECT u.username, u.image, a.id ,a.user_id, a.title, a.slug, a.description, a.body, a.tag_list, a.created_at, a.updated_at FROM users u JOIN articles a ON u.id = a.user_id"

	rows := &sql.Rows{}
	var err error
	if author, ok := filters["author"]; ok {
		query += " WHERE users.username = $1"
		rows, err = st.db.Query(query, author)

	} else if tag, ok := filters["tag"]; ok {
		query += " WHERE $1 = ANY(tag_list)"
		rows, err = st.db.Query(query, tag)

	} else {
		rows, err = st.db.Query(query)
	}

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, userID int
		var username, slug, title string
		var bodySQL, descriptionSQL, imageSQL sql.NullString
		var tagList []string
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&username,
			&imageSQL,
			&id,
			&userID,
			&title,
			&slug,
			&descriptionSQL,
			&bodySQL,
			pq.Array(&tagList),
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		image := new(string)
		if imageSQL.Valid {
			*image = imageSQL.String
		}

		description := new(string)
		if descriptionSQL.Valid {
			*description = descriptionSQL.String
		}

		body := new(string)
		if bodySQL.Valid {
			*body = bodySQL.String
		}

		articles = append(articles, &article.Article{
			Author: &article.Author{
				ID:       userID,
				Username: username,
				Image:    *image,
			},
			ID:          id,
			Title:       title,
			Slug:        slug,
			Description: description,
			Body:        body,
			TagList:     tagList,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}
	rows.Close()

	return articles, nil
}

func (st *Storage) GetArticleWithID(id int) (*article.Article, error) {

	var userID int
	var username, slug, title string
	var bodySQL, descriptionSQL, imageSQL sql.NullString
	var tagList []string
	var createdAt, updatedAt time.Time

	err := st.db.
		QueryRow("SELECT u.username, u.image, a.user_id, a.title, a.slug, a.description, a.body, a.tag_list, a.created_at, a.updated_at FROM users u JOIN articles a ON u.id = a.user_id WHERE a.id=$1", id).
		Scan(
			&username,
			&imageSQL,
			&userID,
			&title,
			&slug,
			&descriptionSQL,
			&bodySQL,
			pq.Array(&tagList),
			&createdAt,
			&updatedAt,
		)
	if err != nil {
		return nil, err
	}

	image := new(string)
	if imageSQL.Valid {
		*image = imageSQL.String
	}

	description := new(string)
	if descriptionSQL.Valid {
		*description = descriptionSQL.String
	}

	body := new(string)
	if bodySQL.Valid {
		*body = bodySQL.String
	}

	return &article.Article{
		Author: &article.Author{
			ID:       userID,
			Username: username,
			Image:    *image,
		},
		ID:          id,
		Title:       title,
		Slug:        slug,
		Description: description,
		Body:        body,
		TagList:     tagList,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
