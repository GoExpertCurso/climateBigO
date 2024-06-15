package web

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	dtos "github.com/GoExpertCurso/catchAllTheZips/internal/entity/DTOs"
	"github.com/GoExpertCurso/catchAllTheZips/pkg"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

type WebHandler struct {
	Tracer trace.Tracer
}

func NewWebHandler(tracer trace.Tracer) *WebHandler {
	return &WebHandler{
		Tracer: tracer,
	}
}

func (h *WebHandler) CatchZipHandler(w http.ResponseWriter, r *http.Request) {
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
		w.Write([]byte(callTemperatureAPI(cepDTO.Cep)))
	} else {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
	}
}

func callTemperatureAPI(cep string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+os.Getenv("HOST_WTT")+":"+os.Getenv("PORT_WTT")+"/"+cep, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport,
			otelhttp.WithSpanNameFormatter(func(_ string, req *http.Request) string {
				return "get-cep-temp"
			}),
		),
	}
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
