package storage

import (
	"errors"
	"testing"
)

// newTestEngine é um helper que abre um banco :memory: já com StorageEngine e retorna o *sql.DB.
func newTestEngine(t *testing.T) *StorageEngine {
	t.Helper()
	engine, err := NewStorageEngine(Config{DatabasePath: ":memory:"})
	if err != nil {
		t.Fatalf("falha no NewStorageEngine: %v", err)
	}
	if err := engine.Open(); err != nil {
		t.Fatalf("falha no Open(): %v", err)
	}
	t.Cleanup(func() { _ = engine.Close() })
	return engine
}

func TestMigrationEngine_NilDatabase(t *testing.T) {
	_, err := NewMigrationEngine(nil)
	if !errors.Is(err, ErrNilDatabase) {
		t.Fatalf("esperava ErrNilDatabase, obteve: %v", err)
	}
}

func TestMigrationEngine_RunV1(t *testing.T) {
	engine := newTestEngine(t)

	migrator, err := NewMigrationEngine(engine.DB())
	if err != nil {
		t.Fatalf("falha no NewMigrationEngine: %v", err)
	}

	if err := migrator.Run(); err != nil {
		t.Fatalf("falha no Run(): %v", err)
	}

	// Valida que user_version foi atualizado para 1
	version, err := migrator.CurrentVersion()
	if err != nil {
		t.Fatalf("falha ao consultar CurrentVersion: %v", err)
	}
	if version != 1 {
		t.Fatalf("esperava user_version == 1, obteve: %d", version)
	}

	// Valida que todas as 6 tabelas foram criadas
	expectedTables := []string{
		"registrations", "shared_rules", "chats", "messages", "tools", "voice_sessions",
	}
	db := engine.DB()
	for _, table := range expectedTables {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?;", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("tabela %q não encontrada após migration v1: %v", table, err)
		}
	}

	// Valida que todos os 5 índices foram criados
	expectedIndexes := []string{
		"idx_shared_rules_reg", "idx_chats_owner", "idx_messages_chat_seq",
		"idx_tools_reg", "idx_voice_sessions_chat",
	}
	for _, idx := range expectedIndexes {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='index' AND name=?;", idx,
		).Scan(&name)
		if err != nil {
			t.Errorf("índice %q não encontrado após migration v1: %v", idx, err)
		}
	}
}

func TestMigrationEngine_Idempotency(t *testing.T) {
	engine := newTestEngine(t)

	migrator, err := NewMigrationEngine(engine.DB())
	if err != nil {
		t.Fatalf("falha no NewMigrationEngine: %v", err)
	}

	// Primeira execução: aplica a migration v1
	if err := migrator.Run(); err != nil {
		t.Fatalf("falha na primeira execução do Run(): %v", err)
	}

	// Segunda execução: deve ser no-op e retornar nil
	if err := migrator.Run(); err != nil {
		t.Fatalf("segunda execução do Run() retornou erro inesperado: %v", err)
	}

	// Versão deve continuar em 1
	version, err := migrator.CurrentVersion()
	if err != nil {
		t.Fatalf("falha ao consultar CurrentVersion: %v", err)
	}
	if version != 1 {
		t.Fatalf("esperava user_version == 1, obteve: %d", version)
	}
}

func TestMigrationEngine_TableSchema(t *testing.T) {
	engine := newTestEngine(t)

	migrator, err := NewMigrationEngine(engine.DB())
	if err != nil {
		t.Fatalf("falha no NewMigrationEngine: %v", err)
	}
	if err := migrator.Run(); err != nil {
		t.Fatalf("falha no Run(): %v", err)
	}

	db := engine.DB()

	t.Run("messages possui author_type e author_id", func(t *testing.T) {
		// Tenta inserir uma linha com author_type e author_id para confirmar que os campos existem
		_, err := db.Exec(`
			INSERT INTO registrations (id) VALUES ('reg-test');
		`)
		if err != nil {
			t.Fatalf("falha ao inserir registration: %v", err)
		}
		_, err = db.Exec(`
			INSERT INTO chats (id, owner_registration_id, title) VALUES ('chat-test', 'reg-test', 'Teste');
		`)
		if err != nil {
			t.Fatalf("falha ao inserir chat: %v", err)
		}
		_, err = db.Exec(`
			INSERT INTO messages (id, chat_id, author_type, author_id, role, content, sequence_no)
			VALUES ('msg-test', 'chat-test', 1, 'reg-test', 1, 'Olá', 1);
		`)
		if err != nil {
			t.Fatalf("tabela messages não possui author_type/author_id ou campos obrigatórios: %v", err)
		}
	})

	t.Run("tools possui campo version com default 1.0.0", func(t *testing.T) {
		// Insere uma tool sem especificar version — deve usar o default '1.0.0'
		_, err := db.Exec(`
			INSERT INTO tools (id, registration_id, name, description)
			VALUES ('tool-test', 'reg-test', 'minha-tool', 'Descrição da tool');
		`)
		if err != nil {
			t.Fatalf("falha ao inserir tool: %v", err)
		}
		var version string
		if err := db.QueryRow(`SELECT version FROM tools WHERE id = 'tool-test'`).Scan(&version); err != nil {
			t.Fatalf("falha ao consultar version da tool: %v", err)
		}
		if version != "1.0.0" {
			t.Fatalf("esperava version == '1.0.0', obteve: %q", version)
		}
	})
}
