package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sgp/Internal/model"
	"sgp/Internal/repository/mocks"
	"testing"
	"time"
)

func TestHandlerAgendarConsulta(t *testing.T) {
	payload := map[string]string{
		"alunoId":   "aluno-1",
		"horarioId": "horario-disponivel-1",
	}
	payloadJSON, _ := json.Marshal(payload)

	t.Run("sucesso ao agendar", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/consultas", bytes.NewBuffer(payloadJSON))
		rr := httptest.NewRecorder()

		// Mock da Consulta
		mockConsultaRepo := &mocks.ConsultaRepositoryMock{
			AgendarConsultaFunc: func(ctx context.Context, c model.Consulta) (*model.Consulta, error) {
				c.ID = "consulta-123"
				c.PsicologoID = "psico-1"
				c.Status = "agendada"
				c.Inicio = time.Now()
				c.Fim = time.Now().Add(50 * time.Minute)
				c.DataAgendamento = time.Now()
				return &c, nil
			},
		}

		// Mock do Aluno (Retornando erro para evitar chamada ao EmailService nil na goroutine)
		mockAlunoRepo := &mocks.AlunoRepositoryMock{
			BuscarAlunoPorIDFunc: func(ctx context.Context, id string) (*model.Aluno, error) {
				return nil, errors.New("ignorar email no teste")
			},
		}

		// Mock do Psicologo (Retornando erro para evitar chamada ao EmailService nil na goroutine)
		mockPsicologoRepo := &mocks.PsicologoRepositoryMock{
			BuscarPsicologoPorIDFunc: func(ctx context.Context, id string) (*model.Psicologo, error) {
				return nil, errors.New("ignorar email no teste")
			},
		}

		// Passamos nil para o EmailService pois não queremos testar envio real aqui
		h := NewConsultaHandler(mockConsultaRepo, mockAlunoRepo, mockPsicologoRepo, nil)
		h.HandlerAgendarConsulta(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("status code incorreto: obteve %v, esperava %v", status, http.StatusCreated)
		}

		var novaConsulta model.Consulta
		json.NewDecoder(rr.Body).Decode(&novaConsulta)

		if novaConsulta.ID != "consulta-123" {
			t.Errorf("ID da consulta incorreto, esperava 'consulta-123', obteve '%s'", novaConsulta.ID)
		}
	})
}

func TestHandlerListarConsultasPorPsicologo(t *testing.T) {
	t.Run("sucesso ao listar", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/consultas/psicologo?psicologoId=psico-1&status=aprovada", nil)
		rr := httptest.NewRecorder()

		mockRepo := &mocks.ConsultaRepositoryMock{
			ListarConsultasPorPsicologoFunc: func(ctx context.Context, psicologoID string, statusFiltro string) ([]*model.Consulta, error) {
				if psicologoID == "psico-1" && statusFiltro == "aprovada" {
					return []*model.Consulta{{ID: "1"}, {ID: "2"}}, nil
				}
				return nil, nil
			},
		}

		// Passamos mocks vazios/nil para dependências não usadas neste endpoint
		h := NewConsultaHandler(mockRepo, &mocks.AlunoRepositoryMock{}, &mocks.PsicologoRepositoryMock{}, nil)
		h.HandlerListarConsultasPorPsicologo(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("status code incorreto: obteve %v, esperava %v", status, http.StatusOK)
		}
		var consultas []*model.Consulta
		json.NewDecoder(rr.Body).Decode(&consultas)
		if len(consultas) != 2 {
			t.Errorf("número incorreto de consultas: obteve %d, esperava 2", len(consultas))
		}
	})
}

func TestHandlerAtualizarStatusConsulta(t *testing.T) {
	payload := map[string]string{"status": "confirmada"}
	jsonBody, _ := json.Marshal(payload)

	t.Run("sucesso ao atualizar status", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", "/consultas/consulta-1/status", bytes.NewBuffer(jsonBody))
		req.SetPathValue("id", "consulta-1")
		rr := httptest.NewRecorder()

		mockRepo := &mocks.ConsultaRepositoryMock{
			AtualizaStatusConsultaFunc: func(ctx context.Context, id string, novoStatus string) error {
				return nil // Sucesso
			},
			// Retornar erro na busca para evitar chamar o email service na goroutine
			BuscarConsultaPorIDFunc: func(ctx context.Context, id string) (*model.Consulta, error) {
				return nil, errors.New("ignorar email no teste")
			},
		}

		h := NewConsultaHandler(mockRepo, &mocks.AlunoRepositoryMock{}, &mocks.PsicologoRepositoryMock{}, nil)
		h.HandlerAtualizarStatusConsulta(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("status code incorreto: obteve %v, esperava %v", status, http.StatusOK)
		}
	})
}

func TestHandlerDeletarConsulta(t *testing.T) {
	t.Run("sucesso ao deletar", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/consultas/consulta-1", nil)
		req.SetPathValue("id", "consulta-1")
		rr := httptest.NewRecorder()

		mockRepo := &mocks.ConsultaRepositoryMock{
			DeletarConsultaFunc: func(ctx context.Context, id string) error {
				return nil // Sucesso
			},
		}

		h := NewConsultaHandler(mockRepo, &mocks.AlunoRepositoryMock{}, &mocks.PsicologoRepositoryMock{}, nil)
		h.HandlerDeletarConsulta(rr, req)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("status code incorreto: obteve %v, esperava %v", status, http.StatusNoContent)
		}
	})
}