package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/justinas/nosurf"
)

type Middleware func(http.Handler) http.Handler

// NewMwChain(m1, m2, m3)(myHandler) will chained as m1(m2(m3(myHandler)))
func NewMwChain(mw ...Middleware) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		next := h
		for k := len(mw) - 1; k >= 0; k-- {
			next = mw[k](next)
		}
		return next
	}
}

func NewMwChainFunc(mw ...Middleware) func(http.HandlerFunc) http.Handler {
	return func(h http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler := h
			for k := len(mw) - 1; k >= 0; k-- {
				curH := mw[k]
				nextH := handler
				// update the chain
				handler = func(w http.ResponseWriter, r *http.Request) {
					curH(nextH).ServeHTTP(w, r)
				}
			}
			// Execute the assembled processor chain
			handler.ServeHTTP(w, r)
		})
	}
}

/* TO BE TAKEN CARE BY APIGW
func WithCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}
*/

func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Debug("Http Req Started", "method", r.Method, "url", r.URL)
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Debug("Http Req Served", "method", r.Method, "url", r.URL, "duraton", duration)
	})
}

/* THIS IS NOT POSSIBLE IN LAMBDA
// ServerSentEventsLogging: This will log a message with initial http request and when response is closed.
func WithSseLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Debug("SSE Req Revieved", "method", r.Method, "url", r.URL)
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		logger.Debug("SSE Req Completed", "method", r.Method, "url", r.URL, "duration", duration)
	})
}
*/

/*
func WithTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		orgContext := r.Context()
		newContext := context.WithValue(orgContext, "time", &t)
		newRequest := r.WithContext(newContext)
		next.ServeHTTP(w, newRequest)
	})
}
*/

func WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//apiKey := r.Header.Get("Authorization")
		//if apiKey != "api-key-test" {
		//	http.Error(w, "invalid api-key", http.StatusForbidden)
		//	return
		//}
		next.ServeHTTP(w, r)
	})
}

func WithAlbAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("Alb auth", "headers", r.Header)
		//apiKey := r.Header.Get("Authorization")
		//if apiKey != "api-key-test" {
		//	http.Error(w, "invalid api-key", http.StatusForbidden)
		//	return
		//}
		next.ServeHTTP(w, r)
	})
}

// WithMsg middleware with decorators
func WithMsg(msg string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Example Message:", msg)
			next.ServeHTTP(w, r)
		})
	}
}

// NoSurf is the csrf protection middleware
func WithNoSurf(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)

		csrfHandler.SetBaseCookie(http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})
		return csrfHandler
	}
}

// This is similar to http.TimeoutHandler() but does not send a 503 with html payload
func WithTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
