package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/markbates/goth/gothic"
	inertia "github.com/romsar/gonertia"
)

type Data struct {
	Message string `json:"message"`
}

func logoutRoute(i *inertia.Inertia) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gothic.Logout(w, r)
		deleteCookie(w)
		i.Location(w, r, "/")
		return
	})
}

func loginRoute(i *inertia.Inertia) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		person, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			fmt.Println(w, err)
			return
		}
		username := person.RawData["login"].(string)
		setCookie(w, username)
		i.Location(w, r, "/dashboard")
	})
}

func setupRoute(i *inertia.Inertia) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if person, err := gothic.CompleteUserAuth(w, r); err == nil {
			pageErr := i.Render(w, r, "AuthPage", inertia.Props{"data": person})
			if pageErr != nil {
				handleServerErr(w, pageErr)
				return
			}
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	})
}

func dashboardRoute(i *inertia.Inertia) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated(i, w, r)
		pageErr := i.Render(w, r, "Dashboard", inertia.Props{"url": os.Getenv("VITE_BASE_URL")})
		if pageErr != nil {
			handleServerErr(w, pageErr)
			return
		}
	})
}

type InputData struct {
	SubDomain   string `json:"subdomain"`
	TTL         int    `json:"ttl"`
	Kind        string `json:"kind"`
	Prefix      string `json:"prefix"`
	Host        string `json:"host"`
	Description string `json:"description"`
}

func validatePostWithJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
}

func fromJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var data T
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return data, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &data)

	if err != nil {
		http.Error(w, "Error parsing JSON: "+err.Error(), http.StatusBadRequest)
		return data, err
	}
	return data, nil
}

func dataRoute(i *inertia.Inertia) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated(i, w, r)
		validatePostWithJSON(w, r)
		data, err := fromJSON[InputData](w, r)
		if err != nil {
			return
		}
		fmt.Printf("Received data: %+v\n", data)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Data{
			Message: "successfully sent!",
		})
	})
}

type DomainData struct {
	Kind string `json:"kind"`
}

type DomainsResponse struct {
	Domains []DomainInfo `json:"domains"`
}

type DomainInfo struct {
	Tag  string `json:"tag"`
	Name string `json:"name"`
	Date string `json:"date"`
}

type ApiResponse struct {
	Status  string          `json:"status"`
	Domains []PorkbunDomain `json:"domains"`
}

type PorkbunDomain struct {
	Domain       string `json:"domain"`
	Status       string `json:"status"`
	TLD          string `json:"tld"`
	CreateDate   string `json:"createDate"`
	ExpireDate   string `json:"expireDate"`
	SecurityLock string `json:"securityLock"`
	WhoisPrivacy string `json:"whoisPrivacy"`
	AutoRenew    int    `json:"autoRenew"`
	NotLocal     int    `json:"notLocal"`
}

type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error: %s (code: %d), body : %s", e.Status, e.StatusCode, e.Body)
}

func fromResponse[T any](res *http.Response) (T, error) {
	var result T
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return result, fmt.Errorf("error reading response body: %w", err)
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("error parsing JSON response : %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return result, &HTTPError{
			StatusCode: res.StatusCode,
			Status:     res.Status,
			Body:       string(body),
		}
	}
	return result, nil
}

func processDomains(data ApiResponse) []DomainInfo {
	result := []DomainInfo{}
	for _, item := range data.Domains {
		msg := DomainInfo{Date: item.ExpireDate, Name: item.Domain, Tag: item.Status}
		result = append(result, msg)
	}
	return result
}

func domainsRoute(i *inertia.Inertia) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated(i, w, r)
		validatePostWithJSON(w, r)
		_, err := fromJSON[DomainData](w, r)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var (
			APIKEY    = os.Getenv("DOMAIN_API_KEY")
			APISECRET = os.Getenv("DOMAIN_API_SECRET")
		)
		listAllDomainsURL := "https://api.porkbun.com/api/json/v3/domain/listAll"
		req := map[string]string{
			"secretapikey":  APISECRET,
			"apikey":        APIKEY,
			"start":         "1",
			"includeLabels": "no",
		}
		sendReqJSON, _ := json.Marshal(req)
		res, err := http.Post(listAllDomainsURL, "application/json", bytes.NewBuffer(sendReqJSON))
		domains, err := fromResponse[ApiResponse](res)
		if err != nil {
			return
		}
		json.NewEncoder(w).Encode(DomainsResponse{
			Domains: processDomains(domains),
		})
	})
}

type User struct {
	Name     string `json:"name"`
	LoggedIn bool   `json:"logged_in"`
}

func homeRoute(i *inertia.Inertia) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, cookieErr := getCookie(r)
		u := User{
			Name:     cookie,
			LoggedIn: cookieErr == nil && cookie != "",
		}
		fmt.Println(cookie)
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		err := i.Render(w, r, "Index", inertia.Props{"user": u})
		if err != nil {
			handleServerErr(w, err)
			return
		}
	})

}

func porkbunRoute(i *inertia.Inertia) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated(i, w, r)
		err := i.Render(w, r, "Porkbun")
		if err != nil {
			handleServerErr(w, err)
			return
		}
	})

}

func serveFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	http.ServeFile(w, r, "./public/favicon.ico")
}

func handleServerErr(w http.ResponseWriter, err error) {
	log.Printf("http error: %s\n", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("server error"))
}

func authenticated(i *inertia.Inertia, w http.ResponseWriter, r *http.Request) {
	admin := os.Getenv("ADMIN_NAME")
	cookie, err := getCookie(r)
	if err != nil {
		i.Location(w, r, "/")
		return
	}
	if cookie != admin {
		i.Location(w, r, "/")
		return
	}
}
