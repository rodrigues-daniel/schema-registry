package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigues-daniel/schema-registry/internal/config"
	"github.com/rodrigues-daniel/schema-registry/internal/container"
)

func main() {

	// Carregamento de configurações
	cfg := config.Load()

	// Constroi container de dependências,
	container, err := container.NewContainer(cfg)

	if err != nil {
		log.Fatal("Falha ao construir o container:", err)
	}
	defer container.Close()

	// Setup router
	r := setupRouter(container)

	// Inicia servidor
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// encerramento controlado
	go func() {
		log.Printf("Servidor iniciado em: %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Falha do Servidor: ", err)
		}
	}()

	// Espera por sinal de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Servidor se desligando..")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Servidor forcado ao desligamento", err)
	}

	log.Println("Servidor interrompido")

}

func setupRouter(container *container.Container) *chi.Mux {

	r := chi.NewRouter()

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "healthy"}`))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {

		r.Route("/schemas", func(r chi.Router) {
			r.Post("/", container.SchemaHandler.CreateSchema)
			r.Get("/", container.SchemaHandler.GetSchema)

		})

	})

	return r
}
