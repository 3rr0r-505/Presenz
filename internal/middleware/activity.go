// internal/middleware/activity.go

package middleware

import (
	"net/http"

	"github.com/3rr0r-505/Presenz/internal/services"
)

// Activity wraps a handler, updating the killswitch's last-activity
// timestamp on every HTTP request before passing through.
func Activity(ks *services.Killswitch) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ks.UpdateActivity()
			next.ServeHTTP(w, r)
		})
	}
}
