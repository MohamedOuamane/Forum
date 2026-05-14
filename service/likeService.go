package forum

import (
	"database/sql"
	"fmt"
	model "forum/model"
	"time"
)

type LikeService struct {
	DB          *sql.DB
	UserService *UserService
}

func NewLikeService(db *sql.DB, userService *UserService) *LikeService {
	return &LikeService{
		DB:          db,
		UserService: userService,
	}
}

// Like or dislike a post
func (s *LikeService) LikeDislikePost(value, postID, userID int) error {
	if s.IfExistsPost(postID, userID) {
		_, err := s.DB.Exec(`
			UPDATE Likes
			SET value = ? , created_at = ? 
			WHERE post_id = ? AND user_id = ? AND comment_id IS NULL
		`, value, time.Now(), postID, userID)
		return err
	} else {
		_, err := s.DB.Exec(`
			INSERT INTO Likes(user_id, post_id, value, created_at)
			VALUES (?, ?, ?, ?)
		`, userID, postID, value, time.Now())
		return err
	}
}

// Like or dislike a comment
func (s *LikeService) LikeDislikeComment(value, postID, userID, commentID int) error {
	if s.IfExistsComment(commentID, userID) {
		_, err := s.DB.Exec(`
			UPDATE Likes
			SET value = ?, created_at = ?
			WHERE user_id = ? AND comment_id = ? AND post_id IS NULL
		`, value, time.Now(), userID, commentID)
		return err
	} else {
		_, err := s.DB.Exec(`
			INSERT INTO Likes(user_id, comment_id, value, created_at)
			VALUES (?, ?, ?, ?)
		`, userID, commentID, value, time.Now())
		return err
	}
}

// Check if a like already exists for a post
func (s *LikeService) IfExistsPost(postID, userID int) bool {
	var count int
	query := "SELECT COUNT(*) FROM Likes WHERE user_id = ? AND post_id = ? AND comment_id IS NULL"
	err := s.DB.QueryRow(query, userID, postID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// Check if a like already exists for a comment
func (s *LikeService) IfExistsComment(commentID, userID int) bool {
	var count int
	query := "SELECT COUNT(*) FROM Likes WHERE user_id = ? AND comment_id = ? AND post_id IS NULL"
	err := s.DB.QueryRow(query, userID, commentID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// Get all likes for a post
func (s *LikeService) GetLikesPost(postID int) ([]model.Like, error) {
	rows, err := s.DB.Query("SELECT id, user_id, post_id, comment_id, value FROM Likes WHERE post_id = ? AND comment_id IS NULL", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []model.Like
	for rows.Next() {
		var like model.Like
		err := rows.Scan(&like.Id, &like.User.Id, &like.Post.Id, &like.Comment.Id, &like.Value)
		if err != nil {
			return nil, err
		}
		likes = append(likes, like)
	}

	// Populate user info
	for i := range likes {
		likes[i].User, err = s.UserService.GetUser(likes[i].User.Id)
		if err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return likes, nil
}

// Get all likes for a comment
func (s *LikeService) GetLikesComment(commentID int) ([]model.Like, error) {
	rows, err := s.DB.Query("SELECT id, user_id, post_id, comment_id, value FROM Likes WHERE comment_id = ? AND post_id IS NULL", commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []model.Like
	for rows.Next() {
		var like model.Like
		err := rows.Scan(&like.Id, &like.User.Id, &like.Post.Id, &like.Comment.Id, &like.Value)
		if err != nil {
			return nil, err
		}
		likes = append(likes, like)
	}

	// Populate user info
	for i := range likes {
		likes[i].User, err = s.UserService.GetUser(likes[i].User.Id)
		if err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return likes, nil
}

// Get number of likes and dislikes for a post
func (s *LikeService) GetLikeNbrPost(postID int) (likeCount, dislikeCount int, err error) {
	err = s.DB.QueryRow(`
		SELECT 
			COUNT(CASE WHEN value = 1 THEN 1 END),
			COUNT(CASE WHEN value = -1 THEN 1 END)
		FROM Likes
		WHERE post_id = ? AND comment_id IS NULL
	`, postID).Scan(&likeCount, &dislikeCount)
	return
}

// Get number of likes and dislikes for a comment
func (s *LikeService) GetLikeNbrComment(commentID int) (likeCount, dislikeCount int, err error) {
	err = s.DB.QueryRow(`
		SELECT 
			COUNT(CASE WHEN value = 1 THEN 1 END),
			COUNT(CASE WHEN value = -1 THEN 1 END)
		FROM Likes
		WHERE comment_id = ? AND post_id IS NULL
	`, commentID).Scan(&likeCount, &dislikeCount)
	return
}

func (s *LikeService) GetLikedPostsByUser(userID int) ([]model.Post, error) {
	rows, err := s.DB.Query(`
		SELECT p.id, p.title, p.content
		FROM posts p
		JOIN Likes l ON p.id = l.post_id
		WHERE l.user_id = ? AND l.value = 1 AND l.comment_id IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post

	for rows.Next() {
		var p model.Post
		err := rows.Scan(&p.Id, &p.Title, &p.Content)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func (s *LikeService) GetLikedCommentsByUser(userID int) ([]model.Comment, error) {
	rows, err := s.DB.Query(`
		SELECT c.id, c.content, p.id, p.title
		FROM comments c
		JOIN Likes l ON c.id = l.comment_id
		JOIN posts p ON c.post_id = p.id
		WHERE l.user_id = ? AND l.value = 1 AND l.post_id IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment

	for rows.Next() {
		var c model.Comment
		var post model.Post

		err := rows.Scan(&c.Id, &c.Content, &post.Id, &post.Title)
		if err != nil {
			return nil, err
		}

		c.Post = post
		comments = append(comments, c)
	}

	return comments, nil
}

func (s *LikeService) DeleteLikeDislikePost(postId int, userId int) error {
	result, err := s.DB.Exec(`
		DELETE FROM Likes WHERE user_id = ? AND post_id = ? AND comment_id IS NULL
	`, userId, postId)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no like found to delete")
	}

	return nil
}

func (s *LikeService) DeleteLikeDislikeComments(commentId int, userId int) error {
	result, err := s.DB.Exec(`
		DELETE FROM Likes WHERE user_id = ? AND comment_id = ? AND post_id IS NULL
	`, userId, commentId)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no like found to delete")
	}

	return nil
}

func (s *LikeService) GetUserLikeStatus(postID int, userID int) (int, error) {
	var value int
	err := s.DB.QueryRow(`
		SELECT value FROM Likes
		WHERE post_id = ? AND user_id = ? AND comment_id IS NULL
	`, postID, userID).Scan(&value)
	
	if err != nil {
		return 0, nil
	}
	return value, nil
}

func (s *LikeService) GetUserCommentLikeStatus(commentID int, userID int) (int, error) {
	var value int
	err := s.DB.QueryRow(`
		SELECT value FROM Likes
		WHERE comment_id = ? AND user_id = ? AND post_id IS NULL
	`, commentID, userID).Scan(&value)
	
	if err != nil {
		return 0, nil
	}
	return value, nil
}