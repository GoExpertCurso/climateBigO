# ClimateBigO

This project is the implementation of two system, one validate the code zip and the other make the request e catchs the info about the wether in the location gets of zip code. In this project we also implemented observability using open telemetry. 

# Features

## Validate the zip code - catchAllTheZips - Core

  - Url: [http://localhost:8080](http://localhost:8080/)
    - Method: **Post**
    - Body:
      `
        {
	        "cep": "01153000"
        }
      `
    - Responses:
       - invalid zipcode - 422

## Search the temperature - whatsTheTemperature

  - Url: [http://localhost:8787/70070550](http://localhost:8787/70070550)
    - Method: **Get**
    - Responses:
      - `
           { "city: "São Paulo", "temp_C": 28.5, "temp_F": 28.5, "temp_K": 28.5 }
       ` - 200
      - invalid zipcode - 422
      - can not find zipcode - 404

# Execution

To run this project you need have docker/docker-compose installed and configure the api key of the whetherapi in the file of docker-compose.

Follow the nexts steps to execute project:
  1. Build the project: **docker-compose build**
  2. Execute the project: **docker-compose up** 

# Observability

To see the project's observability dashboards, access the following links:

- **Prometheus**: [dashboard](http://localhost:9090/)
- **Zipkin**: [dashboard](http://localhost:9411/)

