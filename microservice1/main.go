package main

import (
	"encoding/json"
	"log"
	"net/http"

	"hello_world/observability"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type Response struct {
	StatusCode int
	Data       any
}

func main() {
	// Initialize OpenTelemetry
	shutdown := observability.InitTracer("service1")
	defer shutdown()

	r := chi.NewRouter()

	// Root endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, span := otel.Tracer("service1").Start(r.Context(), "GET /")
		defer span.End()

		w.Header().Set("Content-Type", "application/json")
		resp := Response{
			StatusCode: 200,
			Data:       map[string]any{"message": "Hello from MS1"},
		}
		json.NewEncoder(w).Encode(resp)

		span.AddEvent("response_sent")
	})

	// Call Service2 endpoint
	r.Get("/call-service2", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("service1").Start(r.Context(), "GET /call-service2")
		defer span.End()

		// Child span for the actual HTTP call
		httpSpanCtx, httpSpan := otel.Tracer("service1").Start(ctx, "call_service2_http")
		req, _ := http.NewRequestWithContext(httpSpanCtx, "GET", "http://localhost:8082/service2", nil)
		otel.GetTextMapPropagator().Inject(httpSpanCtx, propagation.HeaderCarrier(req.Header))

		resp2, err := http.DefaultClient.Do(req)
		if err != nil {
			httpSpan.End()
			http.Error(w, "Failed to call Service 2", http.StatusInternalServerError)
			return
		}
		defer resp2.Body.Close()
		httpSpan.End() // mark the end of HTTP call span

		var service2Resp Response
		if err := json.NewDecoder(resp2.Body).Decode(&service2Resp); err != nil {
			http.Error(w, "Failed to parse Service 2 response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := Response{
			StatusCode: 200,
			Data: map[string]any{
				"Service_1_Response": "Successful",
				"Service_2_Response": service2Resp.Data,
			},
		}
		json.NewEncoder(w).Encode(resp)

		span.AddEvent("response_sent")
	})

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	addr := "0.0.0.0:8081"
	log.Printf("Starting MS1 server on %s...\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
