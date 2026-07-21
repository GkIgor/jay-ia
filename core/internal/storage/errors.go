package storage

import "errors"

var (
	// ErrInvalidConfig indica que a configuração fornecida para o StorageEngine é inválida.
	ErrInvalidConfig = errors.New("storage: configuração inválida")

	// ErrStorageInitialization indica falha no processo de inicialização física ou aplicação de pragmas.
	ErrStorageInitialization = errors.New("storage: falha na inicialização do armazenamento")

	// ErrPragmaFailed indica falha na execução dos pragmas de infraestrutura.
	ErrPragmaFailed = errors.New("storage: falha ao aplicar pragmas de infraestrutura")

	// ErrNilDatabase indica que o *sql.DB fornecido ao motor ou repositório é nulo.
	ErrNilDatabase = errors.New("storage: banco de dados não pode ser nulo")

	// ErrMigrationFailed indica falha na execução de uma migration (DDL ou commit da transação).
	ErrMigrationFailed = errors.New("storage: falha na execução da migration")

	// Erros técnicos da infraestrutura de banco de dados
	ErrUniqueViolation     = errors.New("storage: violação de restrição única (unique)")
	ErrForeignKeyViolation = errors.New("storage: violação de chave estrangeira (foreign key)")

	// Erros semânticos do repositório/domínio
	ErrNotFound         = errors.New("storage: registro não encontrado")
	ErrAlreadyExists    = errors.New("storage: registro já existe")
	ErrInvalidArgument  = errors.New("storage: argumento inválido")
	ErrDeleteRestricted = errors.New("storage: remoção restrita por dependência existente")
	ErrInvalidOwner     = errors.New("storage: proprietário do registro não existe")
	ErrInvalidChat      = errors.New("storage: chat de destino não existe")
)
