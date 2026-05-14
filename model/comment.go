package forum

import "time"

type Comment struct {
	Id       int
	User     User
	Post     Post
	Content  string
	Creation time.Time
	Likes    int
	Dislikes int
	UserLike int
}

type CommentView struct {
	Comment Comment
	UserId  int
}
