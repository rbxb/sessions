package sessiongate

import (
	"net/http"

	"github.com/rbxb/httpfilter"
	"github.com/rbxb/sessions"
)

func NewSessionGate(store sessions.Store) httpfilter.OpFunc {
	return func(w http.ResponseWriter, req httpfilter.FilterRequest) string {
		if req.GetSession() == nil {
			req.SetSession(store.Get(w, req.Request))
		}
		return ""
	}
}
