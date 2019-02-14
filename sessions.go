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
	head * session
	id int
	ok chan bool
}

func(store * Store) New(w http.ResponseWriter, v interface{}) {
	s := &session{
		id: 		store.id,
		uuid: 		uuid.New().String(),
		expires: 	time.Now().Add(10 * time.Minute),
		v: 			v,
	}
	store.id++

	cur := store.head
	if cur == nil {
		store.head = s
	} else {
		for cur.next != nil {
			cur = cur.next
		}
		cur.next = s
	}

	cookieValue := strconv.Itoa(s.id) + "," + s.uuid

	cookie := http.Cookie{
		Name: 		"session",
		Value: 		cookieValue,
		Expires: 	s.expires,
	}

	http.SetCookie(w, &cookie)
}

func(store * Store) Kill() {
	close(store.ok)
}

func(store * Store) Get(w http.ResponseWriter, req *http.Request) interface{} {
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

	for s := store.head; s != nil; s = s.next {
		if s.id == id {
			if s.uuid == uuid {
				if s.expires.Unix() > time.Now().Unix() {
					s.expires.Add(10 * time.Minute)
					cookie.Expires = s.expires
					http.SetCookie(w, cookie)
					return s.v
				} else {
					store.deleteSession(s)
					return nil
				}
				return nil
			}
		}
	}
	return nil
}

func(store * Store) deleteSession(s * session) {
	if s != nil {
		c := store.head
		for c != nil && c.next != s {
			c = c.next
		}
		if c != nil {
			c.next = c.next.next
		}
	}
}

func(store * Store) Clean() {
	for s := store.head; s != nil; s = s.next {
		if s.expires.Unix() < time.Now().Unix() {
			store.deleteSession(s)
		}
	}
}

func(store * Store) Sessions() []interface{} {
	length := 0
	for s := store.head; s != nil; s = s.next {
		length++
	}
	sessions := make([]interface{}, length)
	i := 0
	for s := store.head; s != nil; s = s.next {
		sessions[i] = s.v
		i++
	}
	return sessions
}

func(store * Store) Cleaner(delay time.Duration) {
	go func(){
		for _, ok := <- store.ok; ok; store.ok <- true {
			time.Sleep(delay)
			store.Clean()
		}
	}()
}