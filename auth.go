package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

const (
	LoginCookieName = "user-cookie"
	MaxAge          = int(time.Hour * 24 * 30 / time.Second)
)

func InitiateAuth() {
	key := os.Getenv("AUTH_SECRET")
	productionMode := os.Getenv("GO_ENV") == "production"
	authURL := os.Getenv("AUTH_URL")
	callbackURL := fmt.Sprintf("%s/auth/%s/callback", authURL, "github")
	githubClientId := os.Getenv("GITHUB_ID")
	githubClientSecret := os.Getenv("GITHUB_SECRET")
	store := sessions.NewCookieStore([]byte(key))
	fmt.Println("MaxAge", MaxAge)
	store.Options.MaxAge = MaxAge
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = productionMode
	store.Options.SameSite = http.SameSiteLaxMode
	gothic.Store = store
	goth.UseProviders(github.New(githubClientId, githubClientSecret, callbackURL))
}

func setCookie(w http.ResponseWriter, value string) {
	domain := os.Getenv("AUTH_URL")
	productionMode := os.Getenv("GO_ENV") == "production"
	cookie := &http.Cookie{
		Name:     LoginCookieName,
		Value:    value,
		Path:     "/",
		Domain:   domain,
		MaxAge:   MaxAge,
		Secure:   productionMode,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func getCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(LoginCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func deleteCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     LoginCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Now().Add(-24 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}
