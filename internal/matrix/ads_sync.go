package matrix

import (
	"fmt"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type AccountDataStore struct {
	eventName event.Type
}

func (AccountDataStore) syncOpts() {}

func (ads AccountDataStore) Configure(c *mautrix.Client) error {
	c.Syncer.(*mautrix.DefaultSyncer).FilterJSON.AccountData = mautrix.FilterPart{
		Limit: 20,
		NotTypes: []event.Type{
			ads.eventName,
		},
	}

	c.Store = mautrix.NewAccountDataStore(ads.eventName.Type, c)

	return nil
}

func WithEventType(evtType event.Type) SyncStoreOption[AccountDataStore] {
	return func(o *AccountDataStore) {
		o.eventName = evtType
	}
}

var (
	// AccountDataStoreEventType is the event type for the account data store
	accountDataStoreEventName = "com.unerror.athenais.%s.account_data_store"
)

// NewAccountDataStoreEventType creates a new account data store event type
func NewAccountDataStoreEventType(userID id.UserID) event.Type {
	return event.NewEventType(fmt.Sprintf(accountDataStoreEventName, userID))
}
