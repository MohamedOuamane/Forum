package forum

import (
	"database/sql"
	"fmt"
	model "forum/model"
	"log"
	"strconv"
	"strings"
	"time"
)

type PostService struct {
	DB              *sql.DB
	CategoryService *CategoryService
	ImgService      *ImgService
	LikeService     *LikeService
	UserService     *UserService
}

func NewPostService(db *sql.DB, categoryService *CategoryService, imgService *ImgService, likeService *LikeService, userService *UserService) *PostService {
	return &PostService{
		DB:              db,
		CategoryService: categoryService,
		ImgService:      imgService,
		LikeService:     likeService,
		UserService:     userService,
	}
}

/*
@CreatePost creation of post
@creatPost PARAM
*/
func (s *PostService) CreatePost(user model.User, title string, content string, categories []string) (int64, error) {

	if strings.TrimSpace(content) == "" {
		return 0, fmt.Errorf("content cannot be empty")
	}

	query := `INSERT INTO Posts(user_id, title, content, created_at) VALUES (?, ?, ?, ?)`
	result, err := s.DB.Exec(query, user.Id, title, content, time.Now())
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, v := range categories {

		categoryID, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}

		query := `INSERT INTO PostCategories(post_id, category_id) VALUES (?, ?)`
		_, err = s.DB.Exec(query, id, categoryID)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}
func (s *PostService) GetPosts() ([]model.Post, error) {
	var posts []model.Post

	rows, err := s.DB.Query(`SELECT p.id, p.user_id, p.title, p.content,p.created_at,  
								SUM(
									CASE 
										WHEN l.value = 1 THEN 1
										WHEN l.value = -1 THEN -1
										ELSE 0
									END
								) AS score
							FROM posts p
							INNER JOIN Likes l on p.id = l.post_id
							WHERE l.value=1
							GROUP BY p.id
							ORDER BY p.created_at DESC
							LIMIT 3`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var p model.Post
		var score int
		err := rows.Scan(&p.Id, &p.User.Id, &p.Title, &p.Content, &p.Creation, &score)
		if err != nil {
			log.Fatal(err)
		}
		p.Images, err = s.ImgService.GetPostImages(p.Id)
		if err != nil {
			log.Fatal(err)
		}
		like, dislike, err := s.LikeService.GetLikeNbrPost(p.Id)
		if err != nil {
			log.Fatal(err)
		}
		p.Like = like
		p.Dislike = dislike
		user, err := s.UserService.GetUser(p.User.Id)
		if err != nil {
			log.Fatal(err)
		}
		p.User = user
		p.Category, err = s.CategoryService.GetCategoryPost(p.Id)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return posts, nil
}
func (s *PostService) GetPost(id int) (model.Post, error) {
	var post model.Post

	err := s.DB.QueryRow("SELECT * FROM Posts WHERE id= ?", id).Scan(&post.Id, &post.User.Id, &post.Title, &post.Content, &post.Creation)
	if err != nil {
		return post, err
	}
	post.Images, err = s.ImgService.GetPostImages(post.Id)
	if err != nil {
		log.Fatal(err)
	}
	like, dislike, err := s.LikeService.GetLikeNbrPost(post.Id)
	if err != nil {
		log.Fatal(err)
	}
	post.Like = like
	post.Dislike = dislike
	user, err := s.UserService.GetUser(post.User.Id)
	if err != nil {
		log.Fatal(err)
	}
	post.User = user
	post.Category, err = s.CategoryService.GetCategoryPost(post.Id)
	if err != nil {
		log.Fatal(err)
	}
	return post, nil
}
func (s *PostService) GetPostsCategory(category_id int) ([]int, error) {
	var posts []int

	rows, err := s.DB.Query(
		"SELECT post_id FROM PostCategories WHERE category_id = ?",
		category_id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p int

		if err := rows.Scan(&p); err != nil {
			return nil, err
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
func (s *PostService) GetPostsByUser(userID int) ([]model.Post, error) {
	var posts []model.Post

	query := `SELECT id, user_id, title, content, created_at FROM Posts WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := s.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p model.Post
		err := rows.Scan(&p.Id, &p.User.Id, &p.Title, &p.Content, &p.Creation)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *PostService) GetPostsForCategory(category_id int) ([]model.Post, error) {
	var posts []model.Post

	rows, err := s.DB.Query(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at
		FROM Posts p
		JOIN PostCategories pc ON pc.post_id = p.id
		WHERE pc.category_id = ?
		ORDER BY p.created_at DESC
	`, category_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p model.Post

		if err := rows.Scan(&p.Id, &p.User.Id, &p.Title, &p.Content, &p.Creation); err != nil {
			return nil, err
		}

		p.Images, err = s.ImgService.GetPostImages(p.Id)
		if err != nil {
			log.Fatal(err)
		}
		like, dislike, err := s.LikeService.GetLikeNbrPost(p.Id)
		if err != nil {
			log.Fatal(err)
		}
		p.Like = like
		p.Dislike = dislike

		user, err := s.UserService.GetUser(p.User.Id)
		if err != nil {
			log.Fatal(err)
		}
		p.User = user
		p.Category, err = s.CategoryService.GetCategoryPost(p.Id)
		if err != nil {
			log.Fatal(err)
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *PostService) DeletePost(postID int, userID int) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	// rollback automatique si erreur
	defer tx.Rollback()

	queryImg := "DELETE FROM Pictures WHERE post_id = ?"
	_, err6 := tx.Exec(queryImg, postID)
	if err6 != nil {
		return err6
	}

	queryRep := "DELETE FROM Reports WHERE post_id = ?"
	_, err5 := tx.Exec(queryRep, postID)
	if err5 != nil {
		return err5
	}

	// queryComLike := "DELETE FROM Likes WHERE post_id = ? "
	// _, err8 := s.DB.Exec(queryComLike, postID)
	// if err8 != nil {
	// 	return err8
	// }

	queryCom := "DELETE FROM Comments WHERE post_id = ?"
	_, err4 := tx.Exec(queryCom, postID)
	if err4 != nil {
		return err4
	}

	queryLike := "DELETE FROM Likes WHERE post_id = ?"
	_, err3 := tx.Exec(queryLike, postID)
	if err3 != nil {
		return err3
	}

	queryCat := "DELETE FROM PostCategories WHERE post_id = ?"
	_, err2 := tx.Exec(queryCat, postID)
	if err2 != nil {
		return err2
	}

	query := "DELETE FROM Posts WHERE Posts.id = ? AND Posts.user_id = ?"
	result, errPost := tx.Exec(query, postID, userID)
	if errPost != nil {
		return errPost
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("post not found")
	}

	// commit si tout OK
	return tx.Commit()
}
func (s *PostService) EditPost(postID int, content, title string, categories []string, deletedImages string) error {

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	// rollback auto if error
	defer tx.Rollback()

	// manage deleted images
	deletedSlice := strings.Split(deletedImages, ",")
	for _, img := range deletedSlice {
		queryImg := "DELETE FROM Pictures WHERE path = ?"
		_, errImg := tx.Exec(queryImg, img)
		if errImg != nil {
			return errImg
		}
	}

	// manage categories
	queryDelteCategories := "DELETE FROM PostCategories WHERE post_id = ?"
	_, errDeleteCategories := tx.Exec(queryDelteCategories, postID)
	if errDeleteCategories != nil {
		return errDeleteCategories
	}

	for _, cat := range categories {
		queryInsertCategories := "INSERT INTO PostCategories (post_id, category_id) VALUES (?,?)"
		_, errInsertCategories := tx.Exec(queryInsertCategories, postID, cat)
		if errInsertCategories != nil {
			return errInsertCategories
		}
	}

	// Update post data
	postIdStr := strconv.Itoa(postID)

	queryUpdate := "UPDATE Posts SET content = ?, title = ? WHERE id = ?"
	result, errUpdate := tx.Exec(queryUpdate, content, title, postIdStr)
	if errUpdate != nil {
		return errUpdate
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("post not found")
	}

	// commit ok
	return tx.Commit()
}
