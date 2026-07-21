package storage

import "errors"

var (
	// ErrInvalidConfig indica que a configuração fornecida para o StorageEngine é inválida.
	ErrInvalidConfig = errors.New("storage: configuração inválida")

	// ErrStorageInitialization indica falha no processo de inicialização física ou aplicação de pragmas.
	ErrStorageInitialization = errors.New("storage: falha na inicialização do armazenamento")

	// ErrPragmaFailed indica falha na execução dos pragmas de infraestrutura.
	ErrPragmaFailed = errors.New("storage: falha ao aplicar pragmas de infraestrutura")

	// ErrNilDatabase indica que o *sql.DB fornecido ao MigrationEngine é nulo.
	ErrNilDatabase = errors.New("storage: banco de dados não pode ser nulo")

	// ErrMigrationFailed indica falha na execução de uma migration (DDL ou commit da transação).
	ErrMigrationFailed = errors.New("storage: falha na execução da migration")
)
