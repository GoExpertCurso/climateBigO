package web

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	dtos "github.com/GoExpertCurso/catchAllTheZips/internal/entity/DTOs"
	"github.com/GoExpertCurso/catchAllTheZips/pkg"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/GoExpertCurso/catchAllTheZips")

func CatchZipHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "SearchZipCode")
	defer span.End()

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
		w.Write([]byte(callTemperatureAPI(cepDTO.Cep)))
	} else {
		log.Println("Oh no, zip code escaped!")
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
	}
}

func callTemperatureAPI(cep string) []byte {
	host_wtt := os.Getenv("HOST_WTT")
	port_wtt := os.Getenv("PORT_WTT")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+host_wtt+":"+port_wtt+"/"+cep, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
	}
	if res != nil {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
		}

		var tempDTO dtos.TempResponseDTO
		err = json.Unmarshal(body, &tempDTO)
		if err != nil {
			log.Printf("Error parsing response body: %v", err)
		}
		json, _ := json.Marshal(tempDTO)
		return json
	}
	return nil
}
