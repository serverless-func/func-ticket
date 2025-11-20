package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"time"
)

type searchConfig struct {
	Username string `form:"username"`
	Password string `form:"password"`
	Days     int    `form:"days"`
}

type HttpResponse struct {
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "pong")
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(failed("method not allowed"))
			return
		}
		var cfg searchConfig
		err := json.NewDecoder(r.Body).Decode(&cfg)
		defer func() {
			_ = r.Body.Close()
		}()

		if err != nil {
			w.WriteHeader(http.StatusOK) // 原代码用StatusOK，可根据规范改为400
			_ = json.NewEncoder(w).Encode(failed("missing required body"))
			return
		}

		if cfg.Days == 0 {
			cfg.Days = 30
		}

		mails, err := search(cfg)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(failed(err.Error()))
			return
		}
		var tickets []Ticket
		for _, mail := range mails {
			if len(mail) == 0 {
				continue
			}
			ts, err := parse(mail)
			if err != nil {
				continue
			} else {
				tickets = append(tickets, ts...)
			}
		}
		slices.SortFunc(tickets, func(a, b Ticket) int {
			ta, _ := parseDepartDatetime(a.DepartDate, a.DepartTime)
			tb, _ := parseDepartDatetime(b.DepartDate, b.DepartTime)
			return tb.Compare(ta)
		})
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(data(tickets))
	})

	port := os.Getenv("FC_SERVER_PORT")
	if port == "" {
		port = "9000"
	}

	log.Println("Listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func failed(msg string) HttpResponse {
	return HttpResponse{
		Msg:       msg,
		Timestamp: time.Now().Unix(),
	}
}

func data(data interface{}) HttpResponse {
	return HttpResponse{
		Msg:       "success",
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}
