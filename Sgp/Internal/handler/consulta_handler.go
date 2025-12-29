package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sgp/Internal/model"
	"sgp/Internal/repository"
	"sgp/Internal/service" // Novo import para o serviço de e-mail
	"time"
)

const Timeout = 5 * time.Second

type ConsultaHandler struct {
	Repo          repository.ConsultaRepository
	AlunoRepo     repository.AlunoRepository     // Dependência adicionada
	PsicologoRepo repository.PsicologoRepository // Dependência adicionada
	EmailService  *service.EmailService          // Dependência adicionada
}

// NewConsultaHandler atualizado com as novas dependências
func NewConsultaHandler(
	repo repository.ConsultaRepository,
	alunoRepo repository.AlunoRepository,
	psicologoRepo repository.PsicologoRepository,
	emailService *service.EmailService,
) *ConsultaHandler {
	return &ConsultaHandler{
		Repo:          repo,
		AlunoRepo:     alunoRepo,
		PsicologoRepo: psicologoRepo,
		EmailService:  emailService,
	}
}

func (h *ConsultaHandler) HandlerAgendarConsulta(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		AlunoID   string `json:"alunoId"`
		HorarioID string `json:"horarioId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Requisicao invalida", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if payload.AlunoID == "" || payload.HorarioID == "" {
		http.Error(w, "campos alunoId e horarioId sao obrigatorios", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), Timeout)
	defer cancel()

	consulta := model.Consulta{
		AlunoID:   payload.AlunoID,
		HorarioID: payload.HorarioID,
	}

	novaConsulta, err := h.Repo.AgendarConsulta(ctx, consulta)
	if err != nil {
		log.Printf("ERRO ao agendar consulta: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// --- ENVIO DE EMAIL (ASSÍNCRONO) ---
	go func() {
		// Contexto independente da requisição HTTP
		bgCtx := context.Background()

		// Busca dados completos para montar o e-mail
		aluno, errA := h.AlunoRepo.BuscarAlunoPorID(bgCtx, novaConsulta.AlunoID)
		psico, errP := h.PsicologoRepo.BuscarPsicologoPorID(bgCtx, novaConsulta.PsicologoID)

		if errA == nil && errP == nil {
			dataFormatada := novaConsulta.Inicio.Format("02/01/2006 às 15:04")

			// Dispara o e-mail
			errEmail := h.EmailService.EnviarNotificacaoAgendamento(
				aluno.Email,
				aluno.Nome,
				psico.Nome,
				dataFormatada,
			)
			if errEmail != nil {
				log.Printf("ERRO ao enviar email (Resend): %v", errEmail)
			}
		} else {
			log.Printf("ERRO ao buscar dados para email de agendamento: AlunoErr: %v, PsicoErr: %v", errA, errP)
		}
	}()
	// -----------------------------------

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(novaConsulta)
}

func (h *ConsultaHandler) HandlerListarConsultasPorPsicologo(w http.ResponseWriter, r *http.Request) {
	log.Println("--- INÍCIO: HandlerListarConsultasPorPsicologo foi chamado ---")

	psicologoId := r.URL.Query().Get("psicologoId")

	if psicologoId == "" {
		http.Error(w, "o psicologoId é obrigatorio", http.StatusBadRequest)
		return
	}
	log.Printf("Buscando consultas para o psicologoId: %s", psicologoId)

	statusFiltro := r.URL.Query().Get("status")

	ctx, cancel := context.WithTimeout(r.Context(), Timeout)
	defer cancel()

	consultas, err := h.Repo.ListarConsultasPorPsicologo(ctx, psicologoId, statusFiltro)

	if err != nil {
		log.Printf("ERRO ao listar consultas por psicologo: %v", err)
		http.Error(w, "erro ao listar consultas por psicologo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(consultas)
	log.Println("--- FIM: HandlerListarConsultasPorPsicologo finalizado com sucesso ---")
}

func (h *ConsultaHandler) HandlerListarConsultasPorAluno(w http.ResponseWriter, r *http.Request) {
	log.Println("--- INÍCIO: HandlerListarConsultasPorAluno foi chamado ---")

	alunoId := r.URL.Query().Get("alunoId")
	if alunoId == "" {
		http.Error(w, "o alunoId é obrigatorio", http.StatusBadRequest)
		return
	}

	log.Printf("Buscando consultas para o alunoId: %s", alunoId)

	ctx, cancel := context.WithTimeout(r.Context(), Timeout)
	defer cancel()

	consultas, err := h.Repo.ListarConsultasPorAluno(ctx, alunoId)
	if err != nil {
		log.Printf("ERRO ao listar consultas por aluno: %v", err)
		http.Error(w, "erro ao listar consultas por aluno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(consultas)

	log.Println("--- FIM: HandlerListarConsultasPorAluno finalizado com sucesso ---")
}

func (h *ConsultaHandler) HandlerAtualizarStatusConsulta(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "o id da consulta e obrigatorio", http.StatusBadRequest)
		return
	}
	var payload struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "requisicao invalida", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if payload.Status == "" {
		http.Error(w, "o campo status é obrigatorio", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), Timeout)
	defer cancel()

	// 1. Atualiza no banco
	if err := h.Repo.AtualizaStatusConsulta(ctx, id, payload.Status); err != nil {
		log.Printf("ERRO ao atualizar status da consulta: %v", err)
		http.Error(w, "erro ao atualizar o status da consulta", http.StatusInternalServerError)
		return
	}

	// --- ENVIO DE EMAIL (ASSÍNCRONO) ---
	go func() {
		bgCtx := context.Background()

		// Busca a consulta para saber quem é o aluno
		consulta, err := h.Repo.BuscarConsultaPorID(bgCtx, id)
		if err == nil {
			// Busca os dados do aluno para pegar o e-mail
			aluno, errA := h.AlunoRepo.BuscarAlunoPorID(bgCtx, consulta.AlunoID)
			if errA == nil {
				errEmail := h.EmailService.EnviarNotificacaoAtualizacaoStatus(
					aluno.Email,
					aluno.Nome,
					payload.Status,
				)
				if errEmail != nil {
					log.Printf("ERRO ao enviar email de status (Resend): %v", errEmail)
				}
			} else {
				log.Printf("ERRO ao buscar aluno para notificação de status: %v", errA)
			}
		} else {
			log.Printf("ERRO ao buscar consulta para notificação de status: %v", err)
		}
	}()
	// -----------------------------------

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "status da consulta atualizado com sucesso"})
}

func (h *ConsultaHandler) HandlerDeletarConsulta(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "o id da consulta e obrigatorio", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), Timeout)
	defer cancel()

	if err := h.Repo.DeletarConsulta(ctx, id); err != nil {
		log.Printf("ERRO ao deletar consulta: %v", err)
		http.Error(w, "erro ao deletar consulta", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}