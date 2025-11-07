package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rodrigues-daniel/data-platform/internal/schema"

	"github.com/rodrigues-daniel/data-platform/internal/api"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
)

func main() {
	// Conectar ao NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	// Criar bucket KV para schemas
	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket:   "schema-registry",
		TTL:      0, // Permanent
		Storage:  nats.FileStorage,
		Replicas: 1,
	})
	if err != nil {
		log.Fatalf("Failed to create KV bucket: %v", err)
	}

	// Inicializar registry
	registry := schema.NewRegistry(kv, js)

	// Configurar HTTP router
	router := mux.NewRouter()
	handlers := api.NewHandlers(registry)

	// Rotas da API
	router.HandleFunc("/schemas/{subject}/versions", handlers.RegisterSchemaHandler).Methods("POST")
	router.HandleFunc("/schemas/{subject}/versions/{version}", handlers.GetSchemaHandler).Methods("GET")
	router.HandleFunc("/subjects", handlers.ListSubjectsHandler).Methods("GET")
	router.HandleFunc("/subjects/{subject}/versions", handlers.ListVersionsHandler).Methods("GET")
	router.HandleFunc("/config/{subject}", handlers.ConfigHandler).Methods("GET", "PUT", "DELETE")
	router.HandleFunc("/compatibility/subjects/{subject}/versions", handlers.CompatibilityHandler).Methods("POST")
	router.HandleFunc("/validate/{subject}", handlers.ValidateHandler).Methods("POST")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Iniciar servidor
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Println("Schema Registry server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Aguardar shutdown
	waitForShutdown(server)
}

func waitForShutdown(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
