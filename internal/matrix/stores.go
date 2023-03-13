package matrix

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
)

// This file defines the options for the Matrix client. The stateOpts and syncOpts
// interfaces are used to make sure that only valid options are passed to the
// configuration options.
//
// SyncStoreOption and StateStoreOption are the configuration options for the
// sync and state stores respectively. They are used to pass options to the
// client.
//
// MemorySyncStore and MemoryStateStore are the memory sync and state storesa
// respectively. They are used to pass options to the client.

// syncOpts is an interface that is implemented by all sync store options
type syncOpts interface {
	syncOpts()
	Configure(*mautrix.Client) error
}

// SyncStoreOptions represents which Sync storage engine Matrix should use.
type SyncStoreOption[T syncOpts] func(*T)

// MemorySyncStore uses the memory sync store
type MemorySyncStore struct{}

func (m MemorySyncStore) syncOpts() {}

func (m MemorySyncStore) Configure(c *mautrix.Client) error {
	c.Store = mautrix.NewMemorySyncStore()

	return nil
}

// WithMemorySyncStore uses the memory sync store
func WithMemorySyncStore() SyncStoreOption[MemorySyncStore] {
	return func(*MemorySyncStore) {}
}

// stateOpts is an interface that is implemented by all state store options
type stateOpts interface {
	stateOpts()
	Configure(*mautrix.Client) error
}

// StateStoreOptions represents which State storage engine Matrix should use.
type StateStoreOption[T stateOpts] func(*T)

// MemoryStateStore uses the memory state store
type MemoryStateStore struct{}

func (m MemoryStateStore) stateOpts() {}

func (m MemoryStateStore) Configure(c *mautrix.Client) error {
	c.StateStore = mautrix.NewMemoryStateStore()

	return nil
}

// WithMemoryStateStore uses the memory state store
func WithMemoryStateStore() StateStoreOption[MemoryStateStore] {
	return func(*MemoryStateStore) {}
}

type chStoreOpts interface {
	chStoreOpts()
	Get() any
	Managed() bool
}

// CryptoHelperStoreOption represents which CryptoHelper storage engine Matrix should use.
type CryptoHelperStoreOption[T chStoreOpts] func(*T)

// MemoryCryptoStore uses the memory crypto store
type MemoryCryptoStore struct {
	crypto.Store
}

func (m MemoryCryptoStore) chStoreOpts() {}

// Get returns the crypto store
func (m MemoryCryptoStore) Get() any {
	return m.Store
}

// Managed returns whether the store is managed crypto store
func (m MemoryCryptoStore) Managed() bool { return false }

// WithMemoryCryptoStore uses the memory crypto store
func WithMemoryCryptoStore(save func() error) CryptoHelperStoreOption[MemoryCryptoStore] {
	return func(o *MemoryCryptoStore) {
		o.Store = crypto.NewMemoryStore(save)
	}
}
