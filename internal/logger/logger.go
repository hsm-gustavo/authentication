package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
)

type Config struct {
	Env string // "production", "development", "local"
	Level string // "debug", "info", "warn", "error"
	Output io.Writer // a saída do logger, padrão para os.Stderr (o console)
}

func New(cfg Config) *slog.Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
		AddSource: false,
	}

	tintOpts := &tint.Options{
		Level: level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// pega o atributo source
			if a.Key == slog.SourceKey {
				// tenta converter o valor para *slog.Source
				source, ok := a.Value.Any().(*slog.Source)
				// se conseguir, retorna source com o nome do arquivo e linha, caso contrário retorna o atributo original
				if ok {
					return slog.String("source", filepath.Base(source.File)+":"+string(rune(source.Line)))
				}
			}
			return a
		},
		TimeFormat: time.Kitchen,
		NoColor: false,
	}
	
	var handler slog.Handler

	if cfg.Env == "production" {
		// ambiente de produção normalmente usa JSON para melhor integração com sistemas de log centralizados
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		opts.AddSource = true
		// ambiente de desenvolvimento e local usam texto para facilitar a leitura no console
		handler = tint.NewHandler(cfg.Output, tintOpts)
	}

	return slog.New(handler)
}