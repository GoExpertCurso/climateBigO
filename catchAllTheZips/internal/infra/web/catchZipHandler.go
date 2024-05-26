package web

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	dtos "github.com/GoExpertCurso/catchAllTheZips/internal/entity/DTOs"
	"github.com/GoExpertCurso/catchAllTheZips/pkg"
)

func CatchZipHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("catching the zip.....")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var cepDTO dtos.CepDTO
	err = json.Unmarshal(body, &cepDTO)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	if pkg.IdentifyZipCode(cepDTO.Cep) {
		log.Println("zip caught successfully")
		w.Write([]byte("Valid zip code"))
	} else {
		log.Println("Oh no, zip code escaped!")
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
	}
}
