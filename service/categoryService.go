package forum

import (
	"database/sql"
	model "forum/model"
	"log"
)

type CategoryService struct {
	DB *sql.DB
}

func NewCategoryService(db *sql.DB) *CategoryService {
	return &CategoryService{
		DB: db,
	}
}

func (s *CategoryService) GetCategories() ([]model.Category, error) {
	var categories []model.Category

	rows, err := s.DB.Query("SELECT * FROM Categories")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c model.Category
		var logo sql.NullString
		err := rows.Scan(&c.Id, &c.Name, &logo)
		if err != nil {
			log.Fatal(err)
		}

		if logo.Valid {
			c.Logo = logo.String
		} else {
			c.Logo = ""
		}

		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return categories, nil
}
func (s *CategoryService) GetCategoryPost(post_id int) ([]model.Category, error) {
	var categories []model.Category

	rows, err := s.DB.Query(`
		SELECT c.id, c.name, c.logo
		FROM Categories c
		JOIN PostCategories pc ON pc.category_id = c.id
		WHERE pc.post_id = ?
	`, post_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c model.Category
		var logo sql.NullString

		if err := rows.Scan(&c.Id, &c.Name, &logo); err != nil {
			return nil, err
		}

		if logo.Valid {
			c.Logo = logo.String
		}

		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
