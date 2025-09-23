package server

import (
	"attmgt-web/internal/database"
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"runtime/debug"
)

type ErrorPayload struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

type Resp[T any] struct {
	Val    T
	Err    error
	Status int
}

// Validator is an object that can be validated.
type Validator interface {
	// Valid checks the object and returns any problems. If len(problems) == 0 then the object is valid.
	Valid(ctx context.Context) (problems map[string]string)
}

//go:embed "assets"
var assets embed.FS

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	//Register auth routes
	//auth.RegisterAuthRoutes(mux)

	// secureCsrf := false
	// if s.env == "prod" {
	// 	secureCsrf = true
	// }

	mw := NewMwChainFunc(WithAlbAuth) //	auth.WithAuth,
	//WithNoSurf(secureCsrf),
	//WithTimeout(time.Second),

	// Register routes
	mux.HandleFunc("GET /hello", s.HelloWorldHandler)
	mux.Handle("GET /hello-auth", mw(s.HelloWorldHandler))
	mux.HandleFunc("GET /health", s.healthHandler)

	//mux.Handle("GET /employee", mw(Exec(s.listAuthors, s.db, false)))
	//mux.Handle("POST /author", mw(ValidateReqExec(s.createAuthor, s.db, true)))

	mux.Handle("GET /employee-web", mw(s.employeeListWeb))

	fileServer := http.FileServer(http.FS(assets))
	mux.Handle("GET /assets/", fileServer)

	return mux
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello World"}
	writeJsonResponse(w, r, http.StatusOK, resp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJsonResponse(w, r, http.StatusOK, s.db.Health())
}

func (s *Server) employeeListWeb(w http.ResponseWriter, r *http.Request) {
	employees, err := s.db.Queries().GetEmpByIds(r.Context(), []string{})
	if err != nil {
		writeInternalServerError(w, r, err)
		return
	}
	td := s.addDefaultData(&templateData{
		Data: map[string]any{
			"employees": employees,
		},
	}, r)

	if err := s.RenderTemplate(w, r, "employee-list", td); err != nil {
		writeInternalServerError(w, r, err)
		return
	}
}

/*
func (s *Server) listAuthorsWeb(w http.ResponseWriter, r *http.Request) {
	authors, err := s.db.Queries().ListAuthors(r.Context())
	if err != nil {
		writeInternalServerError(w, r, err)
		return
	}

	td := s.addDefaultData(&templateData{
		Data: map[string]any{
			"authors": authors,
		},
	}, r)

	if err := s.RenderTemplate(w, r, "author-list", td); err != nil {
		writeInternalServerError(w, r, err)
		return
	}

}

func (s *Server) listAuthors(r *http.Request, q *database.Queries) Resp[[]database.Author] {
	authors, err := q.ListAuthors(r.Context())
	if err != nil {
		return Resp[[]database.Author]{Err: err, Status: http.StatusInternalServerError}
	}

	return Resp[[]database.Author]{Val: authors, Status: http.StatusOK}
}

func (s *Server) createAuthor(c database.CreateAuthorParams, r *http.Request, qtx *database.Queries) Resp[database.Author] {
	author, err := qtx.CreateAuthor(r.Context(), c)
	if err != nil {
		return Resp[database.Author]{Err: err, Status: http.StatusInternalServerError}
	}

	return Resp[database.Author]{Val: author, Status: http.StatusOK}
}
*/

// Encoding error is not retured but handled in the function itself.
// do no leak the interval error outside
func writeInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	//Log the error that caused this
	log.Error(err.Error(), "err", err.Error(), "method", r.Method, "url", r.URL, "stack", debug.Stack())

	//Attempt to write response
	writeJsonResponse(w, r, http.StatusInternalServerError, ErrorPayload{
		Message: http.StatusText(http.StatusInternalServerError),
		Error:   true,
	})
}

// encoding error is not retured but handled in the function itself.
func writeJsonResponse[T any](w http.ResponseWriter, r *http.Request, status int, v T, headers ...http.Header) {
	if len(headers) > 0 {
		maps.Copy(w.Header(), headers[0])
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error("encode error", "error", err.Error(), "method", r.Method, "url", r.URL, "stack", debug.Stack())
	}
}

// Exec converts an error-returning handler to a standard http.HandlerFunc.
// It creates a db transaction if required and provides a query wrapper.
func Exec[RespType any](f func(*http.Request, *database.Queries) Resp[RespType], db database.Service, withTx bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var qtx *database.Queries
		var tx *sql.Tx
		var err error

		if withTx {
			tx, qtx, err = db.BeginTx(r.Context(), nil)
			if err != nil {
				writeInternalServerError(w, r, err)
				return
			}
			defer tx.Rollback()
		} else {
			qtx = db.Queries()
		}

		resp := f(r, qtx)

		if resp.Err != nil {
			//TODO Handle error types
			writeInternalServerError(w, r, resp.Err)
		} else {
			if withTx {
				tx.Commit()
			}
			writeJsonResponse(w, r, resp.Status, resp.Val)
		}
	}
}

// ValidateReqExec converts an error-returning handler to a standard http.HandlerFunc.
// It validates and decodes the request into an onject.
// It creates a db transaction if required and provides a query wrapper.
func ValidateReqExec[RespType any, ReqType Validator](f func(ReqType, *http.Request, *database.Queries) Resp[RespType], db database.Service, withTx bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, problems, err := DecodeValid[ReqType](r)

		if len(problems) > 0 {
			writeJsonResponse(w, r, http.StatusUnprocessableEntity, problems)
			return
		} else if err != nil {
			writeInternalServerError(w, r, err)
		}

		var qtx *database.Queries
		var tx *sql.Tx

		if withTx {
			tx, qtx, err = db.BeginTx(r.Context(), nil)
			if err != nil {
				writeInternalServerError(w, r, err)
				return
			}
			defer tx.Rollback()
		} else {
			qtx = db.Queries()
		}

		resp := f(body, r, qtx)

		if resp.Err != nil {
			//TODO Handle error types
			writeInternalServerError(w, r, resp.Err)
		} else {
			if withTx {
				tx.Commit()
			}
			writeJsonResponse(w, r, resp.Status, resp.Val)
		}
	}
}

func DecodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}
	if problems := v.Valid(r.Context()); len(problems) > 0 {
		return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
	}
	return v, nil, nil
}
