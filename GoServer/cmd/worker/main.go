package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"webauthn-server/internal/config"
	jobdb "webauthn-server/internal/db"
	jobemail "webauthn-server/internal/jobs/emailclient"
	jobllm "webauthn-server/internal/jobs/llm"
	jobtemporal "webauthn-server/internal/jobs/temporal"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	database, err := jobdb.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	if err := database.InitializeDatabase(context.Background()); err != nil {
		log.Fatalf("initialize database: %v", err)
	}

	temporalClient, err := client.Dial(client.Options{HostPort: cfg.TemporalAddress})
	if err != nil {
		log.Fatalf("connect to Temporal at %s: %v", cfg.TemporalAddress, err)
	}
	defer temporalClient.Close()

	activities := &jobtemporal.JobActivities{
		Email: selectEmailClient(cfg, database.Pool()),
		LLM: jobllm.NewClient(jobllm.Config{
			Provider:      cfg.LLMProvider,
			GeminiAPIKey:  cfg.GeminiAPIKey,
			ModelName:     selectModel(cfg),
			OllamaBaseURL: cfg.OllamaBaseURL,
		}),
		Repo: &jobtemporal.SQLCardEventRepository{DB: database.Pool()},
	}

	jobWorker, err := jobtemporal.NewWorker(jobtemporal.WorkerConfig{
		Client:     temporalClient,
		Activities: activities,
	})
	if err != nil {
		log.Fatalf("create Temporal worker: %v", err)
	}

	go func() {
		if err := jobWorker.Run(worker.InterruptCh()); err != nil {
			log.Printf("Temporal worker stopped: %v", err)
		}
	}()

	log.Printf("Temporal worker listening on %s", jobtemporal.CreditLimitTaskQueue)
	<-shutdownSignal()
	jobWorker.Stop()
}

func selectEmailClient(cfg *config.Config, database *sql.DB) jobtemporal.EmailClient {
	if cfg.EmailAPI == "gmail_api" {
		return jobemail.NewGoogleClient(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_REDIRECT_URL"), func(ctx context.Context, email string) (string, error) {
			var refreshToken string
			err := database.QueryRowContext(ctx, `
				SELECT refresh_token
				FROM google_oauth_tokens
				WHERE gmail = ?
				ORDER BY last_updated DESC
				LIMIT 1
			`, email).Scan(&refreshToken)
			if err != nil {
				return "", err
			}
			return refreshToken, nil
		})
	}
	return jobemail.NewIMAPClient(cfg.IMAPHost, cfg.IMAPAppPassword)
}

func selectModel(cfg *config.Config) string {
	if cfg.LLMProvider == "gemini" {
		return cfg.GeminiModel
	}
	return cfg.OllamaModel
}

func shutdownSignal() <-chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	return signals
}
