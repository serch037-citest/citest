package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth"
	"github.com/lestrrat/go-jwx/jwk"
)

var TokenAuth *jwtauth.JWTAuth

func init() {
	set, err := jwk.ParseString(`{"keys":[{"use":"sig","kty":"RSA","kid":"a2aa9739-d753-4a0d-87ee-61f101050277","alg":"RS256","n":"zpjSl0ySsdk_YC4ZJYYV-cSznWkzndTo0lyvkYmeBkW60YHuHzXaviHqonY_DjFBdnZC0Vs_QTWmBlZvPzTp4Oni-eOetP-Ce3-B8jkGWpKFOjTLw7uwR3b3jm_mFNiz1dV_utWiweqx62Se0SyYaAXrgStU8-3P2Us7_kz5NnBVL1E7aEP40aB7nytLvPhXau-YhFmUfgykAcov0QrnNY0DH0eTcwL19UysvlKx6Uiu6mnbaFE1qx8X2m2xuLpErfiqj6wLCdCYMWdRTHiVsQMtTzSwuPuXfH7J06GTo3I1cEWN8Mb-RJxlosJA_q7hEd43yYisCO-8szX0lgCasw","e":"AQAB"}]}`)
	if err != nil {
		log.Panic(err)
	}
	public, private := set.Keys[0].Materialize()
	TokenAuth = jwtauth.New("RS256", public, private)

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	// _, tokenString, _ := tokenAuth.Encode(jwt.MapClaims{"user_id": 123})
	// fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string // Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)                             // Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host)) // Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	} // Return the request as a string
	return strings.Join(request, "\n")
}

// Middleware decodes the share session cookie and packs the session into context
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// for _, cookie := range r.Cookies() {
			// 	fmt.Fprint(w, cookie.Name)
			// 	log.Println(cookie.Name)
			// }
			// log.Println(formatRequest(r))
			// log.Println(r.Cookies())
			// c, err := r.Cookie("auth-cookie")

			// Allow unauthenticated users in
			// if err != nil || c == nil {
			// 	next.ServeHTTP(w, r)
			// 	return
			// }

			_, claims, err := jwtauth.FromContext(r.Context())
			// Allow unauthenticated users in
			if claims == nil || err != nil {
				log.Println(err)
				next.ServeHTTP(w, r)
				return
			}
			identity := claims["session"].(map[string]interface{})["identity"]
			log.Println(identity)
			// parsed, err := gabs.ParseJSON(claims)
			// log.Println(parsed)
			// log.Println(err)

			next.ServeHTTP(w, r)
		})
	}
}
