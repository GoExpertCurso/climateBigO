package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GoExpertCurso/catchAllTheZips/internal/infra/web"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.15.0"
)

func main() {
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown TracerProvider %v", err)
		}
	}()

	mp := initMetrics()
	defer func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error stopping metric controller: %v", err)
		}
	}()

	r := mux.NewRouter()
	r.Use(otelmux.Middleware("server"))

	r.HandleFunc("/", web.CatchZipHandler)

	handler := otelhttp.NewHandler(r, "http.server")

	srv := &http.Server{
		Addr:    ":" + "8080",
		Handler: handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Println("Server started on :8080")
	go func() {
		log.Println("Server running on port", "8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", "8080", err)
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Prometheus metrics exposed on /metrics")
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

/* func initTracer() *trace.TracerProvider {
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		log.Fatalf("failed to initialize Zipkin exporter %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("catchAllTheZips"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp
} */

func initMetrics() *metric.MeterProvider {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatalf("failed to initialize prometeus exporter %v", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(exporter),
	)
	otel.SetMeterProvider(mp)
	return mp
}

func initTracer() *trace.TracerProvider {
	// Set up the OTLP exporter
	endpoint := "collector:4317"
	if envEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); envEndpoint != "" {
		endpoint = envEndpoint
	}
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		log.Fatalf("failed to initialize OTLP exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("catchAllTheZips"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp
}
