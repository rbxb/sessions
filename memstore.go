package sessions

import (
	"net/http"
	"sync"
	"time"
)

type MemStore struct {
	sync.Mutex
	a             []*Session
	cleaninterval time.Duration
	lastclean     time.Time
}

func NewMemStore() *MemStore {
	return &MemStore{
		a:             make([]*Session, 0),
		cleaninterval: time.Second * 4,
		lastclean:     time.Now(),
	}
}

func (store *MemStore) New(w http.ResponseWriter, lifetime time.Duration) *Session {
	store.Lock()
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
	store.Unlock()
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
	store.Lock()
	store.clean(false)
	s := store.search(num)
	if s != nil && s.expires.Unix() > now.Unix() && auth == s.auth {
		s.expires = now.Add(s.lifetime)
		cookie.Expires = s.expires
		http.SetCookie(w, cookie)
	}
	store.Unlock()
	return s
}

func (store *MemStore) End(s *Session) {
	s.expires = time.Now()
	store.Lock()
	store.clean(false)
	store.Unlock()
}

func (store *MemStore) Has(s *Session) bool {
	return store.search(s.num) == s
}

func (store *MemStore) Iterate(f func(*Session) bool) {
	store.Lock()
	for _, s := range store.a {
		if !f(s) {
			break
		}
	}
	store.Unlock()
}

func (store *MemStore) Clean() {
	store.Lock()
	store.clean(true)
	store.Unlock()
}

func (store *MemStore) SetCleanInterval(i time.Duration) {
	store.Lock()
	store.cleaninterval = i
	store.Unlock()
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
				onEnd := s.OnEndFunc
				if onEnd != nil {
					go onEnd(s)
				}
			} else if move > 0 {
				store.a[i-move] = s
			}
		}
		store.a = store.a[:len(store.a)-move]
		store.lastclean = now
	}
}
