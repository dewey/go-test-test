package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

// apiResponse is a https://ipinfo.io response
type apiResponse struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
}

// payload is our default API response
type payload struct {
	Country string `json:"country,omitempty"`
	Error   string `json:"error,omitempty"`
}

func main() {
	client := &http.Client{}
	endpoint := "https://ipinfo.io"
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html>Example: <a href=/locate?ip=5.135.186.41>/locate?ip=5.135.186.41</a> to show country for IP</html>"))
	})
	r.Get("/locate", locateHandler(client, endpoint))
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// locateHandler is a wrapper around an geo IP request to https://ipinfo.io
func locateHandler(client *http.Client, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.URL.Query().Get("ip")
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/json", endpoint, ip), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		var pl payload
		switch resp.StatusCode {
		case http.StatusNotFound:
			pl.Error = errors.New("could not locate country info for ip").Error()
			b, err := json.Marshal(pl)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			w.Write(b)
		case http.StatusOK:
			var res apiResponse
			if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			pl.Country = res.Country
			b, err := json.Marshal(pl)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
