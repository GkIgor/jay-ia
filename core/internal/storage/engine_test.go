package storage

import (
	"testing"
)

func TestStorageEngine_New(t *testing.T) {
	t.Run("Configuração Inválida - Caminho Vazio", func(t *testing.T) {
		_, err := NewStorageEngine(Config{DatabasePath: ""})
		if err != ErrInvalidConfig {
			t.Fatalf("esperava ErrInvalidConfig, obteve: %v", err)
		}
	})

	t.Run("Configuração Válida - Suporte a Default Timeout", func(t *testing.T) {
		engine, err := NewStorageEngine(Config{DatabasePath: ":memory:"})
		if err != nil {
			t.Fatalf("erro inesperado ao instanciar: %v", err)
		}
		if engine.config.BusyTimeoutMs != 5000 {
			t.Fatalf("esperava BusyTimeoutMs = 5000, obteve: %d", engine.config.BusyTimeoutMs)
		}
	})
}

func TestStorageEngine_OpenClose_Idempotency(t *testing.T) {
	engine, err := NewStorageEngine(Config{DatabasePath: ":memory:"})
	if err != nil {
		t.Fatalf("falha no NewStorageEngine: %v", err)
	}

	// 1. Estado inicial antes do Open
	if engine.DB() != nil {
		t.Fatalf("esperava DB() == nil antes do Open()")
	}

	// 2. Primeiro Open()
	if err := engine.Open(); err != nil {
		t.Fatalf("falha no primeiro Open(): %v", err)
	}
	if engine.DB() == nil {
		t.Fatalf("esperava DB() != nil após Open()")
	}

	// 3. Segundo Open() (Idempotência)
	if err := engine.Open(); err != nil {
		t.Fatalf("esperava nil no segundo Open(), obteve: %v", err)
	}

	// 4. Primeiro Close()
	if err := engine.Close(); err != nil {
		t.Fatalf("falha no primeiro Close(): %v", err)
	}
	if engine.DB() != nil {
		t.Fatalf("esperava DB() == nil após Close()")
	}

	// 5. Segundo Close() (Idempotência)
	if err := engine.Close(); err != nil {
		t.Fatalf("esperava nil no segundo Close(), obteve: %v", err)
	}
}

func TestStorageEngine_Reopen(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := tempDir + "/reopen_test.db"

	engine, err := NewStorageEngine(Config{DatabasePath: dbPath})
	if err != nil {
		t.Fatalf("falha no NewStorageEngine: %v", err)
	}

	// Ciclo 1: Open -> Close
	if err := engine.Open(); err != nil {
		t.Fatalf("falha no primeiro Open(): %v", err)
	}
	if err := engine.Close(); err != nil {
		t.Fatalf("falha no primeiro Close(): %v", err)
	}

	// Ciclo 2: Re-Open -> Re-Close
	if err := engine.Open(); err != nil {
		t.Fatalf("falha no segundo Open() após Close(): %v", err)
	}
	if engine.DB() == nil {
		t.Fatalf("esperava DB() != nil após re-abertura")
	}
	if err := engine.Close(); err != nil {
		t.Fatalf("falha no segundo Close(): %v", err)
	}
}

func TestStorageEngine_PragmasValidation(t *testing.T) {
	engine, err := NewStorageEngine(Config{DatabasePath: ":memory:"})
	if err != nil {
		t.Fatalf("falha no NewStorageEngine: %v", err)
	}

	if err := engine.Open(); err != nil {
		t.Fatalf("falha ao abrir o banco em memória: %v", err)
	}
	defer engine.Close()

	db := engine.DB()

	// Validar Foreign Keys == 1
	var fk int
	if err := db.QueryRow("PRAGMA foreign_keys;").Scan(&fk); err != nil {
		t.Fatalf("falha ao consultar PRAGMA foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Fatalf("esperava foreign_keys == 1, obteve: %d", fk)
	}

	// Validar Busy Timeout == 5000
	var timeout int
	if err := db.QueryRow("PRAGMA busy_timeout;").Scan(&timeout); err != nil {
		t.Fatalf("falha ao consultar PRAGMA busy_timeout: %v", err)
	}
	if timeout != 5000 {
		t.Fatalf("esperava busy_timeout == 5000, obteve: %d", timeout)
	}

	// Validar Journal Mode
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode;").Scan(&journalMode); err != nil {
		t.Fatalf("falha ao consultar PRAGMA journal_mode: %v", err)
	}
	if journalMode == "" {
		t.Fatalf("esperava journalMode não vazio")
	}
}

func TestStorageEngine_CreateParentDirectories(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := tempDir + "/sub/folder/nested_jay.db"

	engine, err := NewStorageEngine(Config{DatabasePath: nestedPath})
	if err != nil {
		t.Fatalf("falha no NewStorageEngine: %v", err)
	}

	if err := engine.Open(); err != nil {
		t.Fatalf("falha ao abrir banco em diretório inexistente: %v", err)
	}
	defer engine.Close()

	if engine.DB() == nil {
		t.Fatalf("esperava DB() != nil após criar pastas pai")
	}
}

func TestStorageEngine_NegativeCases(t *testing.T) {
	t.Run("Acesso pós Close", func(t *testing.T) {
		engine, err := NewStorageEngine(Config{DatabasePath: ":memory:"})
		if err != nil {
			t.Fatalf("falha no NewStorageEngine: %v", err)
		}
		_ = engine.Open()
		_ = engine.Close()

		if engine.DB() != nil {
			t.Fatalf("DB() deve retornar nil após Close()")
		}
	})
}
