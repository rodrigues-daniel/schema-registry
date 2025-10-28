package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Get("/", handleHome)

	r.Get("/ola/{nome}", handleOla)

	fmt.Println("Servidor iniciado na porta :8080")
	http.ListenAndServe(":8080", r)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Bem-vindo ao serviço de exemplo!"))
}

func handleOla(w http.ResponseWriter, r *http.Request) {

	nome := chi.URLParam(r, "nome")

	mensagem := fmt.Sprintf("Olá, %s!", nome)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(mensagem))
}
