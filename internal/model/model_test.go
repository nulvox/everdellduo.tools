package model

import "testing"

func TestNewGroupNonCampaign(t *testing.T) {
	g := NewGroup("g1", "Friday Duel", GameTypeDuel, []string{"Ash", "Bay"})
	if g.Game == nil {
		t.Fatal("expected Game to be set for non-campaign group")
	}
	if g.Campaign != nil {
		t.Fatal("expected Campaign to be nil for non-campaign group")
	}
	if len(g.Game.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(g.Game.Players))
	}
	for _, name := range DefaultResourceOrder {
		if g.Game.Players[0].Resources[name] != 0 {
			t.Fatalf("expected resource %s to start at 0", name)
		}
	}
}

func TestNewGroupCampaign(t *testing.T) {
	g := NewGroup("g2", "Campaign One", GameTypeCampaign, []string{"Ash"})
	if g.Campaign == nil {
		t.Fatal("expected Campaign to be set for campaign group")
	}
	if g.Game != nil {
		t.Fatal("expected Game to be nil for campaign group")
	}
	if len(g.Campaign.Chapters) != 0 {
		t.Fatal("expected campaign to start with no chapters")
	}
}
