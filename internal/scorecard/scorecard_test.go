package scorecard

import (
	"testing"

	"everdellduo.tools/internal/model"
)

func TestComputeSumsOnlyCompletedGames(t *testing.T) {
	campaign := model.Campaign{}
	model.AddChapter(&campaign, "Chapter 1", false)
	g1, _ := model.StartChapterGame(&campaign, 0, "g1", model.GameTypeChallenge, []string{"Ash"}, 0)
	_ = model.SetScoreBreakdownEntry(g1, "Ash", "cards", 15)
	model.CompleteGame(g1)

	// A second, in-progress attempt should not count yet.
	_, _ = model.StartChapterGame(&campaign, 0, "g2", model.GameTypeChallenge, []string{"Ash"}, 0)

	model.AddChapter(&campaign, "Chapter 2", false)
	g3, _ := model.StartChapterGame(&campaign, 1, "g3", model.GameTypeChallenge, []string{"Ash"}, 0)
	_ = model.SetScoreBreakdownEntry(g3, "Ash", "cards", 9)
	model.CompleteGame(g3)

	sc := Compute(&campaign)
	if sc.Totals["Ash"] != 24 {
		t.Fatalf("expected total 24, got %d", sc.Totals["Ash"])
	}
	if len(sc.History) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(sc.History))
	}
}

func TestComputeNilCampaign(t *testing.T) {
	sc := Compute(nil)
	if len(sc.Totals) != 0 || len(sc.History) != 0 {
		t.Fatal("expected empty scorecard for nil campaign")
	}
}
