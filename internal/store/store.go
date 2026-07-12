// Package store defines the serializable manifest/backup formats used to
// persist groups durably (localStorage day-to-day, JSON file for backups).
// It is pure Go with no syscall/js dependencies.
package store

import (
	"encoding/json"
	"errors"
	"sort"

	"everdellduo.tools/internal/model"
)

// BackupVersion is stamped into every exported backup so future versions of
// the app can detect and migrate older backups.
const BackupVersion = 1

// Summary is the lightweight manifest entry shown in the group list.
type Summary struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	GameType  model.GameType `json:"gameType"`
	UpdatedAt string         `json:"updatedAt"`
}

// Backup is the full durable export: every group, in full, plus a version
// tag. This is what gets written to a downloaded .json file and what an
// import reads back.
type Backup struct {
	Version int           `json:"version"`
	Groups  []model.Group `json:"groups"`
}

// NewBackup builds a Backup from a set of groups.
func NewBackup(groups []model.Group) Backup {
	// Keep exports deterministic (sorted by name then id) so diffs between
	// backups are meaningful.
	sorted := make([]model.Group, len(groups))
	copy(sorted, groups)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Name != sorted[j].Name {
			return sorted[i].Name < sorted[j].Name
		}
		return sorted[i].ID < sorted[j].ID
	})
	return Backup{Version: BackupVersion, Groups: sorted}
}

// MarshalBackup serializes a Backup to indented JSON, suitable for a
// downloadable file.
func MarshalBackup(b Backup) (string, error) {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseBackup deserializes a backup JSON document, validating the version
// is one this build understands.
func ParseBackup(data string) (Backup, error) {
	var b Backup
	if err := json.Unmarshal([]byte(data), &b); err != nil {
		return Backup{}, err
	}
	if b.Version == 0 && len(b.Groups) == 0 {
		return Backup{}, errors.New("not a recognized everdellduo backup file")
	}
	if b.Version > BackupVersion {
		return Backup{}, errors.New("backup was created by a newer version of this app")
	}
	return b, nil
}
