package sessions

import (
	"time"
	"net/http"
	"strings"
	"strconv"
	"github.com/google/uuid"
)

type session struct {
	id int
	uuid string
	expires time.Time
	v interface{}
	next * session
}

type Store struct {
	* session
	id int
	lock chan byte
	lifetime time.Duration
}

func NewStore(lifetime time.Duration) * Store {
	return &Store{
		session: 	nil,
		id: 		0,
		lock: 		make(chan byte, 1),
		lifetime: 	lifetime,
	}
}

func(store * Store) NewSession(w http.ResponseWriter, v interface{}) {
	s := &session{
		uuid: 		uuid.New().String(),
		expires: 	time.Now().Add(store.lifetime),
		v: 			v,
	}

	<- store.lock
	s.id = store.id
	store.id++
	cur := store.session
	if cur == nil {
		store.session = s
	} else {
		for cur.next != nil {
			cur = cur.next
		}
		cur.next = s
	}
	store.lock <- 0

	cookieValue := strconv.Itoa(s.id) + "," + s.uuid
	cookie := http.Cookie{
		Name: 		"session",
		Value: 		cookieValue,
		Expires: 	s.expires,
	}
	http.SetCookie(w, &cookie)
}

func(store * Store) Get(w http.ResponseWriter, req * http.Request) interface{} {
	cookie, err := req.Cookie("session")
	if err != nil || cookie.Expires.Unix() > time.Now().Unix() {
		return nil
	}

	split := strings.Split(cookie.Value, ",")
	id, err := strconv.Atoi(split[0])
	if err != nil {
		return nil
	}
	uuid := split[1]

	<- store.lock
	var v interface{} = nil
	for s := store.session; s != nil; s = s.next {
		if s.id == id {
			if s.uuid == uuid {
				if s.expires.Unix() > time.Now().Unix() {
					s.expires.Add(10 * time.Minute)
					cookie.Expires = s.expires
					http.SetCookie(w, cookie)
					v = s.v
				} else {
					store.del(s)
				}
			}
		}
	}
	store.lock <- 0
	return v
}

func(store * Store) del(s * session) {
	if s != nil {
		<- store.lock
		c := store.session
		for c != nil && c.next != s {
			c = c.next
		}
		if c != nil {
			c.next = c.next.next
		}
		store.lock <- 0
	}
}

func(store * Store) Clean() {
	<- store.lock
	for s := store.session; s != nil; s = s.next {
		if s.expires.Unix() < time.Now().Unix() {
			store.del(s)
		}
	}
	store.lock <- 0
}