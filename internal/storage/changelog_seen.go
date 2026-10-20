package storage

import "github.com/go-musicfox/go-musicfox/internal/types"

// ChangelogSeen tracks the last version for which the user has seen the
// changelog popup. Stored as a JSON-serialized KV model in BoltDB, under a
// key of tapefox's own: it shares go-musicfox's data directory but
// numbers its versions afresh.
type ChangelogSeen struct {
	Version string `json:"version"`
}

func (c ChangelogSeen) GetDbName() string   { return types.AppDBName }
func (c ChangelogSeen) GetTableName() string { return "default_bucket" }
func (c ChangelogSeen) GetKey() string      { return "deck_changelog_seen" }
