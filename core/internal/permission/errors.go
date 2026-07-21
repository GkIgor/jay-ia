package permission

import "errors"

var (
	// ErrInvalidArgument indica que a requisição de acesso possui argumentos obrigatórios inválidos ou vazios.
	ErrInvalidArgument = errors.New("permission: parâmetros de requisição inválidos ou vazios")
)
