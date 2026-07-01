package middleware

import (
	"encoding/json"
	"net/http"
	"regexp"

	"gateway/internal/store"
)

// userPathRe matches the user-scoped resources exposed by the demo API, e.g.
// /api/users/2 or /api/users/2/orders, capturing the target user id.
var userPathRe = regexp.MustCompile(`^/api/users/([^/]+)`)

// IDORGuard blocks a caller from reaching another user's resource. When the
// target user id in the path differs from the user id the token is scoped to,
// it raises an idor alert and answers 403.
func IDORGuard(alerts AlertStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := stateFrom(r)
		m := userPathRe.FindStringSubmatch(r.URL.Path)
		if st == nil || m == nil {
			next.ServeHTTP(w, r)
			return
		}

		target := m[1]
		if target == st.UserID {
			next.ServeHTTP(w, r)
			return
		}

		reqID := &st.RequestID
		meta, _ := json.Marshal(map[string]string{
			"token_user":  st.UserID,
			"target_user": target,
			"path":        r.URL.Path,
		})
		_ = alerts.InsertAlert(r.Context(), store.AlertInput{
			RequestID: reqID,
			SourceIP:  clientIP(r),
			AuthSub:   st.AuthSubject,
			AlertType: "idor",
			Severity:  3,
			Reason:    "attempt to access another user's resource",
			Metadata:  meta,
		})
		http.Error(w, "forbidden", http.StatusForbidden)
	})
}
