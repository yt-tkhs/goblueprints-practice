package main

import (
	"net/http"
	"strings"
	"log"
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
)

type authHandler struct {

	// 認証後の遷移先をラップする
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("auth"); err == http.ErrNoCookie || cookie.Value == "" {
		// Unauthorized
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		panic(err.Error())
	} else {
		// Authorized
		h.next.ServeHTTP(w, r)
	}
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	elems := strings.Split(r.URL.Path, "/")

	action := elems[2]
	providerName := elems[3]

	switch action {
	case "login":
		provider, err := gomniauth.Provider(providerName)
		if err != nil {
			log.Fatalln("Failed to load provider. :", provider, "-", err)
		}

		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln("Failed to call GetBeginAuthURL. :", provider, "-", err)
		}

		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider(providerName)
		if err != nil {
			log.Fatalln("Failed to load provider. :", provider, "-", err)
		}

		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Fatalln("Failed to complete authentication. :", provider, "-", err)
		}

		user, err := provider.GetUser(creds)
		if err != nil {
			log.Fatalln("Failed to fetch user. :", provider, "-", err)
		}

		authCookieValue := objx.New(map[string]interface{} {
			"name": user.Name(),
			"avatar_url": user.AvatarURL(),
			"email": user.Email(),
		}).MustBase64()

		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  authCookieValue,
			Path:   "/",
		})

		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(w, "This action '%s' is not supported.", action)
	}
}