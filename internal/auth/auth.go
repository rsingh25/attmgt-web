package auth

import (
	"attmgt-web/internal/logger"
	"attmgt-web/internal/util"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/coreos/go-oidc"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
)

type ClaimsPage struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	Claims       jwt.MapClaims
}

var (
	clientID     = util.MustGetenvStr("OAUTH_CLIENT_ID")
	clientSecret = util.MustGetenvStr("OAUTH_CLIENT_SECRET")
	callbackURL  = util.MustGetenvStr("OAUTH_CALLBACK_URL")
	issuerURL    = util.MustGetenvStr("OAUTH_ISSUER_URL")
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	jwks         *keyfunc.JWKS
	log          *slog.Logger
)

func init() {
	log = logger.Logger.With("package", "server")
	var err error
	// Initialize OIDC provider
	provider, err = oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		log.Error("Failed to create OIDC provider:", "err", err)
		panic(err)
	}

	// Set up OAuth2 config
	oauth2Config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "openid", "phone"},
	}

	// Make the HTTP GET request
	jwksResp, err := http.Get(issuerURL + "/.well-known/jwks.json")

	// Check if the request was successful (status code 200 OK)
	if err != nil || jwksResp.StatusCode != http.StatusOK {
		log.Error("Error fetching jwks.json", "status", jwksResp.StatusCode)
		panic("Error fetching jwks.json")
	}
	defer jwksResp.Body.Close()

	// Read the response body
	jwksBytes, err := io.ReadAll(jwksResp.Body)
	if err != nil {
		log.Error("Error reading jwks.json", "err", err)
		panic("Error reading jwks.json")
	}

	// Create the JWKS from the resource at the given URL.
	jwks, err = keyfunc.NewJSON(jwksBytes)
	if err != nil {
		log.Error("Error create keyfunc using jwks", "err", err)
		panic("Error create keyfunc using jwks")
	}

}

func RegisterAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/home", handleHome)
	mux.HandleFunc("/auth/login", handleLogin)
	mux.HandleFunc("/auth/logout", handleLogout)
	mux.HandleFunc("/auth/callback", handleCallback)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `
        <html>
        <body>
            <h1>Welcome to Cognito OIDC Go App</h1>
            <a href="/auth/login">Login with Cognito</a>
        </body>
        </html>`
	fmt.Fprint(w, html)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	expireAccessCookie(w)

	state := generateStateOauthCookie(w)

	url := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {

	stateCookie, _ := r.Cookie("oauthstate")
	state := r.FormValue("state")
	if state != stateCookie.Value {
		log.Error("invalid oauth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	// Exchange the authorization code for a token
	rawToken, err := oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	idToken, ok := rawToken.Extra("id_token").(string)
	if !ok {
		log.Error("ID Token not found in rawtoken")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Parse the token (do signature verification for your use case in production)
	token, _, err := new(jwt.Parser).ParseUnverified(rawToken.AccessToken, jwt.MapClaims{})
	if err != nil {
		log.Error("Error parsing token", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Check if the token is valid and extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid claims", http.StatusBadRequest)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	//https://www.alexedwards.net/blog/working-with-cookies-in-go
	//Set the cookie
	accessTokenCookie := http.Cookie{
		Name:     "access_token",
		Value:    idToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	expiteTs, expExist := getExpireTs(claims)
	if expExist {
		accessTokenCookie.Expires = expiteTs.Time
	}

	http.SetCookie(w, &accessTokenCookie)

	// Prepare data for rendering the template
	/*pageData := ClaimsPage{
			AccessToken:  rawToken.AccessToken,
			RefreshToken: rawToken.RefreshToken,
			IDToken:      idToken,
			Claims:       claims,
		}

		// Define the HTML template
		tmpl := `
	    <html>
	        <body>
			    <a href="https://us-east-1.console.aws.amazon.com/console/logout!doLogout">Logout</a>
	            <h1>User Information</h1>
	            <h1>JWT Claims</h1>
	            <p><strong>Access Token:</strong> {{.AccessToken}}</p>
	            <p><strong>Refresh Token:</strong> {{.RefreshToken}}</p>
	            <p><strong>Id Token:</strong> {{.IDToken}}</p>
	            <ul>
	                {{range $key, $value := .Claims}}
	                    <li><strong>{{$key}}:</strong> {{$value}}</li>
	                {{end}}
	            </ul>
	            <a href="/auth/logout">Logout</a>
	        </body>
	    </html>`

		// Parse and execute the template
		t := template.Must(template.New("claims").Parse(tmpl))
		t.Execute(w, pageData)
	*/

	redirectCookie, err := r.Cookie("oauth_redirect_after_login")
	if err == nil {
		redirectURL := redirectCookie.Value
		// Clear the cookie to prevent future unintended redirects
		setRedirectCookie(w, "", -1)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	// Here, you would clear the session or cookie if stored.
	expireAccessCookie(w)
	http.Redirect(w, r, "/home", http.StatusFound)
}

// Auth middleware
func WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := r.Cookie("access_token")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				setRedirectCookie(w, r.RequestURI, 300)
				http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			default:
				setRedirectCookie(w, r.RequestURI, 300)
				http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			}
			return
		}

		// Parse the token (do signature verification for your use case in production)
		token, err := jwt.Parse(accessToken.Value, jwks.Keyfunc) //TODO , jwt.WithValidMethods
		if err != nil {
			log.Error("Failed to parse JWT", "err", err)
			setRedirectCookie(w, r.RequestURI, 300)
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}

		if !token.Valid {
			log.Error("Failed to parse JWT", "err", err)
			setRedirectCookie(w, r.RequestURI, 300)
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}

		//TODO Validate and match with path TODO
		//TODO update request context with email and role
		next.ServeHTTP(w, r)
	})
}

func setRedirectCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_redirect_after_login",
		Value:    value, //r.URL.Path, // Or r.RequestURI for full path with query
		Path:     "/",
		HttpOnly: true,
		MaxAge:   maxAge, //300 or 5 minutes
	})
}

func expireAccessCookie(w http.ResponseWriter) {
	//Delete the cookie
	accessTokenCookie := http.Cookie{
		Name:   "id_token",
		Value:  "",
		MaxAge: -1,
	}
	//cookie.Expires = expirationTime.Time
	http.SetCookie(w, &accessTokenCookie)
}

func getExpireTs(claims jwt.MapClaims) (*jwt.NumericDate, bool) {
	v, ok := claims["exp"]
	if ok {
		switch exp := v.(type) {
		case float64:
			if exp == 0 {
				return nil, false
			} else {
				return newNumericDateFromSeconds(exp), true
			}
		case json.Number:
			v, _ := exp.Float64()
			return newNumericDateFromSeconds(v), true
		}
	}
	return nil, false
}

// newNumericDateFromSeconds creates a new *NumericDate out of a float64 representing a
// UNIX epoch with the float fraction representing non-integer seconds.
func newNumericDateFromSeconds(f float64) *jwt.NumericDate {
	round, frac := math.Modf(f)
	return jwt.NewNumericDate(time.Unix(int64(round), int64(frac*1e9)))
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}
