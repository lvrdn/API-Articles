package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/mdigger/translit"
)

type StorageDBA struct {
	db *sql.DB
}

func NewStorageDBA(db *sql.DB) *StorageDBA {
	return &StorageDBA{
		db: db,
	}
}

func (st *StorageDBA) GetErrNoUpdate() error {
	return fmt.Errorf("no data to update")
}

func (st *StorageDBA) GetAuthorWithID(id string) (string, string, error) {
	var author, bio string

	err := st.db.QueryRow("SELECT username, bio FROM users WHERE id = $1", id).Scan(&author, &bio)
	if err != nil {
		return "", "", err
	}
	fmt.Println("AuthoR:", author)
	return author, bio, nil
}

func (st *StorageDBA) Add(new *Article) error {
	var LastInsertId int
	new.Slug = translit.Ru(new.Title)
	err := st.db.QueryRow(`INSERT INTO 
	articles(user_id,body,description,favorited,favorites_count,slug,title,tag_list,created_at,updated_at) 
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) 
	RETURNING id`,
		new.Author.ID, new.Body, new.Description, new.Favorited, new.FavoritesCount, new.Slug, new.Title, pq.Array(new.TagList), new.CreatedAt, new.UpdatedAt,
	).Scan(&LastInsertId)

	if err != nil {
		return err
	}

	if LastInsertId == 0 {
		return fmt.Errorf("no last insert id")
	}

	return nil
}

func (st *StorageDBA) Update(article *Article, userID string) error {

	query := "UPDATE articles SET "
	placeholderNum := 1
	args := make([]interface{}, 0)

	if article.Body != "" {
		query += fmt.Sprintf("body = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, article.Body)
	}

	if article.Description != "" {
		query += fmt.Sprintf("description = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, article.Description)
	}

	if article.Title != "" {
		query += fmt.Sprintf("title = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, article.Title)
	}

	if len(article.TagList) != 0 {
		query += fmt.Sprintf("tag_list = $%v, ", placeholderNum)
		placeholderNum++
		args = append(args, article.TagList)
	}

	if placeholderNum == 1 {
		return st.GetErrNoUpdate()
	}
	article.UpdatedAt = time.Now()
	query += fmt.Sprintf("updated_at = $%v WHERE slug = $%v and user_id = $%v", placeholderNum, placeholderNum+1, placeholderNum+2)

	args = append(args, article.UpdatedAt, article.Slug, userID)

	_, err := st.db.Exec(query, args...)

	if err != nil {
		return err
	}
	return nil
}

func (st *StorageDBA) Delete(slug, userID string) error {
	_, err := st.db.Exec("DELETE FROM articles WHERE slug = $1 and user_id = $2", slug, userID)
	if err != nil {
		return err
	}
	return nil
}

func (st *StorageDBA) Get(m map[string]string) ([]*Article, error) {
	articles := []*Article{}
	query := "SELECT users.username, users.bio, articles.* FROM users JOIN articles ON users.id = articles.user_id"

	rows := &sql.Rows{}
	var err error
	if author, ok := m["author"]; ok {
		query += " WHERE users.username = $1"
		rows, err = st.db.Query(query, author)

	} else if _, ok := m["tag"]; ok {
		query += " WHERE $1 = ANY(tag_list)"
		rows, err = st.db.Query(query, m["tag"])

	} else {
		rows, err = st.db.Query(query)
	}

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var username, bio, id, user_id, body, description, slug, title string
		var favorited bool
		var tagList []string
		var favoritesCount int
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&username,
			&bio,
			&id,
			&user_id,
			&body,
			&description,
			&favorited,
			&favoritesCount,
			&slug,
			&title,
			pq.Array(&tagList),
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			fmt.Println("ERROR HERE:", err)
			return nil, err
		}
		articles = append(articles, &Article{
			Author: &UserProfile{
				ID:       user_id,
				Username: username,
				Bio:      bio,
			},
			ID:             id,
			Body:           body,
			Description:    description,
			Favorited:      favorited,
			FavoritesCount: favoritesCount,
			Slug:           slug,
			Title:          title,
			TagList:        tagList,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		})
	}
	rows.Close()

	return articles, nil
}
