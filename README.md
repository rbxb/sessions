# sessions

This package is for session management.

Sessions are stored in a `Store`

`func NewStore(lifetime time.Duration) * Store`
Creates a new `Store`.
Each session in this store will have a lifetime of `lifetime`.

`func(store * Store) NewSession(w http.ResponseWriter, v interface{})`
Adds a new session to `store` and sets the session cookie on the client.
The empty interface `v` should be used by the application to store session data.

`func(store * Store) Get(w http.ResponseWriter, req * http.Request) interface{}`
Uses the client's cookie to get the session from `store` and returns the `v` interface for that session.
Returns `nil` if the session is not in the store or if it's expired.
Refreshes the session lifetime in the store and in the client's cookie if it exists.

`func(store * Store) Clean()`
Removes any expired sessions from `store`.