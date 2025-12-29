package mocks

import (
	"context"
	"sgp/Internal/model"
	"sgp/Internal/repository"
)

var _ repository.ConsultaRepository = &ConsultaRepositoryMock{}

type ConsultaRepositoryMock struct {
	AgendarConsultaFunc             func(ctx context.Context, consulta model.Consulta) (*model.Consulta, error)
	AtualizaStatusConsultaFunc      func(ctx context.Context, id string, novoStatus string) error
	ListarConsultasPorPsicologoFunc func(ctx context.Context, psicologoID string, statusFiltro string) ([]*model.Consulta, error)
	ListarConsultasPorAlunoFunc     func(ctx context.Context, alunoID string) ([]*model.Consulta, error)
	DeletarConsultaFunc             func(ctx context.Context, id string) error
	BuscarConsultaPorIDFunc         func(ctx context.Context, id string) (*model.Consulta, error)
}

func (m *ConsultaRepositoryMock) AgendarConsulta(ctx context.Context, c model.Consulta) (*model.Consulta, error) {
	return m.AgendarConsultaFunc(ctx, c)
}

func (m *ConsultaRepositoryMock) AtualizaStatusConsulta(ctx context.Context, id string, s string) error {
	return m.AtualizaStatusConsultaFunc(ctx, id, s)
}

func (m *ConsultaRepositoryMock) ListarConsultasPorPsicologo(ctx context.Context, pID string, sF string) ([]*model.Consulta, error) {
	return m.ListarConsultasPorPsicologoFunc(ctx, pID, sF)
}

func (m *ConsultaRepositoryMock) ListarConsultasPorAluno(ctx context.Context, aID string) ([]*model.Consulta, error) {
	return m.ListarConsultasPorAlunoFunc(ctx, aID)
}

func (m *ConsultaRepositoryMock) DeletarConsulta(ctx context.Context, id string) error {
	return m.DeletarConsultaFunc(ctx, id)
}

func (m *ConsultaRepositoryMock) BuscarConsultaPorID(ctx context.Context, id string) (*model.Consulta, error) {
	return m.BuscarConsultaPorIDFunc(ctx, id)
}