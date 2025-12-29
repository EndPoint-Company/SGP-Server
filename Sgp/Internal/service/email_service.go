package service

import (
	"fmt"
	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	Client *resend.Client
}

func NewEmailService(apiKey string) *EmailService {
	client := resend.NewClient(apiKey)
	return &EmailService{Client: client}
}

// EnviarNotificacaoAgendamento envia e-mail para o aluno e/ou psicólogo
func (s *EmailService) EnviarNotificacaoAgendamento(emailDestino, nomeAluno, nomePsicologo, dataHora string) error {
	
	htmlContent := fmt.Sprintf(`
		<h1>Olá, %s!</h1>
		<p>Sua consulta foi agendada com sucesso. Te atualizaremos sobre qualquer mudança.</p>
		<p><strong>Psicólogo:</strong> %s</p>
		<p><strong>Data/Hora:</strong> %s</p>
		<br>
		<p>Atenciosamente,<br>Equipe SGP</p>
	`, nomeAluno, nomePsicologo, dataHora)

	params := &resend.SendEmailRequest{
		From:    "SGP <onboarding@resend.dev>",
		To:      []string{emailDestino},
		Subject: "Confirmação de Agendamento - SGP",
		Html:    htmlContent,
	}

	_, err := s.Client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("erro ao enviar email pelo Resend: %v", err)
	}

	return nil
}

// EnviarNotificacaoAtualizacaoStatus avisa sobre mudança de status (ex: confirmada, cancelada)
func (s *EmailService) EnviarNotificacaoAtualizacaoStatus(emailDestino, nomeAluno, novoStatus string) error {
	
	htmlContent := fmt.Sprintf(`
		<h1>Olá, %s!</h1>
		<p>O status da sua consulta foi atualizado.</p>
		<p><strong>Novo Status:</strong> %s</p>
		<br>
		<p>Acesse a plataforma para mais detalhes.</p>
	`, nomeAluno, novoStatus)

	params := &resend.SendEmailRequest{
		From:    "SGP <onboarding@resend.dev>",
		To:      []string{emailDestino},
		Subject: fmt.Sprintf("Atualização na Consulta: %s", novoStatus),
		Html:    htmlContent,
	}

	_, err := s.Client.Emails.Send(params)
	return err
}