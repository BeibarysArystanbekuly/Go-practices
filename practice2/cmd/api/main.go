package main

import (
	"log"
	"net/http"

	"github.com/DayFay1/Go-practices/Practice2/internal/handlers"
	"github.com/DayFay1/Go-practices/Practice2/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	user := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetUser(w, r)
		case http.MethodPost:
			handlers.PostUser(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.Handle("/user", middleware.APIKeyAuth(user))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
