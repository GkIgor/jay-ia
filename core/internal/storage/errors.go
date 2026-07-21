package storage

import "errors"

var (
	// ErrInvalidConfig indica que a configuração fornecida para o StorageEngine é inválida.
	ErrInvalidConfig = errors.New("storage: configuração inválida")

	// ErrDatabaseOpenFailed indica falha na abertura física do banco de dados SQLite.
	ErrDatabaseOpenFailed = errors.New("storage: falha ao abrir banco de dados")

	// ErrPragmaFailed indica falha na execução dos pragmas de infraestrutura.
	ErrPragmaFailed = errors.New("storage: falha ao aplicar pragmas de infraestrutura")
)
