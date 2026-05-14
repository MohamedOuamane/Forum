package forum

import "time"

type User struct {
	Id       int
	Username string
	Email    string
	Password *string
	Creation time.Time
	Role     Role
	GoogleId *string
	GithubId *string
	Avatar   *string
}
