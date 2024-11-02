package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/markbates/goth/gothic"
	gonanoid "github.com/matoous/go-nanoid/v2"
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
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}
	var domains []string
	db.Select(&domains, `SELECT name from owned_domains ORDER BY name DESC`)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated(i, w, r)
		pageErr := i.Render(w, r, "Dashboard", inertia.Props{"url": os.Getenv("VITE_BASE_URL"), "domains": domains})
		if pageErr != nil {
			handleServerErr(w, pageErr)
			return
		}
	})
}

type InputData struct {
	Domain      string `json:"domain"`
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
		var (
			APIKEY    = os.Getenv("DOMAIN_API_KEY")
			APISECRET = os.Getenv("DOMAIN_API_SECRET")
		)
		createRecord := fmt.Sprintf("https://api.porkbun.com/api/json/v3/dns/create/%s", data.Domain)
		secs := 60
		ttl := data.TTL * secs
		req := map[string]string{
			"secretapikey": APISECRET,
			"apikey":       APIKEY,
			"type":         data.Kind,
			"content":      data.Host,
			"ttl":          fmt.Sprintf("%d", ttl),
		}
		if data.Prefix != "*" {
			req["name"] = data.Prefix
		}
		reqJSON, _ := json.Marshal(req)
		res, err := http.Post(createRecord, "application/json", bytes.NewBuffer(reqJSON))
		fmt.Println(res)
		if err == nil {
			db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
			if err != nil {
				log.Fatalln(err)
			}
			recordId, _ := gonanoid.New()
			tx := db.MustBegin()
			tx.MustExec(`INSERT INTO domain ("id","identifier","requestType","baseDomain","targetHost","TTL","description") VALUES ($1,$2,$3,$4,$5,$6,$7)`, recordId, data.Prefix, data.Kind, data.Domain, data.Host, ttl, data.Description)
			tx.Commit()
		}
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
	Tag    string `json:"tag"`
	Name   string `json:"name"`
	Target string `json:"target"`
	Date   string `json:"date"`
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
	AutoRenew    any    `json:"autoRenew"`
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

type OwnedDomains struct {
	Status     string    `db:"status" json:"status"`
	Name       string    `db:"name" json:"name"`
	ExpiresAt  time.Time `db:"expires_at" json:"expires_at"`
	ObtainedAt time.Time `db:"obtained_at" json:"obtained_at"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

func insertIntoCreatedDB(domains []PorkbunDomain) {
	table := `
	CREATE TABLE IF NOT EXISTS owned_domains (
		status VARCHAR(100) NOT NULL,
		name VARCHAR(100) UNIQUE NOT NULL,
		expires_at TIMESTAMP,
		obtained_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (name)
	)
	`
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(table)
	tx := db.MustBegin()
	for _, domain := range domains {
		expired_at, dateErr := time.Parse("2006-01-02 15:04:05", domain.ExpireDate)
		if dateErr != nil {
			fmt.Println(dateErr)
		}
		created_at, cdateErr := time.Parse("2006-01-02 15:04:05", domain.CreateDate)
		if cdateErr != nil {
			fmt.Println(dateErr)
		}
		tx.NamedExec("INSERT INTO owned_domains (status,name,expires_at,obtained_at) VALUES (:status,:name,:expires_at,:obtained_at) ON CONFLICT (name) DO UPDATE SET status = EXCLUDED.status", &OwnedDomains{Status: domain.Status, Name: domain.Domain, ExpiresAt: expired_at, ObtainedAt: created_at})
	}
	tx.Commit()
}

type DBRowDomain struct {
	Identifier  string    `db:"identifier"`
	RequestType string    `db:"requestType"`
	BaseDomain  string    `db:"baseDomain"`
	TargetHost  string    `db:"targetHost"`
	CreatedAt   time.Time `db:"createdAt"`
}

func processDBResults() []DomainInfo {
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}
	row := []DBRowDomain{}
	db.Select(&row, `SELECT "identifier","requestType", "baseDomain", "targetHost", "createdAt" FROM domain;`)
	result := []DomainInfo{}
	for _, item := range row {
		msg := DomainInfo{Date: item.CreatedAt.Format("2006.01.02"), Name: fmt.Sprintf("%s.%s", item.Identifier, item.BaseDomain), Tag: item.RequestType, Target: item.TargetHost}
		result = append(result, msg)
	}
	return result
}

func processDomains(data ApiResponse) []DomainInfo {
	result := []DomainInfo{}
	for _, item := range data.Domains {
		msg := DomainInfo{Date: item.ExpireDate, Name: item.Domain, Tag: item.Status}
		result = append(result, msg)
	}
	insertIntoCreatedDB(data.Domains)
	return result
}

func domainsRoute(i *inertia.Inertia) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated(i, w, r)
		validatePostWithJSON(w, r)
		choice, err := fromJSON[DomainData](w, r)
		if err != nil {
			return
		}
		var domainInfo []DomainInfo
		w.Header().Set("Content-Type", "application/json")
		switch choice.Kind {
		case "fetch-listing":
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
				fmt.Println(err)
				return
			}
			domainInfo = processDomains(domains)
		default:
			domainInfo = processDBResults()
		}
		json.NewEncoder(w).Encode(DomainsResponse{
			Domains: domainInfo,
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
		db, dbErr := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
		if dbErr != nil {
			log.Fatalln(dbErr)
		}
		rows := []OwnedDomains{}
		db.Select(&rows, `SELECT status, name, expires_at FROM owned_domains ORDER BY expires_at ASC;`)
		err := i.Render(w, r, "Porkbun", inertia.Props{"domains": rows})
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
