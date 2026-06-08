package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hsm-gustavo/authentication/internal/config"
	"github.com/hsm-gustavo/authentication/internal/logger"
	"github.com/hsm-gustavo/authentication/internal/routes"
)

func main() {
	config := config.Load()
	
	loggerCfg := logger.Config{
		Env: config.Env,
		Level: "debug",
	}

	log := logger.New(loggerCfg)

	log.Info("Iniciando o servidor", slog.String("PORT", config.Port))

	srv := &http.Server{
		Addr: ":" + config.Port,
		Handler: routes.Setup(log),
		// tempos de timeout para melhorar a resiliência do servidor
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// canal de tamanho 1 que recebe apenas sinais do SO
	done := make(chan os.Signal, 1)
	// registra os sinais de interrupção (Ctrl+C) e término (kill) para o canal done
	// toda vez que o programa receber um desses sinais, ele será enviado para o canal done
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Não foi possível iniciar o servidor", "erro", err)
			os.Exit(1)
		}
	}()

	<-done
	log.Info("Servidor parou")

	// tempo máximo para aguardar as requisições em andamento antes de forçar o encerramento do servidor
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Não foi possível desligar o servidor", "erro", err)
		os.Exit(1)
	}
	
	log.Info("Servidor desligado com sucesso")
}