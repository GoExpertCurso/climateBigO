package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GoExpertCurso/whatsTheTemperature/configs"
	"github.com/GoExpertCurso/whatsTheTemperature/internal/web"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.1"
)

func main() {
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown TracerProvider %v", err)
		}
	}()

	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.Use(otelmux.Middleware("server"))

	r.HandleFunc("/{cep}", web.SearchZipCode)

	handler := otelhttp.NewHandler(r, "http.server")

	srv := &http.Server{
		Addr:    ":" + configs.WEB_SERVER_PORT,
		Handler: handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server running on port", configs.WEB_SERVER_PORT)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", configs.WEB_SERVER_PORT, err)
		}
	}()

	<-stop

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")

	//http.ListenAndServe(":"+configs.WEB_SERVER_PORT, handler)
}

func initTracer() *trace.TracerProvider {
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		log.Fatalf("failed to initialize Zipkin exporter %v", err)
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("whatsTheTemperature"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp
}
