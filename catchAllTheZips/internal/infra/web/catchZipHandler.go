package web

import (
	"encoding/json"
	"io"
	"net/http"

	dtos "github.com/GoExpertCurso/catchAllTheZips/internal/entity/DTOs"
	identifyZipCode "github.com/GoExpertCurso/catchAllTheZips/pkg/identifyZipCode"
)

func catchZipHandler(w http.ResponseWriter, r *http.Request) {
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

	if identifyZipCode.IdentifyZipCode(cepDTO.Cep) {
		w.Write([]byte("Valid zip code"))
	} else {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
	}

}
