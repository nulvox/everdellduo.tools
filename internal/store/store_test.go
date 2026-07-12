package store

import (
	"testing"

	"everdellduo.tools/internal/model"
)

func TestBackupRoundTrip(t *testing.T) {
	groups := []model.Group{
		model.NewGroup("g1", "Zeta Duel", model.GameTypeDuel, []string{"Ash", "Bay"}, model.DifficultyNormal),
		model.NewGroup("g2", "Alpha Campaign", model.GameTypeCampaign, []string{"Ash"}, model.DifficultyHard),
	}
	backup := NewBackup(groups)
	if backup.Groups[0].Name != "Alpha Campaign" {
		t.Fatalf("expected deterministic sort by name, got %s first", backup.Groups[0].Name)
	}

	data, err := MarshalBackup(backup)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	parsed, err := ParseBackup(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(parsed.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(parsed.Groups))
	}
	if parsed.Version != BackupVersion {
		t.Fatalf("expected version %d, got %d", BackupVersion, parsed.Version)
	}
}

func TestParseBackupRejectsGarbage(t *testing.T) {
	if _, err := ParseBackup(`{"not":"a backup"}`); err == nil {
		t.Fatal("expected error for unrecognized document")
	}
}

func TestParseBackupRejectsNewerVersion(t *testing.T) {
	future := `{"version":999,"groups":[]}`
	if _, err := ParseBackup(future); err == nil {
		t.Fatal("expected error for future version")
	}
}
