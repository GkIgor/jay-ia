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

	// ErrNotFound é retornado quando o registro solicitado não existe no banco.
	ErrNotFound = errors.New("storage: registro não encontrado")

	// ErrAlreadyExists é retornado quando se tenta inserir um registro com ID já existente.
	ErrAlreadyExists = errors.New("storage: registro já existe")

	// ErrInvalidArgument é retornado quando um argumento obrigatório (ex: id) é inválido ou vazio.
	ErrInvalidArgument = errors.New("storage: argumento inválido")

	// ErrDeleteRestricted é retornado quando a remoção é bloqueada por dependência via FK RESTRICT.
	ErrDeleteRestricted = errors.New("storage: remoção restrita por dependência existente")
)
