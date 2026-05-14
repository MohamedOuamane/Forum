package forum

import "time"

type Post struct {
	Id       int
	User     User
	Title    string
	Content  string
	Creation time.Time
	Category []Category
	Images   []string
	Like     int
	Dislike  int
}

type PostView struct {
	Post             Post
	UserId           int
	ImagesString     string
	CategoriesString string
	UserLike int
}
