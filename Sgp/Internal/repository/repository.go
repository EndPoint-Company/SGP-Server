// ConsultaRepositoryImpl implements the ConsultaRepository interface

package repository

import (
	"context"
	"sgp/Internal/model"
)

type AlunoRepository interface {
	CriarAluno(ctx context.Context, aluno model.Aluno) (*model.Aluno, error)
	ListarAlunos(ctx context.Context) ([]*model.Aluno, error)
	BuscarAlunoPorID(ctx context.Context, id string) (*model.Aluno, error)
	AtualizarAluno(ctx context.Context, id string, aluno model.Aluno) error
	DeletarAluno(ctx context.Context, id string) error
	GetAlunoIDPorNome(ctx context.Context, nome string) (string, error)
}

type PsicologoRepository interface {
	CriarPsicologo(ctx context.Context, psicologo model.Psicologo) (*model.Psicologo, error)
	ListarPsicologos(ctx context.Context) ([]*model.Psicologo, error)
	BuscarPsicologoPorID(ctx context.Context, id string) (*model.Psicologo, error)
	AtualizarPsicologo(ctx context.Context, id string, psicologo model.Psicologo) error
	DeletarPsicologo(ctx context.Context, id string) error
	GetPsicologoIDPorNome(ctx context.Context, nome string) (string, error)
}

type HorarioDisponivelRepository interface {
	CriarHorario(ctx context.Context, horario model.HorarioDisponivel) (*model.HorarioDisponivel, error)
	ListarHorariosPorPsicologo(ctx context.Context, psicologoID string, status string) ([]*model.HorarioDisponivel, error)
	BuscarHorarioPorID(ctx context.Context, id string) (*model.HorarioDisponivel, error)
	AtualizarStatusHorario(ctx context.Context, id string, novoStatus string) error
	DeletarHorario(ctx context.Context, id string) error
}

type ConsultaRepository interface {
	AgendarConsulta(ctx context.Context, consulta model.Consulta) (*model.Consulta, error)
	AtualizaStatusConsulta(ctx context.Context, id string, novoStatus string) error
	ListarConsultasPorPsicologo(ctx context.Context, psicologoID string, statusFiltro string) ([]*model.Consulta, error)
	ListarConsultasPorAluno(ctx context.Context, alunoID string) ([]*model.Consulta, error)
	DeletarConsulta(ctx context.Context, id string) error
	BuscarConsultaPorID(ctx context.Context, id string) (*model.Consulta, error)
}