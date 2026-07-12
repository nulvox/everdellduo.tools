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

// ErrInvalidSkunkZone is returned when a skunk operation references a zone
// other than events, meadow, or farm.
var ErrInvalidSkunkZone = errors.New("invalid skunk zone")

// ErrSkunkPositionTaken is returned when a skunk would move onto a
// position already occupied by another skunk in the same zone.
var ErrSkunkPositionTaken = errors.New("skunk position already taken in this zone")

// ErrSkunkNotFound is returned when a skunk operation references an
// out-of-range index.
var ErrSkunkNotFound = errors.New("skunk not found")

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

func validSkunkZone(zone SkunkZone) bool {
	for _, z := range KnownSkunkZones {
		if z == zone {
			return true
		}
	}
	return false
}

func meadowOccupiedPositions(g *Game) map[int]bool {
	occ := map[int]bool{}
	for _, s := range g.Trackers.Skunks {
		if s.Zone == SkunkZoneMeadow {
			occ[s.Position] = true
		}
	}
	return occ
}

// AdjustMarker moves a named board marker (e.g. "sun", "moon") by delta,
// one step at a time, skipping over any position occupied by a meadow
// skunk and clamping to [MinMarkerPosition, MaxMarkerPosition].
func AdjustMarker(g *Game, name string, delta int) {
	if g.Trackers.Markers == nil {
		g.Trackers.Markers = map[string]int{}
	}
	step := 1
	remaining := delta
	if delta < 0 {
		step = -1
		remaining = -delta
	}
	occ := meadowOccupiedPositions(g)
	pos := g.Trackers.Markers[name]
	for i := 0; i < remaining; i++ {
		next := pos + step
		for next >= MinMarkerPosition && next <= MaxMarkerPosition && occ[next] {
			next += step
		}
		if next < MinMarkerPosition {
			pos = MinMarkerPosition
			break
		}
		if next > MaxMarkerPosition {
			pos = MaxMarkerPosition
			break
		}
		pos = next
	}
	g.Trackers.Markers[name] = pos
}

// NextFreeSkunkPosition returns the smallest non-negative position not
// already occupied by another skunk in the given zone.
func NextFreeSkunkPosition(g *Game, zone SkunkZone) int {
	occ := map[int]bool{}
	for _, s := range g.Trackers.Skunks {
		if s.Zone == zone {
			occ[s.Position] = true
		}
	}
	pos := 0
	for occ[pos] {
		pos++
	}
	return pos
}

// AddSkunk appends a new skunk marker to the given zone, automatically
// placed at the next free position in that zone.
func AddSkunk(g *Game, zone SkunkZone) error {
	if !validSkunkZone(zone) {
		return ErrInvalidSkunkZone
	}
	pos := NextFreeSkunkPosition(g, zone)
	g.Trackers.Skunks = append(g.Trackers.Skunks, Skunk{Zone: zone, Position: pos})
	return nil
}

// SetSkunkPosition moves the skunk at index i to a new position within its
// current zone. Returns ErrSkunkPositionTaken if another skunk already
// occupies that position in the same zone.
func SetSkunkPosition(g *Game, i, position int) error {
	if i < 0 || i >= len(g.Trackers.Skunks) {
		return ErrSkunkNotFound
	}
	if position < 0 {
		position = 0
	}
	zone := g.Trackers.Skunks[i].Zone
	for j, s := range g.Trackers.Skunks {
		if j != i && s.Zone == zone && s.Position == position {
			return ErrSkunkPositionTaken
		}
	}
	g.Trackers.Skunks[i].Position = position
	return nil
}

// SetSkunkZone moves the skunk at index i into a new zone, automatically
// placing it at the next free position within that zone.
func SetSkunkZone(g *Game, i int, zone SkunkZone) error {
	if i < 0 || i >= len(g.Trackers.Skunks) {
		return ErrSkunkNotFound
	}
	if !validSkunkZone(zone) {
		return ErrInvalidSkunkZone
	}
	if g.Trackers.Skunks[i].Zone == zone {
		return nil
	}
	pos := NextFreeSkunkPosition(g, zone)
	g.Trackers.Skunks[i].Zone = zone
	g.Trackers.Skunks[i].Position = pos
	return nil
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
