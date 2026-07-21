package storage

import "strings"

// translateSQLiteError mapeia erros do driver modernc.org/sqlite para erros técnicos genéricos de infraestrutura.
func translateSQLiteError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "UNIQUE constraint failed"):
		return ErrUniqueViolation
	case strings.Contains(msg, "FOREIGN KEY constraint failed"):
		return ErrForeignKeyViolation
	default:
		return err
	}
}
