package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("GO_ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: .env file not found")
		}
	}

	InitiateAuth()

	i := InertiaStart()

	mux := http.NewServeMux()

	mux.Handle("/", i.Middleware(homeRoute(i)))
	mux.Handle("/dashboard", i.Middleware(dashboardRoute(i)))
	mux.Handle("/auth", i.Middleware(setupRoute(i)))
	mux.Handle("/auth/{provider}/callback", i.Middleware(loginRoute(i)))
	mux.Handle("/logout", i.Middleware(logoutRoute(i)))
	mux.Handle("/porkbun", i.Middleware(porkbunRoute(i)))
	mux.Handle("/data", dataRoute())
	mux.Handle("/domains", domainsRoute())
	mux.HandleFunc("/favicon.ico", serveFavicon)
	mux.Handle("/build/", http.StripPrefix("/build/", http.FileServer(http.Dir("./public/build"))))

	http.ListenAndServe(":3000", mux)
}
