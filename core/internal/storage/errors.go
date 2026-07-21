package storage

import "errors"

var (
	// ErrInvalidConfig indica que a configuração fornecida para o StorageEngine é inválida.
	ErrInvalidConfig = errors.New("storage: configuração inválida")

	// ErrStorageInitialization indica falha no processo de inicialização física ou aplicação de pragmas no banco de dados.
	ErrStorageInitialization = errors.New("storage: falha na inicialização do armazenamento")

	// ErrPragmaFailed indica falha na execução dos pragmas de infraestrutura.
	ErrPragmaFailed = errors.New("storage: falha ao aplicar pragmas de infraestrutura")
)
