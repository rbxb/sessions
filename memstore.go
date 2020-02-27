package sessions

import (
	"net/http"
	"time"
)

type MemStore struct {
	lock          chan byte
	a             []*Session
	cleaninterval time.Duration
	lastclean     time.Time
}

func NewMemStore() *MemStore {
	return &MemStore{
		lock:          make(chan byte, 1),
		a:             make([]*Session, 0),
		cleaninterval: time.Second * 4,
		lastclean:     time.Now(),
	}
}

func (store *MemStore) New(w http.ResponseWriter, lifetime time.Duration) *Session {
	store.lock <- 0
	store.clean(false)
	var num uint32 = 0
	if len(store.a) > 0 {
		num = store.a[len(store.a)-1].num + 1
	}
	b := make([]byte, 36)
	s := &Session{
		OnEndFunc: nil,
		Value:     nil,
		num:       num,
		auth:      newCookieAuth(b[4:]),
		lifetime:  lifetime,
		expires:   time.Now().Add(lifetime),
	}
	store.a = append(store.a, s)
	<-store.lock
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   encodeCookie(num, b),
		Expires: s.expires,
	})
	return s
}

func (store *MemStore) Get(w http.ResponseWriter, req *http.Request) *Session {
	now := time.Now()
	cookie, err := req.Cookie("session")
	if err != nil || cookie.Expires.Unix() > now.Unix() {
		return nil
	}
	num, auth, err := decodeCookie(cookie.Value)
	if err != nil {
		return nil
	}
	store.lock <- 0
	store.clean(false)
	s := store.search(num)
	if s != nil && s.expires.Unix() > now.Unix() && auth == s.auth {
		s.expires = now.Add(s.lifetime)
		cookie.Expires = s.expires
		http.SetCookie(w, cookie)
	}
	<-store.lock
	return s
}

func (store *MemStore) End(s *Session) {
	s.expires = time.Now()
	store.lock <- 0
	store.clean(false)
	<-store.lock
}

func (store *MemStore) Has(s *Session) bool {
	return store.search(s.num) == s
}

func (store *MemStore) Clean() {
	store.lock <- 0
	store.clean(true)
	<-store.lock
}

func (store *MemStore) SetCleanInterval(i time.Duration) {
	store.lock <- 0
	store.cleaninterval = i
	<-store.lock
}

func (store *MemStore) search(num uint32) *Session {
	a := store.a
	for p := len(a) / 2; len(a) > 0; p = len(a) / 2 {
		if a[p].num > num {
			a = a[:p]
		} else if a[p].num < num {
			a = a[p+1:]
		} else {
			return a[p]
		}
	}
	return nil
}

func (store *MemStore) clean(force bool) {
	now := time.Now()
	if force || store.lastclean.Add(store.cleaninterval).Unix() < now.Unix() {
		move := 0
		for i, s := range store.a {
			if s.expires.Unix() < now.Unix() {
				move++
				go s.OnEndFunc(s)
			} else if move > 0 {
				store.a[i-move] = s
			}
		}
		store.a = store.a[:len(store.a)-move]
		store.lastclean = now
	}
}
