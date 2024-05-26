package main

import (
	"log"
	"net/http"

	"github.com/GoExpertCurso/catchAllTheZips/internal/infra/web"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", web.CatchZipHandler)

	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", r)
}
