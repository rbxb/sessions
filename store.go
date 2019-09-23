package sessions

import (
	"net/http"
	"time"
)

type Store interface {
	New(http.ResponseWriter,time.Duration) * Session
	Get(http.ResponseWriter, * http.Request) * Session
	End(* Session)
}