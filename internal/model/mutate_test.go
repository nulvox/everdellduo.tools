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
	_ = AddSkunk(&game, SkunkZoneEvents)
	_ = AddSkunk(&game, SkunkZoneEvents)
	if game.Trackers.Skunks[0].Position != 0 || game.Trackers.Skunks[1].Position != 1 {
		t.Fatalf("expected auto-placed positions 0 and 1, got %+v", game.Trackers.Skunks)
	}
	if err := SetSkunkPosition(&game, 0, 1); err != ErrSkunkPositionTaken {
		t.Fatalf("expected ErrSkunkPositionTaken, got %v", err)
	}
	if err := SetSkunkPosition(&game, 0, 4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if game.Trackers.Skunks[0].Position != 4 {
		t.Fatalf("expected skunk 0 at 4, got %d", game.Trackers.Skunks[0].Position)
	}
	if err := SetSkunkZone(&game, 0, SkunkZoneMeadow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if game.Trackers.Skunks[0].Zone != SkunkZoneMeadow || game.Trackers.Skunks[0].Position != 0 {
		t.Fatalf("expected skunk moved to meadow at 0, got %+v", game.Trackers.Skunks[0])
	}
	if err := AddSkunk(&game, "lake"); err != ErrInvalidSkunkZone {
		t.Fatalf("expected ErrInvalidSkunkZone, got %v", err)
	}
	RemoveSkunk(&game, 1)
	if len(game.Trackers.Skunks) != 1 {
		t.Fatalf("expected 1 skunk remaining, got %d", len(game.Trackers.Skunks))
	}
}

func TestAdjustMarkerClampsAtSix(t *testing.T) {
	game := NewGame("gid", GameTypeDuel, []string{"Ash"})
	AdjustMarker(&game, "sun", 10)
	if game.Trackers.Markers["sun"] != MaxMarkerPosition {
		t.Fatalf("expected sun clamped to %d, got %d", MaxMarkerPosition, game.Trackers.Markers["sun"])
	}
	AdjustMarker(&game, "sun", -20)
	if game.Trackers.Markers["sun"] != MinMarkerPosition {
		t.Fatalf("expected sun clamped to %d, got %d", MinMarkerPosition, game.Trackers.Markers["sun"])
	}
}

func TestAdjustMarkerSkipsMeadowSkunks(t *testing.T) {
	game := NewGame("gid", GameTypeChallenge, []string{"Ash"})
	_ = AddSkunk(&game, SkunkZoneMeadow) // lands at position 0
	_ = SetSkunkPosition(&game, 0, 1)
	_ = AddSkunk(&game, SkunkZoneMeadow) // lands at position 0 (1 is taken)

	// Sun starts at 0, occupied meadow cells are at 0 and 1. Moving one
	// logical step forward must skip both, landing on 2.
	AdjustMarker(&game, "sun", 1)
	if game.Trackers.Markers["sun"] != 2 {
		t.Fatalf("expected sun to skip occupied meadow cells and land on 2, got %d", game.Trackers.Markers["sun"])
	}
}

func TestMarkerAndComplete(t *testing.T) {
	game := NewGame("gid", GameTypeDuel, []string{"Ash"})
	AdjustMarker(&game, "sun", 2)
	if game.Trackers.Markers["sun"] != 2 {
		t.Fatalf("expected sun marker at 2, got %d", game.Trackers.Markers["sun"])
	}
	CompleteGame(&game)
	if game.Status != StatusCompleted || game.CompletedAt == nil {
		t.Fatal("expected game marked completed with timestamp")
	}
}

func TestRemoveChapter(t *testing.T) {
	campaign := Campaign{}
	AddChapter(&campaign, "Chapter 1")
	AddChapter(&campaign, "Chapter 2")
	if err := RemoveChapter(&campaign, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(campaign.Chapters) != 1 || campaign.Chapters[0].Name != "Chapter 2" {
		t.Fatalf("expected only Chapter 2 to remain, got %+v", campaign.Chapters)
	}
	if err := RemoveChapter(&campaign, 5); err != ErrChapterNotFound {
		t.Fatalf("expected ErrChapterNotFound, got %v", err)
	}
}

func TestRemoveChapterGame(t *testing.T) {
	campaign := Campaign{}
	AddChapter(&campaign, "Chapter 1")
	_, _ = StartChapterGame(&campaign, 0, "g1", GameTypeChallenge, []string{"Ash"})
	_, _ = StartChapterGame(&campaign, 0, "g2", GameTypeChallenge, []string{"Ash"})
	if err := RemoveChapterGame(&campaign, 0, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(campaign.Chapters[0].Games) != 1 || campaign.Chapters[0].Games[0].ID != "g2" {
		t.Fatalf("expected only g2 to remain, got %+v", campaign.Chapters[0].Games)
	}
	if err := RemoveChapterGame(&campaign, 0, 9); err != ErrChapterGameNotFound {
		t.Fatalf("expected ErrChapterGameNotFound, got %v", err)
	}
	if err := RemoveChapterGame(&campaign, 9, 0); err != ErrChapterNotFound {
		t.Fatalf("expected ErrChapterNotFound, got %v", err)
	}
}

func TestNextSeasonProgression(t *testing.T) {
	cases := []struct {
		in       Season
		wantNext Season
		wantOK   bool
	}{
		{SeasonWinter, SeasonSpring, true},
		{SeasonSpring, SeasonSummer, true},
		{SeasonSummer, SeasonAutumn, true},
		{SeasonAutumn, "", false},
		{"", SeasonSpring, true}, // unrecognized treated as winter
	}
	for _, c := range cases {
		next, ok := NextSeason(c.in)
		if next != c.wantNext || ok != c.wantOK {
			t.Fatalf("NextSeason(%q) = (%q, %v), want (%q, %v)", c.in, next, ok, c.wantNext, c.wantOK)
		}
	}
}

func TestRandomizeSkunksWithRandNudgesOnCollisionAndWraps(t *testing.T) {
	game := NewGame("gid", GameTypeChallenge, []string{"Ash"})
	_ = AddSkunk(&game, SkunkZoneMeadow)
	_ = AddSkunk(&game, SkunkZoneMeadow)
	_ = AddSkunk(&game, SkunkZoneMeadow)
	// A rand source that always returns 5 (=> candidate position 6 every
	// time) forces every skunk to collide and nudge/wrap: 6, then 1, then 2.
	alwaysFive := func(n int) int { return 5 }
	RandomizeSkunksWithRand(&game, alwaysFive)

	seen := map[int]bool{}
	for _, s := range game.Trackers.Skunks {
		if s.Position < 1 || s.Position > MaxMarkerPosition {
			t.Fatalf("expected position in [1,%d], got %d", MaxMarkerPosition, s.Position)
		}
		if seen[s.Position] {
			t.Fatalf("expected unique positions within zone, got duplicate %d", s.Position)
		}
		seen[s.Position] = true
	}
	if game.Trackers.Skunks[0].Position != 6 || game.Trackers.Skunks[1].Position != 1 || game.Trackers.Skunks[2].Position != 2 {
		t.Fatalf("expected nudge/wrap sequence 6,1,2; got %+v", game.Trackers.Skunks)
	}
}

func TestRandomizeSkunksNeverPlacesOnZero(t *testing.T) {
	game := NewGame("gid", GameTypeChallenge, []string{"Ash"})
	_ = AddSkunk(&game, SkunkZoneFarm)
	alwaysZero := func(n int) int { return 0 } // candidate position 1
	RandomizeSkunksWithRand(&game, alwaysZero)
	if game.Trackers.Skunks[0].Position != 1 {
		t.Fatalf("expected position 1, got %d", game.Trackers.Skunks[0].Position)
	}
}

func TestAdvanceSeasonResetsBoardAndProgresses(t *testing.T) {
	game := NewGame("gid", GameTypeChallenge, []string{"Ash", "Bay"})
	if game.Season != SeasonWinter {
		t.Fatalf("expected new game to start in winter, got %s", game.Season)
	}
	AdjustMarker(&game, "sun", 4)
	AdjustMarker(&game, "moon", 3)
	_ = AddSkunk(&game, SkunkZoneMeadow)
	_ = AdjustResource(&game, "Ash", "worker", -2)
	if game.Players[0].Resources["worker"] != 1 {
		t.Fatalf("expected Ash to have 1 worker left, got %d", game.Players[0].Resources["worker"])
	}

	gameOver := AdvanceSeason(&game)
	if gameOver {
		t.Fatal("did not expect game over when advancing from winter")
	}
	if game.Season != SeasonSpring {
		t.Fatalf("expected season to advance to spring, got %s", game.Season)
	}
	if game.Trackers.Markers["sun"] != 0 || game.Trackers.Markers["moon"] != 0 {
		t.Fatalf("expected sun/moon reset to 0, got sun=%d moon=%d", game.Trackers.Markers["sun"], game.Trackers.Markers["moon"])
	}
	for _, p := range game.Players {
		if p.Resources["worker"] != StartingWorkers {
			t.Fatalf("expected %s to have %d workers, got %d", p.Name, StartingWorkers, p.Resources["worker"])
		}
	}
}

func TestAdvanceSeasonSignalsGameOverAtAutumn(t *testing.T) {
	game := NewGame("gid", GameTypeChallenge, []string{"Ash"})
	game.Season = SeasonAutumn
	AdjustMarker(&game, "sun", 3)

	gameOver := AdvanceSeason(&game)
	if !gameOver {
		t.Fatal("expected AdvanceSeason to signal game over at autumn")
	}
	if game.Season != SeasonAutumn {
		t.Fatalf("expected season to remain autumn, got %s", game.Season)
	}
	if game.Trackers.Markers["sun"] != 3 {
		t.Fatalf("expected sun marker untouched at 3, got %d", game.Trackers.Markers["sun"])
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
