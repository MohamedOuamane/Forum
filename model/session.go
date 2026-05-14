package forum

import "time"

type Session struct {
	Id         int
	User       User
	Token      string
	Expiry     time.Time
	Connection int
}
