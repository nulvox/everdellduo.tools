package model

import "testing"

func TestAdjustResourceClampsAtZero(t *testing.T) {
	game := NewGame("gid", GameTypeDuel, []string{"Ash"})
	if err := AdjustResource(&game, "Ash", "berry", -5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if game.Players[0].Resources["berry"] != 0 {
		t.Fatalf("expected berry clamped to 0, got %d", game.Players[0].Resources["berry"])
	}
}

func TestAdjustResourceStopwatchCap(t *testing.T) {
	game := NewGame("gid", GameTypeCampaign, []string{"Ash"})
	if err := AdjustResource(&game, "Ash", "stopwatch", 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := AdjustResource(&game, "Ash", "stopwatch", 1); err != ErrStopwatchCap {
		t.Fatalf("expected ErrStopwatchCap, got %v", err)
	}
	if game.Players[0].Resources["stopwatch"] != 2 {
		t.Fatalf("expected stopwatch to stay at 2, got %d", game.Players[0].Resources["stopwatch"])
	}
}

func TestAdjustResourceUnknownPlayer(t *testing.T) {
	game := NewGame("gid", GameTypeDuel, []string{"Ash"})
	if err := AdjustResource(&game, "Nope", "berry", 1); err != ErrPlayerNotFound {
		t.Fatalf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestScoreBreakdownRecomputesTotal(t *testing.T) {
	game := NewGame("gid", GameTypeDuel, []string{"Ash"})
	_ = SetScoreBreakdownEntry(&game, "Ash", "cards", 10)
	_ = SetScoreBreakdownEntry(&game, "Ash", "events", 4)
	if game.Players[0].Score != 14 {
		t.Fatalf("expected total 14, got %d", game.Players[0].Score)
	}
	_ = SetScoreBreakdownEntry(&game, "Ash", "cards", 12)
	if game.Players[0].Score != 16 {
		t.Fatalf("expected total 16 after update, got %d", game.Players[0].Score)
	}
}

func TestEventLifecycle(t *testing.T) {
	game := NewGame("gid", GameTypeChallenge, []string{"Ash"})
	AddEvent(&game, "Grand Mausoleum")
	if len(game.Events) != 1 {
		t.Fatal("expected one event")
	}
	SetEventAchieved(&game, 0, true, "Ash")
	if !game.Events[0].Achieved || game.Events[0].AchievedBy != "Ash" {
		t.Fatalf("expected achieved by Ash, got %+v", game.Events[0])
	}
	SetEventAchieved(&game, 0, false, "")
	if game.Events[0].Achieved || game.Events[0].AchievedBy != "" {
		t.Fatalf("expected un-achieved, got %+v", game.Events[0])
	}
	RemoveEvent(&game, 0)
	if len(game.Events) != 0 {
		t.Fatal("expected event removed")
	}
}

func TestSkunkLifecycle(t *testing.T) {
	game := NewGame("gid", GameTypeChallenge, []string{"Ash"})
	AddSkunk(&game, 3)
	AddSkunk(&game, 5)
	SetSkunkPosition(&game, 0, 4)
	if game.Trackers.Skunks[0] != 4 {
		t.Fatalf("expected skunk 0 at 4, got %d", game.Trackers.Skunks[0])
	}
	RemoveSkunk(&game, 1)
	if len(game.Trackers.Skunks) != 1 {
		t.Fatalf("expected 1 skunk remaining, got %d", len(game.Trackers.Skunks))
	}
}

func TestMarkerAndComplete(t *testing.T) {
	game := NewGame("gid", GameTypeDuel, []string{"Ash"})
	SetMarker(&game, "sun", 2)
	if game.Trackers.Markers["sun"] != 2 {
		t.Fatalf("expected sun marker at 2, got %d", game.Trackers.Markers["sun"])
	}
	CompleteGame(&game)
	if game.Status != StatusCompleted || game.CompletedAt == nil {
		t.Fatal("expected game marked completed with timestamp")
	}
}

func TestChapterGameFlow(t *testing.T) {
	campaign := Campaign{}
	AddChapter(&campaign, "Chapter 1: Roots")
	g, err := StartChapterGame(&campaign, 0, "game-1", GameTypeChallenge, []string{"Ash"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(campaign.Chapters[0].Games) != 1 {
		t.Fatal("expected one game in chapter")
	}
	_ = SetScoreBreakdownEntry(g, "Ash", "cards", 20)
	CompleteGame(g)
	if !campaign.Chapters[0].Games[0].CompletedAt.Equal(*g.CompletedAt) {
		t.Fatal("expected in-place mutation to be reflected in campaign slice")
	}
}
