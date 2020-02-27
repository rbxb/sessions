package sessions

import "time"

type Session struct {
	OnEndFunc
	Value    interface{}
	num      uint32
	auth     string
	lifetime time.Duration
	expires  time.Time
}
