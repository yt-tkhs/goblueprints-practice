package main

import (
	"net/http"
	"strings"
	"log"
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
	"crypto/md5"
	"io"
	"github.com/stretchr/gomniauth/common"
)

/*
GetAvatarURL() では, ログイン時に初めてもらえる情報が必要とされた.
しかもそれは実装ごとに異なるため, 今まで通り *client を渡すだけではデータが足りない
そこで, GetAvatarURL()で必要な情報をまとめた ChatUser を作成し, そこにデータを詰める
各GetAvatarURL()は, ChatUserの中から使いたいデータだけを使う
もし, 新規に追加する GetAvatarURL() が ChatUser に入っていない情報を必要とするときは,
ChatUser に追加し、ログイン後にデータをセットして渡してあげればいい。
 */
type ChatUser interface {
	UniqueID()  string
	AvatarURL() string
}

type chatUser struct {
	/*
	 型の埋め込み (type embedding)
	 - chatUser は, common.User を実装したことになる.
	 - ChatUser にある AvatarURL() を実装しなくても, common.User.AvatarURL() が存在するので,
	 - chatUser.AvatarURL() を使ったときに, common.User.AvatarURL() が呼ばれる
	 */
	common.User
	uniqueID string
}

type authHandler struct {

	// 認証後の遷移先をラップする
	next http.Handler
}

func (u chatUser) UniqueID() string  {
	return u.uniqueID
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

		chatUser := &chatUser{
			User:       user,
			uniqueID:   ToMD5(user.Email()),
		}
		avatarURL, err := avatars.GetAvatarURL(chatUser)
		if err != nil {
			log.Fatalln("Failed to GetAvatarURL - ", err)
		}

		authCookieValue := objx.New(map[string]interface{} {
			"user_id": chatUser.uniqueID,
			"name": user.Name(),
			"avatar_url": avatarURL,
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

func ToMD5(str string) string {
	m := md5.New()
	io.WriteString(m, str)
	return fmt.Sprintf("%x", m.Sum(nil))
}