package model

import (
	"errors"
	"time"
)

// ErrPlayerNotFound is returned when an operation references a player name
// that does not exist in the game.
var ErrPlayerNotFound = errors.New("player not found")

// ErrStopwatchCap is returned when a stopwatch adjustment would push the
// count outside the 0-2 range.
var ErrStopwatchCap = errors.New("stopwatch count must stay between 0 and 2")

func findPlayer(g *Game, name string) *PlayerState {
	for i := range g.Players {
		if g.Players[i].Name == name {
			return &g.Players[i]
		}
	}
	return nil
}

// AdjustResource changes a named resource for a player by delta (can be
// negative). Resource values are clamped at 0; the "stopwatch" resource is
// additionally capped at StopwatchCap.
func AdjustResource(g *Game, playerName, resource string, delta int) error {
	p := findPlayer(g, playerName)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Resources == nil {
		p.Resources = map[string]int{}
	}
	next := p.Resources[resource] + delta
	if next < 0 {
		next = 0
	}
	if resource == "stopwatch" && next > StopwatchCap {
		return ErrStopwatchCap
	}
	p.Resources[resource] = next
	return nil
}

// SetScore sets a player's total score directly.
func SetScore(g *Game, playerName string, score int) error {
	p := findPlayer(g, playerName)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.Score = score
	return nil
}

// SetScoreBreakdownEntry sets one named component of a player's score
// breakdown (e.g. "cards", "events", "journey", "prosperity",
// "stopwatches") and recomputes the player's total as the sum of all
// breakdown entries.
func SetScoreBreakdownEntry(g *Game, playerName, category string, value int) error {
	p := findPlayer(g, playerName)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.ScoreBreakdown == nil {
		p.ScoreBreakdown = map[string]int{}
	}
	p.ScoreBreakdown[category] = value
	total := 0
	for _, v := range p.ScoreBreakdown {
		total += v
	}
	p.Score = total
	return nil
}

// AddEvent appends a new tracked event.
func AddEvent(g *Game, name string) {
	g.Events = append(g.Events, Event{Name: name})
}

// RemoveEvent removes the event at index i.
func RemoveEvent(g *Game, i int) {
	if i < 0 || i >= len(g.Events) {
		return
	}
	g.Events = append(g.Events[:i], g.Events[i+1:]...)
}

// SetEventAchieved marks the event at index i achieved/unachieved, optionally
// noting which player achieved it.
func SetEventAchieved(g *Game, i int, achieved bool, achievedBy string) {
	if i < 0 || i >= len(g.Events) {
		return
	}
	g.Events[i].Achieved = achieved
	if achieved {
		g.Events[i].AchievedBy = achievedBy
	} else {
		g.Events[i].AchievedBy = ""
	}
}

// SetMarker sets a named board marker (e.g. "sun", "moon") to a position.
func SetMarker(g *Game, name string, position int) {
	if g.Trackers.Markers == nil {
		g.Trackers.Markers = map[string]int{}
	}
	g.Trackers.Markers[name] = position
}

// AddSkunk appends a new skunk marker at the given position.
func AddSkunk(g *Game, position int) {
	g.Trackers.Skunks = append(g.Trackers.Skunks, position)
}

// SetSkunkPosition updates the position of the skunk marker at index i.
func SetSkunkPosition(g *Game, i, position int) {
	if i < 0 || i >= len(g.Trackers.Skunks) {
		return
	}
	g.Trackers.Skunks[i] = position
}

// RemoveSkunk removes the skunk marker at index i.
func RemoveSkunk(g *Game, i int) {
	if i < 0 || i >= len(g.Trackers.Skunks) {
		return
	}
	g.Trackers.Skunks = append(g.Trackers.Skunks[:i], g.Trackers.Skunks[i+1:]...)
}

// CompleteGame marks a game completed and stamps the completion time.
func CompleteGame(g *Game) {
	now := time.Now().UTC()
	g.Status = StatusCompleted
	g.CompletedAt = &now
}

// AddChapter appends a new empty chapter to a campaign.
func AddChapter(c *Campaign, name string) {
	c.Chapters = append(c.Chapters, Chapter{Name: name, Games: []Game{}})
}

// StartChapterGame starts a new game attempt within the chapter at index i
// and returns a pointer to it (already appended to the chapter).
func StartChapterGame(c *Campaign, chapterIndex int, id string, gameType GameType, playerNames []string) (*Game, error) {
	if chapterIndex < 0 || chapterIndex >= len(c.Chapters) {
		return nil, errors.New("chapter not found")
	}
	game := NewGame(id, gameType, playerNames)
	c.Chapters[chapterIndex].Games = append(c.Chapters[chapterIndex].Games, game)
	return &c.Chapters[chapterIndex].Games[len(c.Chapters[chapterIndex].Games)-1], nil
}

// Touch updates a group's UpdatedAt timestamp. Call after any mutation
// before persisting.
func Touch(g *Group) {
	g.UpdatedAt = time.Now().UTC()
}
