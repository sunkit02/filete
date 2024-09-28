package middleware

import (
	"floader/logging"
	"fmt"
	"net/http"
	"strings"
)

// NOTE: This should come after the request id middleware
func RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqId := ExtractRequestId(r).String()
		logging.Info.Println(reqLogStr(r, reqId, false))
		logging.Debug.Println(reqLogStr(r, reqId, true))
		next.ServeHTTP(w, r)
	})
}

func reqLogStr(r *http.Request, id string, headers bool) string {
	var logStr string
	var sep = ""
	if len(r.URL.RawQuery) > 0 {
		sep = "?"
	}
	path := fmt.Sprintf("%s%s%s", r.URL.Path, sep, r.URL.RawQuery)

	if headers {
		var headerSb strings.Builder
		headerSb.WriteString("Headers {")
		for k, vals := range r.Header {
			headerSb.WriteString(k)
			headerSb.WriteString("=[")
			for i, v := range vals {
				headerSb.WriteString(v)
				if i < len(vals)-1 {
					headerSb.WriteString(", ")
				}
			}
			headerSb.WriteString("], ")
		}
		headerSb.WriteRune('}')

		logStr = fmt.Sprintf("%s %s %s", r.Method, path, headerSb.String())
	} else {
		logStr = fmt.Sprintf("%s %v", r.Method, path)
	}
	return fmt.Sprintf("reqId(%s) %s", id, logStr)
}
