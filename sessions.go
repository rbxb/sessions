package sessions

import (
	"time"
	"net/http"
	"github.com/google/uuid"
)

type SessionCloser interface {
	Close()
}

type session struct {
	uuid string
	expires time.Time
	closer * SessionCloser
}

type Store struct {
	lock chan byte
	lifetime time.Duration
	sessions [] * session
}

func NewStore(lifetime time.Duration) * Store {
	return &Store{
		lock: 		make(chan byte, 1),
		lifetime: 	lifetime,
		sessions: 	make([] * session, 0),
	}
}

func(store * Store) NewSession(w http.ResponseWriter, closer * SessionCloser) {
	s := &session{
		uuid: 		uuid.New().String(),
		expires: 	time.Now().Add(store.lifetime),
		closer: 	closer,
	}

	store.lock <- 0
	store.sessions = append(store.sessions, s)
	<- store.lock

	cookie := http.Cookie{
		Name: 		"session",
		Value: 		s.uuid,
		Expires: 	s.expires,
	}
	http.SetCookie(w, &cookie)
}

func(store * Store) EndSession(closer * SessionCloser) {
	store.lock <- 0
	for i, s := range store.sessions {
		if s.closer == closer {
			(*closer).Close()
			store.sessions = append(store.sessions[:i], store.sessions[i+1:]...)
		}
	}
	<- store.lock
}

func(store * Store) Get(w http.ResponseWriter, req * http.Request) * SessionCloser {
	cookie, err := req.Cookie("session")
	if err != nil || cookie.Expires.Unix() > time.Now().Unix() {
		return nil
	}

	var closer * SessionCloser = nil
	store.lock <- 0
	for i, s := range store.sessions {
		if s.uuid == cookie.Value {
			if s.expires.Unix() > time.Now().Unix() {
				s.expires = time.Now().Add(store.lifetime)
				cookie.Expires = s.expires
				http.SetCookie(w, cookie)
				closer = s.closer
				break
			} else {
				store.sessions = append(store.sessions[:i], store.sessions[i+1:]...)
			}
		}
	}
	<- store.lock
	return closer
}

func(store * Store) Clean() {
	store.lock <- 0
	for i := 0; i < len(store.sessions); i++ {
		if store.sessions[i].expires.Unix() < time.Now().Unix() {
			(*store.sessions[i].closer).Close()
			store.sessions = append(store.sessions[:i], store.sessions[i+1:]...)
			i--
		}
	}
	<- store.lock
}

func(store * Store) SessionCount() int {
	return len(store.sessions)
}