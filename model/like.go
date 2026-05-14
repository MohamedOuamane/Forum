package forum

type Like struct {
	Id      int
	User    User
	Post    Post
	Comment Comment
	Value   int
}
