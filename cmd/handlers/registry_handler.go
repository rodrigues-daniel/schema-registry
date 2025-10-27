package handlers

import (
	"fmt"
	"net/http"
)

type SchemaHandler struct {
}

func NewSchemaRegistryHandler() *SchemaHandler {
	return &SchemaHandler{}
}

func (sh *SchemaHandler) RequestSchemaInfo(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	fmt.Fprintf(w, "Olá, esta é uma resposta em texto simples!")

}
