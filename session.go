package sessions

import "time"

type Session struct {
	Value interface{}
	num uint32
	auth string
	lifetime time.Duration
	expires time.Time
}