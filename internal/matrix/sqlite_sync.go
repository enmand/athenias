package matrix

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

type SQLiteStore struct {
	*sql.DB
}

func NewSQLLiteStore(db *sql.DB) *SQLiteStore {
	s := &SQLiteStore{db}

	if err := s.init(); err != nil {
		panic(err)
	}

	return s
}

func (s *SQLiteStore) SaveFilterID(userID id.UserID, filterID string) {
	_, _ = s.Exec("INSERT INTO filter_ids (user_id, filter_id) VALUES (?, ?)", userID, filterID)
}

func (s *SQLiteStore) LoadFilterID(userID id.UserID) string {
	var filterID string
	err := s.QueryRow("SELECT filter_id FROM filter_ids WHERE user_id = ?", userID).Scan(&filterID)
	if err != nil {
		return ""
	}
	return filterID
}

func (s *SQLiteStore) SaveNextBatch(userID id.UserID, nextBatch string) {
	_, _ = s.Exec("INSERT INTO next_batch (user_id, next_batch) VALUES (?, ?)", userID, nextBatch)
}

func (s *SQLiteStore) LoadNextBatch(userID id.UserID) string {
	var nextBatch string
	err := s.QueryRow("SELECT next_batch FROM next_batch WHERE user_id = ?", userID).Scan(&nextBatch)
	if err != nil {
		return ""
	}
	return nextBatch
}

func (s *SQLiteStore) init() error {
	// build the tables for filter_ids, and next_batch
	_, err := s.Exec(`
CREATE TABLE IF NOT EXISTS filter_ids (
	user_id TEXT NOT NULL,
	filter_id TEXT NOT NULL,
	PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS next_batch (
	user_id TEXT NOT NULL,
	next_batch TEXT NOT NULL,
	PRIMARY KEY (user_id)
);
`)
	if err != nil {
		return err
	}

	return nil
}

var _ mautrix.SyncStore = (*SQLiteStore)(nil)
