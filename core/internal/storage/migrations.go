package storage

import (
	"database/sql"
	"errors"
	"fmt"
)

// MigrationEngine é responsável por detectar a versão atual do esquema SQLite
// e aplicar incrementalmente as migrations pendentes.
// Ele não abre nem fecha conexões; depende exclusivamente do *sql.DB fornecido pelo StorageEngine.
type MigrationEngine struct {
	db *sql.DB
}

// NewMigrationEngine instancia o motor de migrations.
// Retorna ErrNilDatabase se a conexão fornecida for nula.
func NewMigrationEngine(db *sql.DB) (*MigrationEngine, error) {
	if db == nil {
		return nil, ErrNilDatabase
	}
	return &MigrationEngine{db: db}, nil
}

// Run verifica a versão atual do esquema e aplica as migrations pendentes em ordem.
// É idempotente: retorna nil sem executar DDL se o banco já estiver na versão mais recente.
func (m *MigrationEngine) Run() error {
	version, err := m.CurrentVersion()
	if err != nil {
		return fmt.Errorf("storage: falha ao ler versão do esquema: %w", err)
	}

	if version >= 1 {
		return nil
	}

	if err := m.applyV1(); err != nil {
		return err
	}

	return nil
}

// CurrentVersion retorna o número inteiro da versão atual do esquema no banco SQLite.
func (m *MigrationEngine) CurrentVersion() (int, error) {
	var version int
	if err := m.db.QueryRow("PRAGMA user_version;").Scan(&version); err != nil {
		return 0, fmt.Errorf("storage: falha ao consultar user_version: %w", err)
	}
	return version, nil
}

// applyV1 executa todas as instruções DDL da versão 1 dentro de uma única transação atômica.
// Se qualquer instrução falhar, a transação sofre Rollback e o banco permanece em user_version = 0.
func (m *MigrationEngine) applyV1() error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("%w: falha ao iniciar transação da migration v1: %v", ErrMigrationFailed, err)
	}

	stmts := append(migrationV1DDLs, "PRAGMA user_version = 1;")

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("%w: erro ao executar DDL (%s): %v", ErrMigrationFailed, stmt[:min(len(stmt), 60)], err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: falha ao confirmar transação da migration v1: %v", ErrMigrationFailed, err)
	}

	return nil
}

// min retorna o menor de dois inteiros (auxiliar local para truncar mensagens de erro).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Verificação em tempo de compilação que ErrMigrationFailed pode ser desembrulhado.
var _ = errors.Is(fmt.Errorf("%w", ErrMigrationFailed), ErrMigrationFailed)
