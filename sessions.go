package sessions

import (
	"time"
	"net/http"
	"github.com/google/uuid"
)

type session struct {
	uuid string
	expires time.Time
	v * interface{}
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

func(store * Store) NewSession(w http.ResponseWriter, v * interface{}) {
	s := &session{
		uuid: 		uuid.New().String(),
		expires: 	time.Now().Add(store.lifetime),
		v: 			v,
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

func(store * Store) Get(w http.ResponseWriter, req * http.Request) * interface{} {
	cookie, err := req.Cookie("session")
	if err != nil || cookie.Expires.Unix() > time.Now().Unix() {
		return nil
	}

	var v * interface{} = nil
	store.lock <- 0
	for i, s := range store.sessions {
		if s.uuid == cookie.Value {
			if s.expires.Unix() > time.Now().Unix() {
				s.expires = time.Now().Add(store.lifetime)
				cookie.Expires = s.expires
				http.SetCookie(w, cookie)
				v = s.v
				break
			} else {
				store.sessions = append(store.sessions[:i], store.sessions[i+1:]...)
			}
		}
	}
	<- store.lock
	return v
}

func(store * Store) Clean() {
	store.lock <- 0
	for i, s := range store.sessions {
		if s.expires.Unix() < time.Now().Unix() {
			store.sessions = append(store.sessions[:i], store.sessions[i+1:]...)
		}
	}
	<- store.lock
}