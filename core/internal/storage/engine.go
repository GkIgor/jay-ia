package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Config declara os parâmetros de configuração do StorageEngine.
type Config struct {
	DatabasePath  string
	BusyTimeoutMs int
}

// StorageEngine é o gerenciador de infraestrutura e ciclo de vida da conexão SQLite.
type StorageEngine struct {
	db     *sql.DB
	config Config
}

// NewStorageEngine valida a configuração, realiza cópia por valor do Config e instancia o StorageEngine.
func NewStorageEngine(config Config) (*StorageEngine, error) {
	if strings.TrimSpace(config.DatabasePath) == "" {
		return nil, ErrInvalidConfig
	}

	if config.BusyTimeoutMs <= 0 {
		config.BusyTimeoutMs = 5000
	}

	return &StorageEngine{
		config: config,
	}, nil
}

// Open abre a conexão física com o SQLite, garante diretórios pai no SO e aplica os pragmas na ordem determinística.
// É uma operação idempotente (retorna nil se o engine já estiver no estado Ready).
func (e *StorageEngine) Open() error {
	if e.db != nil {
		return nil
	}

	path := e.config.DatabasePath

	// Cria os diretórios pai caso seja um caminho físico e não um banco em memória RAM
	if path != ":memory:" && !strings.HasPrefix(path, "file:") {
		dir := filepath.Dir(path)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return fmt.Errorf("%w: falha ao criar diretório do banco %s: %v", ErrStorageInitialization, dir, err)
			}
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStorageInitialization, err)
	}

	// Aplica Pragmas Fixos de Infraestrutura na Ordem Determinística
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		fmt.Sprintf("PRAGMA busy_timeout = %d;", e.config.BusyTimeoutMs),
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA journal_mode = WAL;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return fmt.Errorf("%w (%s): %v", ErrPragmaFailed, pragma, err)
		}
	}

	e.db = db
	return nil
}

// Close encerra graciosamente a conexão com o banco SQLite.
// É uma operação idempotente (retorna nil se o engine já estiver fechado ou não inicializado).
func (e *StorageEngine) Close() error {
	if e.db == nil {
		return nil
	}

	err := e.db.Close()
	e.db = nil
	if err != nil {
		return fmt.Errorf("storage: falha ao fechar banco de dados: %w", err)
	}

	return nil
}

// DB retorna a referência da conexão *sql.DB para uso pelos Repositórios e Migration Engine.
// Retorna nil se o engine estiver fechado ou não inicializado.
func (e *StorageEngine) DB() *sql.DB {
	return e.db
}
