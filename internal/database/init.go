package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/hsm-gustavo/authentication/internal/database/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Init(dsn string) (*pgxpool.Pool, error) {
	ctx := context.Background()


	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("não foi possível criar pool de conexões: %w", err)
	}

	// tenta pingar o banco de dados para garantir que a conexão está funcionando
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("não foi possível executar ping no banco de dados: %w", err)
	}

	// cria uma fonte de migração usando o sistema de arquivos embutido (embed.go em /migrations)
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("não foi possível criar fonte de migração: %w", err)
	}

	migrateURL := dsn

	// cria um migrator usando a fonte de migração e a URL do banco de dados
	m, err := migrate.NewWithSourceInstance("iofs", d, migrateURL)
	if err != nil {
		return nil, fmt.Errorf("não foi possível criar migrator: %w", err)
	}

	// tenta aplicar as migrações, ignorando o erro ErrNoChange que indica que não há migrações para aplicar
	err = m.Up()
	if !errors.Is(err, migrate.ErrNoChange) {
		if err != nil {
			return nil, fmt.Errorf("falha ao aplicar migrações: %w", err)
		}
	}

	return pool, nil
}