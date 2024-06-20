package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	dto "github.com/GoExpertCurso/whatsTheTemperature/internal/web/entity/DTOs"
	"github.com/GoExpertCurso/whatsTheTemperature/pkg"
	"github.com/go-chi/chi"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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

func (h *WebHandler) SearchZipCode(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	_, span := h.Tracer.Start(ctx, "get-city-name")

	cep := chi.URLParam(r, "cep")
	log.Printf("cep: %s", cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://viacep.com.br/ws/"+cep, nil)
	if err != nil {
		log.Printf("Fail to create the request: %v", err)
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	response, err := client.Do(req)
	if err != nil {
		log.Panic("Error: ", err)
	}

	cepRegex := regexp.MustCompile(`^\d{5}-?\d{3}$`)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err.Error())
	}

	var erroDto dto.ZipCodeError
	_ = json.Unmarshal([]byte(body), &erroDto)
	/* if err != nil {
		fmt.Println("Error decoding response body:", err.Error())
	} */

	if erroDto.Erro {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("can not find zipcode"))
		return
	}

	if !cepRegex.MatchString(cep) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("invalid zipcode"))
		return
	}

	var cepDto dto.Cep
	_ = json.Unmarshal(body, &cepDto)
	defer response.Body.Close()
	span.End()
	h.searchClimate(ctx, w, r, cepDto.Localidade)
}

func (h *WebHandler) searchClimate(ctx context.Context, w http.ResponseWriter, r *http.Request, location string) {
	_, span := h.Tracer.Start(ctx, "get-city-temp")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	defer span.End()

	params := url.Values{}
	params.Add("q", location)
	params.Add("aqi", "no")

	encodedParams := params.Encode()

	baseUrl := "https://api.weatherapi.com/v1/current.json?key=" + os.Getenv("API_KEY")
	requestUrl := baseUrl + "&" + encodedParams

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	response, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	if response.StatusCode != 200 {
		w.Write([]byte("Location not found"))
		return
	}

	weatherResponse, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln("\nError reading response body:", err.Error())
	}

	var weatherDto dto.Wheather
	_ = json.Unmarshal(weatherResponse, &weatherDto)
	defer response.Body.Close()
	var temps dto.TempResponseDTO
	temps.City = weatherDto.Location.Name
	temps.Temp_f = pkg.CalcFarenheit(weatherDto.Current.TempC)
	temps.Temp_k = pkg.CalcKelvin(weatherDto.Current.TempC)
	temps.Temp_c = weatherDto.Current.TempC
	jsonTemp, err := json.Marshal(temps)
	if err != nil {
		log.Fatalln("\nError enconding json:", err.Error())
	}
	w.Write([]byte(jsonTemp))
}
