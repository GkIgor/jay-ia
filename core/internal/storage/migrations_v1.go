package storage

// migrationV1DDLs contém todas as instruções DDL da versão 1 do esquema do Jay Core.
// Estas strings são imutáveis após a versão 1 ser marcada como concluída.
// Alterações futuras devem ser feitas através de um novo slice migrationV2DDLs.
var migrationV1DDLs = []string{
	// Tabela: registrations (Identidades Lógicas)
	`CREATE TABLE IF NOT EXISTS registrations (
		id TEXT PRIMARY KEY NOT NULL,
		metadata_json TEXT NOT NULL DEFAULT '{}',
		status INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);`,

	// Tabela: shared_rules (Regras de Compartilhamento)
	`CREATE TABLE IF NOT EXISTS shared_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		registration_id TEXT NOT NULL,
		target_scope INTEGER NOT NULL DEFAULT 0,
		pattern TEXT NOT NULL,
		match_type INTEGER NOT NULL DEFAULT 1,
		allowed_actions INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE
	);`,

	// Tabela: chats (Conversas)
	`CREATE TABLE IF NOT EXISTS chats (
		id TEXT PRIMARY KEY NOT NULL,
		owner_registration_id TEXT NOT NULL,
		title TEXT NOT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		metadata_json TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		FOREIGN KEY (owner_registration_id) REFERENCES registrations(id) ON DELETE RESTRICT
	);`,

	// Tabela: messages (Mensagens com Autoria Composta)
	`CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY NOT NULL,
		chat_id TEXT NOT NULL,
		author_type INTEGER NOT NULL DEFAULT 1,
		author_id TEXT NOT NULL,
		role INTEGER NOT NULL,
		content TEXT NOT NULL,
		content_type INTEGER NOT NULL DEFAULT 1,
		status INTEGER NOT NULL DEFAULT 1,
		sequence_no INTEGER NOT NULL,
		metadata_json TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
	);`,

	// Tabela: tools (Ferramentas Versionadas - SemVer via campo version)
	`CREATE TABLE IF NOT EXISTS tools (
		id TEXT PRIMARY KEY NOT NULL,
		registration_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		version TEXT NOT NULL DEFAULT '1.0.0',
		schema_json TEXT NOT NULL DEFAULT '{}',
		status INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE
	);`,

	// Tabela: voice_sessions (Sessões de Voz)
	`CREATE TABLE IF NOT EXISTS voice_sessions (
		id TEXT PRIMARY KEY NOT NULL,
		chat_id TEXT NOT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		audio_codec INTEGER NOT NULL DEFAULT 1,
		sample_rate INTEGER NOT NULL DEFAULT 16000,
		channels INTEGER NOT NULL DEFAULT 1,
		metadata_json TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
	);`,

	// Índices de Desempenho
	`CREATE INDEX IF NOT EXISTS idx_shared_rules_reg ON shared_rules(registration_id);`,
	`CREATE INDEX IF NOT EXISTS idx_chats_owner ON chats(owner_registration_id);`,
	`CREATE INDEX IF NOT EXISTS idx_messages_chat_seq ON messages(chat_id, sequence_no ASC);`,
	`CREATE INDEX IF NOT EXISTS idx_tools_reg ON tools(registration_id);`,
	`CREATE INDEX IF NOT EXISTS idx_voice_sessions_chat ON voice_sessions(chat_id);`,
}
