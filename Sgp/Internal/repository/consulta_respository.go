// ConsultaRepositoryImpl implements the ConsultaRepository interface

package repository

import (
	"context"
	"fmt"
	"time"
	"sgp/Internal/model"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type ConsultaRepositoryImpl struct {
	Client *firestore.Client
}

func NewConsultaRepository(client *firestore.Client) *ConsultaRepositoryImpl {
	return &ConsultaRepositoryImpl{Client: client}
}

func (r *ConsultaRepositoryImpl) AgendarConsulta(ctx context.Context, consulta model.Consulta) (*model.Consulta, error) {
	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		horarioRef := r.Client.Collection("horariosDisponiveis").Doc(consulta.HorarioID)
		horarioDoc, err := tx.Get(horarioRef)
		if err != nil {
			return fmt.Errorf("erro ao buscar horário para agendamento: %w", err)
		}

		var horario model.HorarioDisponivel
		horarioDoc.DataTo(&horario)
		if horario.Status != "disponivel" {
			return fmt.Errorf("o horário selecionado não está mais disponível")
		}

		// O horário fica "agendado" para que não possa ser pego por outro aluno
		if err := tx.Update(horarioRef, []firestore.Update{{Path: "status", Value: "agendado"}}); err != nil {
			return err
		}

		// Preenche os dados da consulta com o status pendente
		consulta.Inicio = horario.Inicio
		consulta.Fim = horario.Fim
		consulta.PsicologoID = horario.PsicologoID
		// MODIFICADO: Status inicial agora é "aguardando aprovacao"
		consulta.Status = "aguardando aprovacao"
		consulta.DataAgendamento = time.Now()

		consultaRef := r.Client.Collection("Consultas").NewDoc()
		consulta.ID = consultaRef.ID
		return tx.Create(consultaRef, consulta)
	})

	if err != nil {
		return nil, err
	}

	return &consulta, nil
}

func (r *ConsultaRepositoryImpl) AtualizaStatusConsulta(ctx context.Context, id string, novoStatus string) error {
	consultaRef := r.Client.Collection("Consultas").Doc(id)

	// Se o status for cancelada pelo aluno o horário fica como disponivel novamente.
	if  novoStatus == "cancelada pelo aluno" {
		return r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
			consultaDoc, err := tx.Get(consultaRef)
			if err != nil {
				return fmt.Errorf("erro ao buscar consulta para atualização: %w", err)
			}

			var consulta model.Consulta
			consultaDoc.DataTo(&consulta)

			if consulta.HorarioID != "" {
				horarioRef := r.Client.Collection("horariosDisponiveis").Doc(consulta.HorarioID)
				
				if _, err := tx.Get(horarioRef); err == nil {
					if err := tx.Update(horarioRef, []firestore.Update{{Path: "status", Value: "disponivel"}}); err != nil {
						return fmt.Errorf("erro ao reverter status do horário: %w", err)
					}
				}
			}

			return tx.Update(consultaRef, []firestore.Update{{Path: "status", Value: novoStatus}})
		})
	}

	_, err := consultaRef.Update(ctx, []firestore.Update{
		{Path: "status", Value: novoStatus},
	})
	if err != nil {
		return fmt.Errorf("erro ao atualizar status da consulta com ID '%s': %v", id, err)
	}
	return nil
}

func (r *ConsultaRepositoryImpl) ListarConsultasPorPsicologo(ctx context.Context, psicologoID string, statusFiltro string) ([]*model.Consulta, error) {
	var consultas []*model.Consulta

	query := r.Client.Collection("Consultas").Where("psicologoId", "==", psicologoID)
	if statusFiltro != "" {
		query = query.Where("status", "==", statusFiltro)
	}

	iter := query.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("erro ao iterar sobre consultas do psicologo '%s': %v", psicologoID, err)
		}

		var consulta model.Consulta
		if err := doc.DataTo(&consulta); err != nil {
			fmt.Printf("erro ao converter dados do doc '%s': %v", doc.Ref.ID, err)
			continue
		}

		consulta.ID = doc.Ref.ID
		consultas = append(consultas, &consulta)
	}

	return consultas, nil
}

func (r *ConsultaRepositoryImpl) ListarConsultasPorAluno(ctx context.Context, alunoID string) ([]*model.Consulta, error) {
	var consultas []*model.Consulta

	iter := r.Client.Collection("Consultas").Where("alunoId", "==", alunoID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("erro ao iterar sobre consultas do aluno '%s': %v", alunoID, err)
		}

		var consulta model.Consulta
		if err := doc.DataTo(&consulta); err != nil {
			fmt.Printf("erro ao converter dados do doc '%s': %v", doc.Ref.ID, err)
			continue
		}

		consulta.ID = doc.Ref.ID
		consultas = append(consultas, &consulta)
	}

	return consultas, nil
}

func (r *ConsultaRepositoryImpl) DeletarConsulta(ctx context.Context, id string) error {
	_, err := r.Client.Collection("Consultas").Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("erro ao deletar consulta com ID '%s': %v", id, err)
	}
	return nil
}

func (r *ConsultaRepositoryImpl) BuscarConsultaPorID(ctx context.Context, id string) (*model.Consulta, error) {
	doc, err := r.Client.Collection("Consultas").Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("consulta não encontrada: %w", err)
	}
	var consulta model.Consulta
	if err := doc.DataTo(&consulta); err != nil {
		return nil, err
	}
	consulta.ID = doc.Ref.ID
	return &consulta, nil
}
