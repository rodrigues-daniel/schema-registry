package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rodrigues-daniel/data-platform/internal/api"
	"github.com/rodrigues-daniel/data-platform/internal/schema"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var (
	requestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Número total de requisições recebidas.",
		},
	)

	requestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duração das requisições HTTP.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

type JetStreamAdapter struct {
	js nats.JetStreamContext
}

func (a *JetStreamAdapter) Publish(subj string, data []byte) error {

	err := error(nil)
	_, err = a.js.Publish(subj, data)
	return err
}

func main() {
	// Registra as métricas no registro padrão
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)

	// Inicializar NATS com JetStream
	js, kv, nc, ns := initializeNATSWithJetStream()
	defer cleanupNATS(nc, ns)

	// Inicializar registry e configurar HTTP
	registry := initializeRegistry(js, kv)
	server := setupHTTPServer(registry)

	// Demonstrar funcionamento do KV
	demonstrateKVUsage(kv)

	startMenssageria(js)
	configureConsumer(js)

	// Iniciar servidores e aguardar shutdown

	startServers(server)

}

// initializeNATSWithJetStream configura e inicia o NATS com JetStream
func initializeNATSWithJetStream() (nats.JetStreamContext, nats.KeyValue, *nats.Conn, *server.Server) {
	// Configuração do servidor NATS embutido
	storeDir := getEnv("NATS_STORE_DIR", "./jetstream-data")
	serverName := getEnv("NATS_SERVER_NAME", "schema-registry-standalone")

	// Configuração STANDALONE - sem cluster
	opts := &server.Options{
		ServerName: serverName,
		JetStream:  true,
		StoreDir:   storeDir,
		Port:       getEnvAsInt("NATS_PORT", 4222), // Porta fixa para desenvolvimento
		Host:       getEnv("NATS_HOST", "0.0.0.0"),
		HTTPPort:   getEnvAsInt("NATS_HTTP_PORT", 8222), // Monitoramento
	}

	// Criar e iniciar servidor NATS
	ns, err := server.NewServer(opts)
	if err != nil {
		log.Fatal("Erro ao criar servidor NATS:", err)
	}

	// Configurar logger
	ns.ConfigureLogger()

	// Iniciar servidor
	go ns.Start()

	// Aguardar inicialização
	if !ns.ReadyForConnections(10 * time.Second) {
		log.Fatal("NATS server não iniciou a tempo")
	}

	log.Printf("NATS Server '%s' rodando em: %s", serverName, ns.ClientURL())
	log.Printf("JetStream Store Directory: %s", storeDir)

	// Conectar ao servidor
	nc, err := nats.Connect(ns.ClientURL(),
		nats.Name("Schema-Registry-Client"),
		nats.Timeout(10*time.Second),
		nats.PingInterval(20*time.Second),
		nats.MaxPingsOutstanding(5),
	)
	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}

	// Configurar JetStream
	js, err := nc.JetStream(
		nats.MaxWait(10*time.Second),
		nats.PublishAsyncMaxPending(256),
	)

	if err != nil {
		nc.Close()
		ns.Shutdown()
		log.Fatal("Erro ao criar JetStream:", err)
	}

	// Aguardar JetStream estar pronto
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := waitForJetStream(js, ctx); err != nil {
		nc.Close()
		ns.Shutdown()
		log.Fatal("JetStream não ficou pronto a tempo:", err)
	}

	// Criar bucket KV
	kv, err := createOrGetKVBucket(js, "schemadb")
	if err != nil {
		nc.Close()
		ns.Shutdown()
		log.Fatal("Erro ao criar bucket KV:", err)
	}

	log.Println("Bucket KV 'schemadb' criado/recuperado com sucesso!")

	return js, kv, nc, ns
}

// waitForJetStream aguarda o JetStream estar pronto
func waitForJetStream(js nats.JetStreamContext, ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
			// Tentar uma operação simples para verificar se JetStream está pronto
			_, err := js.AccountInfo()
			if err == nil {
				return nil
			}
			if err != nats.ErrJetStreamNotEnabled {
				log.Printf("Aguardando JetStream... (%v)", err)
			}
		}
	}
}

// createOrGetKVBucket cria ou recupera um bucket KV existente
func createOrGetKVBucket(js nats.JetStreamContext, bucketName string) (nats.KeyValue, error) {
	// Primeiro tentar recuperar se já existe
	kv, err := js.KeyValue(bucketName)
	if err == nil {
		log.Printf("Bucket KV '%s' recuperado (já existia)", bucketName)
		return kv, nil
	}

	// Se não existe, criar novo
	kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket:   bucketName,
		TTL:      0, // Permanent
		Storage:  nats.FileStorage,
		Replicas: 1,
		History:  10, // Manter histórico de 10 versões
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Bucket KV '%s' criado com sucesso", bucketName)
	return kv, nil
}

// initializeRegistry configura o schema registry
func initializeRegistry(js nats.JetStreamContext, kv nats.KeyValue) *schema.Registry {
	storage := schema.NewStorage(kv)
	validator := schema.NewValidator(storage)
	njs := &JetStreamAdapter{js: js}

	registry := schema.NewRegistry(storage, validator, njs)
	log.Println("Schema Registry inicializado com sucesso")
	return registry
}

// setupHTTPServer configura o servidor HTTP com Gorilla Mux
func setupHTTPServer(registry *schema.Registry) *http.Server {
	router := mux.NewRouter()

	// Configurar middlewares
	setupGorillaMiddlewares(router)

	handlers := api.NewHandlers(registry)

	// Configurar rotas da API
	setupAPIRoutes(router, handlers)

	// Servidor HTTP
	server := &http.Server{
		Addr:         getEnv("HTTP_PORT", ":8080"),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

// setupGorillaMiddlewares configura os middlewares
func setupGorillaMiddlewares(router *mux.Router) {
	// Middleware para logging
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
		})
	})

	// Middleware para content-type JSON
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	// Middleware para recovery
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic recovered: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Internal server error",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	})
}

// setupAPIRoutes configura todas as rotas da API com Gorilla Mux
func setupAPIRoutes(router *mux.Router, handlers *api.Handlers) {
	// Health check
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Rotas de Schemas
	router.HandleFunc("/schemas/{subject}/versions", handlers.RegisterSchemaHandler).Methods("POST")
	router.HandleFunc("/schemas/{subject}/versions/{version}", handlers.GetSchemaHandler).Methods("GET")

	// Rotas de Subjects
	router.HandleFunc("/subjects", handlers.ListSubjectsHandler).Methods("GET")
	router.HandleFunc("/subjects/{subject}/versions", handlers.ListVersionsHandler).Methods("GET")

	// Rotas de Configuração
	router.HandleFunc("/config/{subject}", handlers.ConfigHandler).Methods("GET", "PUT", "DELETE")

	// Rotas de Compatibilidade
	router.HandleFunc("/compatibility/subjects/{subject}/versions", handlers.CompatibilityHandler).Methods("POST")
	router.HandleFunc("/validate/{subject}", handlers.ValidateHandler).Methods("POST")

	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	log.Println("Rotas da API configuradas com Gorilla Mux")
}

// healthCheckHandler manipula health checks
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "schema-registry",
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// demonstrateKVUsage demonstra o funcionamento do KV
func demonstrateKVUsage(kv nats.KeyValue) {
	// Exemplo de uso
	_, err := kv.Put("service.version", []byte("1.0.0"))
	if err != nil {
		log.Printf("Aviso: Erro ao inserir no KV: %v", err)
		return
	}

	value, err := kv.Get("service.version")
	if err != nil {
		log.Printf("Aviso: Erro ao ler do KV: %v", err)
		return
	}

	log.Printf("Valor recuperado: %s = %s", value.Key(), string(value.Value()))
}

// startServers inicia os servidores e aguarda shutdown
func startServers(server *http.Server) {
	// Canal para erros do servidor HTTP
	serverErrors := make(chan error, 1)

	// Iniciar servidor HTTP em goroutine
	go func() {
		log.Printf("Schema Registry server starting on %s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Aguardar shutdown
	waitForShutdown(server, serverErrors)
}

// waitForShutdown gerencia o shutdown graceful
func waitForShutdown(server *http.Server, serverErrors chan error) {
	// Canal para sinais do sistema operacional
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	// Aguardar sinal de shutdown ou erro do servidor
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}

	case sig := <-osSignals:
		log.Printf("Received signal: %v", sig)
	}

	log.Println("Shutting down server...")

	// Shutdown graceful
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Server stopped gracefully")
	}
}

// cleanupNATS limpa recursos do NATS
func cleanupNATS(nc *nats.Conn, ns *server.Server) {
	if nc != nil {
		nc.Close()
		log.Println("Conexão NATS fechada")
	}
	if ns != nil {
		ns.Shutdown()
		log.Println("Servidor NATS finalizado")
	}
}

// getEnv obtém variável de ambiente ou valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt obtém variável de ambiente como int
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func startMenssageria(js nats.JetStreamContext) {
	streamname := "EVENTOS_SCHEMA"

	if info, _ := js.StreamInfo(streamname); info != nil {
		log.Printf("Stream %s já existe, pulando criação", streamname)
		return
	}

	_, err := js.AddStream(&nats.StreamConfig{
		Name:     streamname,
		Subjects: []string{"schema.*"},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		log.Fatalf("Erro ao criar stream %s: %v", streamname, err)
	}
	log.Printf("Stream %s criada com sucesso para eventos de schema", streamname)
}

func configureConsumer(js nats.JetStreamContext) {
	streamName := "EVENTOS_SCHEMA"
	consumerName := "SCHEMA_CONSUMER"

	filter := "schema.user"

	// cria o consumer pull-based
	_, err := js.AddConsumer(streamName, &nats.ConsumerConfig{
		Durable:       consumerName,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: filter,
	})
	if err != nil {
		log.Fatalf("Erro ao criar consumer %s: %v", consumerName, err)
	}

	// cria o subscriber pull-based
	sub, err := js.PullSubscribe(filter, consumerName)
	if err != nil {
		log.Fatalf("Erro ao criar PullSubscribe: %v", err)
	}

	log.Printf("Consumer %s aguardando mensagens...", consumerName)

	for {
		msgs, err := sub.Fetch(5)
		if err != nil {
			log.Printf("Erro ao buscar mensagens: %v", err)
			continue
		}
		for _, m := range msgs {
			log.Printf("Recebida: %s", string(m.Data))
			m.Ack()
		}
	}
}
