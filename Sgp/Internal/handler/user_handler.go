package handler

import (
	"encoding/json"
	"net/http"
	"sgp/Internal/repository"
)

type UserHandler struct {
	alunoRepo     repository.AlunoRepository
	psicologoRepo repository.PsicologoRepository
}

func NewUserHandler(a repository.AlunoRepository, p repository.PsicologoRepository) *UserHandler {
	return &UserHandler{
		alunoRepo:     a,
		psicologoRepo: p,
	}
}

// HandlerGetUserRole responde à rota GET /users/{id}/role
func (h *UserHandler) HandlerGetUserRole(w http.ResponseWriter, r *http.Request) {
	// Pega o ID da URL (Go 1.22+)
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID é obrigatório", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// 1. Tenta buscar na coleção de Psicólogos
	// Assumindo que o PsicologoRepository segue o mesmo padrão de nome: BuscarPsicologoPorID
	psico, err := h.psicologoRepo.BuscarPsicologoPorID(r.Context(), id)
	if err == nil && psico != nil {
		json.NewEncoder(w).Encode(map[string]string{"role": "psychologist"})
		return
	}

	// 2. Tenta buscar na coleção de Alunos
	// CORREÇÃO: Nome exato "BuscarAlunoPorID" e passando r.Context()
	aluno, err := h.alunoRepo.BuscarAlunoPorID(r.Context(), id)
	if err == nil && aluno != nil {
		json.NewEncoder(w).Encode(map[string]string{"role": "student"})
		return
	}

	// 3. Se não achou em nenhum, retorna null
	json.NewEncoder(w).Encode(map[string]interface{}{"role": nil})
}