package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sgp/Internal/handler"
	"sgp/Internal/middleware"
	"sgp/Internal/repository"
	"sgp/Internal/service"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/joho/godotenv"
	"github.com/rs/cors" // MODIFICADO: Importação do pacote CORS
	"google.golang.org/api/option"
)

const is_middleware_on = false

func main() {
	err := godotenv.Load("/home/marquin/home/servidor-sgp/SGP-Server/Sgp/Cmd/api/.env")
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}

	creds := os.Getenv("CREDS")

	//---------------Conexao com o firebase------------------//

	ctx := context.Background()
	opt := option.WithCredentialsFile(creds)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("erro ao inicializar firebase: %v", err)
	}

	var authClient *auth.Client
	authClient, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("erro ao inicializar cliente de autenticação: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("erro ao conectar ao firestore: %v", err)
	}
	defer client.Close()


	resendApiKey := os.Getenv("RESEND_API_KEY")
	

	emailService := service.NewEmailService(resendApiKey)

	alunoRepo := repository.NewAlunoRepository(client)
	psicologoRepo := repository.NewPsicologoRepository(client)
	consultaRepo := repository.NewConsultaRepository(client)
	horarioRepo := repository.NewHorarioDisponivelRepository(client)

	alunoHandler := handler.NewAlunoHandler(alunoRepo)
	psicologoHandler := handler.NewPsicologoHandler(psicologoRepo)
	consultaHandler := handler.NewConsultaHandler(consultaRepo, alunoRepo, psicologoRepo, emailService)
	horarioHandler := handler.NewHorarioDisponivelHandler(horarioRepo)

	mux := http.NewServeMux()

	authMiddleware := middleware.NewAuthMiddleware(authClient)

	if is_middleware_on {
		mux.Handle("POST /alunos", authMiddleware.Verify(http.HandlerFunc(alunoHandler.HandlerCriarAluno)))
		mux.Handle("GET /alunos", authMiddleware.Verify(http.HandlerFunc(alunoHandler.HandlerListarAlunos)))
		mux.Handle("GET /alunos/{id}", authMiddleware.Verify(http.HandlerFunc(alunoHandler.HandlerBuscarAlunoPorID)))
		mux.Handle("GET /alunos/nome", authMiddleware.Verify(http.HandlerFunc(alunoHandler.HandlerBuscarAlunoPorNome)))
		mux.Handle("PUT /alunos/{id}", authMiddleware.Verify(http.HandlerFunc(alunoHandler.HandlerAtualizarAluno)))
		mux.Handle("DELETE /alunos/{id}", authMiddleware.Verify(http.HandlerFunc(alunoHandler.HandlerDeletarAluno)))

		mux.Handle("POST /psicologos", authMiddleware.Verify(http.HandlerFunc(psicologoHandler.HandlerCriarPsicologo)))
		mux.Handle("GET /psicologos", authMiddleware.Verify(http.HandlerFunc(psicologoHandler.HandlerListarPsicologos)))
		mux.Handle("GET /psicologos/{id}", authMiddleware.Verify(http.HandlerFunc(psicologoHandler.HandlerBuscarPsicologoPorID)))
		mux.Handle("GET /psicologos/nome", authMiddleware.Verify(http.HandlerFunc(psicologoHandler.HandlerBuscarPsicologoPorNome)))
		mux.Handle("PUT /psicologos/{id}", authMiddleware.Verify(http.HandlerFunc(psicologoHandler.HandlerAtualizarPsicologo)))
		mux.Handle("DELETE /psicologos/{id}", authMiddleware.Verify(http.HandlerFunc(psicologoHandler.HandlerDeletarPsicologo)))

		mux.HandleFunc("POST /consultas", consultaHandler.HandlerAgendarConsulta)
		mux.Handle("GET /consultas/psicologo", authMiddleware.Verify(http.HandlerFunc(consultaHandler.HandlerListarConsultasPorPsicologo)))
		mux.Handle("PATCH /consultas/{id}/status", authMiddleware.Verify(http.HandlerFunc(consultaHandler.HandlerAtualizarStatusConsulta)))
		mux.Handle("DELETE /consultas/{id}", authMiddleware.Verify(http.HandlerFunc(consultaHandler.HandlerDeletarConsulta)))

		mux.HandleFunc("POST /horarios", horarioHandler.HandlerCriarHorario)
		mux.HandleFunc("GET /horarios", horarioHandler.HandlerListarHorarios)
		mux.HandleFunc("DELETE /horarios/{id}", horarioHandler.HandlerDeletarHorario)
	} else {
		mux.HandleFunc("POST /alunos", alunoHandler.HandlerCriarAluno)
		mux.HandleFunc("GET /alunos", alunoHandler.HandlerListarAlunos)
		mux.HandleFunc("GET /alunos/{id}", alunoHandler.HandlerBuscarAlunoPorID)
		mux.HandleFunc("GET /alunos/nome", alunoHandler.HandlerBuscarAlunoPorNome)
		mux.HandleFunc("PUT /alunos/{id}", alunoHandler.HandlerAtualizarAluno)
		mux.HandleFunc("DELETE /alunos/{id}", alunoHandler.HandlerDeletarAluno)

		mux.HandleFunc("GET /consultas/aluno", consultaHandler.HandlerListarConsultasPorAluno)

		mux.HandleFunc("POST /psicologos", psicologoHandler.HandlerCriarPsicologo)
		mux.HandleFunc("GET /psicologos", psicologoHandler.HandlerListarPsicologos)
		mux.HandleFunc("GET /psicologos/{id}", psicologoHandler.HandlerBuscarPsicologoPorID)
		mux.HandleFunc("GET /psicologos/nome", psicologoHandler.HandlerBuscarPsicologoPorNome)
		mux.HandleFunc("PUT /psicologos/{id}", psicologoHandler.HandlerAtualizarPsicologo)
		mux.HandleFunc("DELETE /psicologos/{id}", psicologoHandler.HandlerDeletarPsicologo)

		mux.HandleFunc("POST /consultas", consultaHandler.HandlerAgendarConsulta)
		mux.HandleFunc("GET /consultas/psicologo", consultaHandler.HandlerListarConsultasPorPsicologo)
		mux.HandleFunc("PATCH /consultas/{id}/status", consultaHandler.HandlerAtualizarStatusConsulta)
		mux.HandleFunc("DELETE /consultas/{id}", consultaHandler.HandlerDeletarConsulta)

		mux.HandleFunc("POST /horarios", horarioHandler.HandlerCriarHorario)
		mux.HandleFunc("GET /horarios", horarioHandler.HandlerListarHorarios)
		mux.HandleFunc("DELETE /horarios/{id}", horarioHandler.HandlerDeletarHorario)
	}

	// MODIFICADO: Início da configuração do CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Permite requisições do seu cliente React
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Envolve o mux com o handler do CORS
	handler := c.Handler(mux)
	// MODIFICADO: Fim da configuração do CORS

	port := ":8080"
	server := &http.Server{
		Addr:         port,
		Handler:      handler, // MODIFICADO: Usa o handler com CORS em vez do 'mux' original
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("🐄 bovino na porta %s\n", port)
	log.Fatal(server.ListenAndServe())
}
