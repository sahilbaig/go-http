package main

import (
	"encoding/json"
	"log"
	"net/http"

	"hello_world/observability"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

type Response struct {
	StatusCode int
	Data       any
}

func main() {
	// Initialize OpenTelemetry
	shutdown := observability.InitTracer("service2")
	defer shutdown()

	r := chi.NewRouter()

	// Service2 endpoint
	r.Get("/service2", func(w http.ResponseWriter, r *http.Request) {
		// Extract trace context from incoming headers
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := otel.Tracer("service2").Start(ctx, "GET /service2")
		defer span.End()

		span.SetAttributes(attribute.String("endpoint", "/service2"))

		w.Header().Set("Content-Type", "application/json")
		resp := Response{
			StatusCode: 200,
			Data: map[string]any{
				"ServiceName": "Service2",
			},
		}
		json.NewEncoder(w).Encode(resp)

		span.AddEvent("response_sent")
	})

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	addr := "0.0.0.0:8082"
	log.Printf("Starting MS2 server on %s...\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
