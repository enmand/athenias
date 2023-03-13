package matrix

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/sqlstatestore"
	"maunium.net/go/mautrix/util/dbutil"
)

// WithSQLiteStateStore uses the SQLite state store
type SQLiteStateStore struct {
	*dbutil.Database
}

func (s SQLiteStateStore) stateOpts() {}

func (s SQLiteStateStore) Configure(c *mautrix.Client) error {
	c.StateStore = sqlstatestore.NewSQLStateStore(s.Database, dbutil.ZeroLogger(c.Log))

	if err := c.StateStore.(*sqlstatestore.SQLStateStore).Upgrade(); err != nil {
		return errors.Wrap(err, "failed to upgrade state store")
	}

	return nil
}

// WithSQLiteSyncStore uses the SQLite sync store
func WithSQLiteStateStore(db *dbutil.Database) StateStoreOption[SQLiteStateStore] {
	return func(o *SQLiteStateStore) {
		o.Database = db
	}
}

// SQLiteStore is the SQLite sync store
type SQLiteStore struct {
	*sql.DB
}

func (SQLiteStore) syncOpts() {}

func (s SQLiteStore) Configure(c *mautrix.Client) error {
	ns := new(SQLiteStore)
	ns.DB = s.DB
	c.Store = ns

	return nil
}

// WithSQLiteDB sets the SQLite database to use
func WithSQLiteSyncStore(db *sql.DB) SyncStoreOption[SQLiteStore] {
	return func(o *SQLiteStore) {
		o.DB = db
		if err := o.init(); err != nil {
			panic(err)
		}
	}
}

// SaveFilterID saves the filter ID for the given user ID
func (s *SQLiteStore) SaveFilterID(userID id.UserID, filterID string) {
	_, _ = s.Exec("INSERT INTO filter_ids (user_id, filter_id) VALUES (?, ?)", userID, filterID)
}

// LoadFilterID loads the filter ID for the given user ID
func (s *SQLiteStore) LoadFilterID(userID id.UserID) string {
	var filterID string
	err := s.QueryRow("SELECT filter_id FROM filter_ids WHERE user_id = ?", userID).Scan(&filterID)
	if err != nil {
		return ""
	}
	return filterID
}

// SaveNextBatch saves the next batch for the given user ID
func (s *SQLiteStore) SaveNextBatch(userID id.UserID, nextBatch string) {
	_, _ = s.Exec("INSERT INTO next_batch (user_id, next_batch) VALUES (?, ?)", userID, nextBatch)
}

// LoadNextBatch loads the next batch for the given user ID
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

// WithSQLCryptoStore uses the SQL crypto store
type SQLCryptoStore struct {
	*dbutil.Database
}

func (s SQLCryptoStore) chStoreOpts() {}

// Get returns the crypto store
func (s SQLCryptoStore) Get() any {
	return s.Database
}

// Managed returns whether the store is managed crypto store
func (s SQLCryptoStore) Managed() bool { return true }

// WithSQLCryptoStore uses the SQL crypto store
func WithSQLCryptoStore(db *dbutil.Database) CryptoHelperStoreOption[SQLCryptoStore] {
	return func(o *SQLCryptoStore) {
		o.Database = db
	}
}
