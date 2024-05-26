package main

import (
	"net/http"

	"github.com/GoExpertCurso/whatsTheTemperature/configs"
	"github.com/GoExpertCurso/whatsTheTemperature/internal/web"
	"github.com/gorilla/mux"
)

func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/{city}", web.SearchZipCode)
	http.ListenAndServe(":"+configs.WEB_SERVER_PORT, r)
}
