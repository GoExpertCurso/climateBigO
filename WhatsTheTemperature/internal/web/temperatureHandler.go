package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/GoExpertCurso/whatsTheTemperature/configs"
	dto "github.com/GoExpertCurso/whatsTheTemperature/internal/web/entity/DTOs"
	"github.com/GoExpertCurso/whatsTheTemperature/pkg"
	"github.com/gorilla/mux"
)

func SearchZipCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cep, ok := vars["cep"]
	if !ok {
		fmt.Println("cep não encontrado")
	}

	response, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
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
	SearchClimate(w, r, cepDto.Localidade)
}

func SearchClimate(w http.ResponseWriter, r *http.Request, location string) {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	params := url.Values{}
	params.Add("q", location)
	params.Add("aqi", "no")

	encodedParams := params.Encode()

	baseUrl := "https://api.weatherapi.com/v1/current.json?key=" + configs.APIKEY
	requestUrl := baseUrl + "&" + encodedParams

	response, err := http.Get(requestUrl)
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
	temps.Temp_f = pkg.CalcFarenheit(weatherDto.Current.TempC)
	temps.Temp_k = pkg.CalcKelvin(weatherDto.Current.TempC)
	temps.Temp_c = weatherDto.Current.TempC
	jsonTemp, err := json.Marshal(temps)
	if err != nil {
		log.Fatalln("\nError enconding json:", err.Error())
	}
	w.Write([]byte(jsonTemp))
}
