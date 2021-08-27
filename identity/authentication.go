package identity

import (
	"net/http"

	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// Auditor is an alias for the auditor service
type Auditor audit.AuditorService

// Check wraps a HTTP handler. If authentication fails an error code is returned else the HTTP handler is called
func Check(auditor Auditor, action string, handle func(http.ResponseWriter, *http.Request)) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		auditParams := audit.GetParameters(ctx, r.URL.EscapedPath(), vars)
		logData := audit.ToLogData(auditParams)

		log.Info(ctx, "checking for an identity in request context", log.HTTP(r, 0, 0, nil, nil), logData)

		if err := auditor.Record(ctx, action, audit.Attempted, auditParams); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			request.DrainBody(r)
			return
		}

		// just checking if an identity exists until permissions are being provided.
		if !common.IsCallerPresent(ctx) {
			log.Info(ctx, "no identity found in context of request", log.HTTP(r, 0, 0, nil, nil), logData)

			if auditErr := auditor.Record(ctx, action, audit.Unsuccessful, auditParams); auditErr != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				request.DrainBody(r)
				return
			}

			http.Error(w, "unauthenticated request", http.StatusUnauthorized)
			request.DrainBody(r)
			return
		}

		log.Info(ctx, "identity found in request context, calling downstream handler", log.HTTP(r, 0, 0, nil, nil), logData)

		// The request has been authenticated, now run the clients request
		handle(w, r)
	})
}
