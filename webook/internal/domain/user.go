package domain

import "time"

// User 领域对象, DDD中的entity
type User struct {
	Id       int64
	Email    string
	Password string
	NickName string
	AboutMe  string
	Birthday time.Time
	Ctime    time.Time
}
