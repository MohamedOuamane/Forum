package forum

import (
	"database/sql"
	model "forum/model"
	"log"
	"time"
)

type CommentService struct {
	DB          *sql.DB
	LikeService *LikeService
	UserService *UserService
}

func NewCommentService(db *sql.DB, likeService *LikeService, userService *UserService) *CommentService {
	return &CommentService{
		DB:          db,
		LikeService: likeService,
		UserService: userService,
	}
}
func (C *CommentService) CreateComment(user_id int, post_id int, content string) error {

	query := `INSERT INTO Comments(user_id, post_id,content, created_at)VALUES(?, ?, ?, ?)`
	_, err := C.DB.Exec(query, user_id, post_id, content, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
func (C *CommentService) GetCommentsPost(post_id int) ([]model.Comment, error) {
	var comments []model.Comment
	rows, err := C.DB.Query("SELECT * FROM Comments WHERE post_id=? ORDER BY Comments.created_at DESC", post_id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(&comment.Id, &comment.User.Id, &comment.Post.Id, &comment.Content, &comment.Creation)
		if err != nil {
			return nil, err
		}
		comment.User, err = C.UserService.GetUser(comment.User.Id)
		if err != nil {
			return nil, err
		}
		likes, dislikes, err := C.LikeService.GetLikeNbrComment(comment.Id)
		if err != nil {
			return nil, err
		}
		comment.Likes = likes
		comment.Dislikes = dislikes
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return comments, nil
}

func (c *CommentService) GetCommentsUser(userID int) ([]model.Comment, error) {
	var comments []model.Comment
	rows, err := c.DB.Query("SELECT id, user_id, post_id, content, created_at FROM Comments WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment model.Comment
		var postID int
		if err := rows.Scan(&comment.Id, &comment.User.Id, &postID, &comment.Content, &comment.Creation); err != nil {
			return nil, err
		}
		comment.Post.Id = postID
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

func (s *CommentService) DeleteComment(commentID int, userID int) error {
	queryCom := "DELETE FROM Likes WHERE comment_id = ? "
	_, err2 := s.DB.Exec(queryCom, commentID)
	if err2 != nil {
		return err2
	}

	query := "DELETE FROM Comments WHERE Comments.id = ? AND Comments.user_id = ?"
	_, err := s.DB.Exec(query, commentID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *CommentService) EditComment(NewContent string, commentID int, userID int) error {

	query := "UPDATE Comments SET content = ? WHERE Comments.id = ? AND Comments.user_id = ?"
	_, err := s.DB.Exec(query, NewContent, commentID, userID)
	if err != nil {
		return err
	}
	return nil
}
