package main

import "net/http"

func secureHeaderMiddleware(next http.Handler) http.Handler {
	newHandler := func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		// call the next handler
		next.ServeHTTP(w, request)
	}
	return http.HandlerFunc(newHandler)
}
