package model

import (
	"errors"
	"math/rand"
	"time"
)

// ErrPlayerNotFound is returned when an operation references a player name
// that does not exist in the game.
var ErrPlayerNotFound = errors.New("player not found")

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
// negative). Resource values are clamped at 0.
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
	p.Resources[resource] = next
	return nil
}

// AdjustTeamStopwatches changes the shared team stopwatch count for a
// campaign chapter game by delta (can be negative). Floored at 0; there is
// no upper cap since a chapter's starting count already reflects
// difficulty + any chapter bonus.
func AdjustTeamStopwatches(g *Game, delta int) {
	next := g.TeamStopwatches + delta
	if next < 0 {
		next = 0
	}
	g.TeamStopwatches = next
}

// StopwatchesUsed returns how many of the game's starting team stopwatches
// have been spent so far (never negative).
func StopwatchesUsed(g *Game) int {
	used := g.StartingTeamStopwatches - g.TeamStopwatches
	if used < 0 {
		return 0
	}
	return used
}

// StopwatchBonusKept reports whether the team started the game with at
// least one stopwatch and still has at least one remaining - the
// condition for recording the chapter's stopwatch bonus.
func StopwatchBonusKept(g *Game) bool {
	return g.StartingTeamStopwatches >= 1 && g.TeamStopwatches >= 1
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
// breakdown (e.g. "cards" [1 point per card in hand, displayed to
// players as "Cards (hand)"], "events", "journey", "prosperity") and
// recomputes the player's total as the sum of all breakdown entries.
// Team stopwatches are never a scoreable category here - see
// Game.TeamStopwatches / StopwatchBonusKept.
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

// AddChapter appends a new empty chapter to a campaign. bonusStopwatch
// marks whether this chapter grants an extra team stopwatch (see
// ComputeStartingStopwatches).
func AddChapter(c *Campaign, name string, bonusStopwatch bool) {
	c.Chapters = append(c.Chapters, Chapter{Name: name, BonusStopwatch: bonusStopwatch, Games: []Game{}})
}

// SetChapterBonusStopwatch updates whether the chapter at index i grants a
// bonus stopwatch. This only affects games started after the change.
func SetChapterBonusStopwatch(c *Campaign, chapterIndex int, bonusStopwatch bool) error {
	if chapterIndex < 0 || chapterIndex >= len(c.Chapters) {
		return ErrChapterNotFound
	}
	c.Chapters[chapterIndex].BonusStopwatch = bonusStopwatch
	return nil
}

// StartChapterGame starts a new game attempt within the chapter at index i
// and returns a pointer to it (already appended to the chapter). Campaign
// chapter games are always coop. startingStopwatches (see
// ComputeStartingStopwatches) seeds the shared team stopwatch count.
func StartChapterGame(c *Campaign, chapterIndex int, id string, gameType GameType, playerNames []string, startingStopwatches int) (*Game, error) {
	if chapterIndex < 0 || chapterIndex >= len(c.Chapters) {
		return nil, errors.New("chapter not found")
	}
	game := NewGame(id, gameType, playerNames)
	game.TeamStopwatches = startingStopwatches
	game.StartingTeamStopwatches = startingStopwatches
	c.Chapters[chapterIndex].Games = append(c.Chapters[chapterIndex].Games, game)
	return &c.Chapters[chapterIndex].Games[len(c.Chapters[chapterIndex].Games)-1], nil
}

// ErrChapterNotFound is returned when a chapter operation references an
// out-of-range chapter index.
var ErrChapterNotFound = errors.New("chapter not found")

// ErrChapterGameNotFound is returned when a chapter-game operation
// references an out-of-range game index within the chapter.
var ErrChapterGameNotFound = errors.New("chapter game not found")

// RemoveChapter deletes the chapter at index i, along with every game
// attempt (and thus scorecard history) it contains. This is destructive -
// callers should confirm with the user before calling it.
func RemoveChapter(c *Campaign, i int) error {
	if i < 0 || i >= len(c.Chapters) {
		return ErrChapterNotFound
	}
	c.Chapters = append(c.Chapters[:i], c.Chapters[i+1:]...)
	return nil
}

// RemoveChapterGame deletes the game attempt at gameIndex within the
// chapter at chapterIndex. If that attempt was completed, its score no
// longer contributes to the running scorecard once removed.
func RemoveChapterGame(c *Campaign, chapterIndex, gameIndex int) error {
	if chapterIndex < 0 || chapterIndex >= len(c.Chapters) {
		return ErrChapterNotFound
	}
	games := c.Chapters[chapterIndex].Games
	if gameIndex < 0 || gameIndex >= len(games) {
		return ErrChapterGameNotFound
	}
	c.Chapters[chapterIndex].Games = append(games[:gameIndex], games[gameIndex+1:]...)
	return nil
}

// Touch updates a group's UpdatedAt timestamp. Call after any mutation
// before persisting.
func Touch(g *Group) {
	g.UpdatedAt = time.Now().UTC()
}

// NextSeason returns the season that follows current in the fixed winter
// -> spring -> summer -> autumn progression. If current is unrecognized it
// is treated as winter (the season before spring). If current is already
// autumn, there is no next season (ok is false) - ending autumn ends the
// game rather than looping back to winter.
func NextSeason(current Season) (next Season, ok bool) {
	idx := -1
	for i, s := range SeasonOrder {
		if s == current {
			idx = i
			break
		}
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(SeasonOrder)-1 {
		return "", false
	}
	return SeasonOrder[idx+1], true
}

// RandomizeSkunks reassigns every skunk to a random position within its own
// zone, using math/rand. Positions are 1-6 (skunks can never sit on 0);
// collisions within the same zone are resolved by nudging to the next
// position, wrapping from 6 back to 1.
func RandomizeSkunks(g *Game) {
	RandomizeSkunksWithRand(g, rand.Intn)
}

// RandomizeSkunksWithRand is RandomizeSkunks with an injectable randomness
// source (randIntn(n) must return a value in [0,n)), so the nudge/wrap
// behavior can be tested deterministically.
func RandomizeSkunksWithRand(g *Game, randIntn func(n int) int) {
	byZone := map[SkunkZone][]int{}
	for i, s := range g.Trackers.Skunks {
		byZone[s.Zone] = append(byZone[s.Zone], i)
	}
	for _, indices := range byZone {
		used := map[int]bool{}
		for _, idx := range indices {
			pos := 1 + randIntn(MaxMarkerPosition)
			for used[pos] {
				pos++
				if pos > MaxMarkerPosition {
					pos = 1
				}
			}
			used[pos] = true
			g.Trackers.Skunks[idx].Position = pos
		}
	}
}

// AdvanceSeason performs the "new season" sequence: sun/moon markers reset
// to 0, skunks are re-randomized, every player's workers reset to
// StartingWorkers, and the season counter progresses. It returns true if
// the game was already in autumn (the last season) - in that case nothing
// is changed here, since ending autumn ends the game rather than advancing
// to a season that doesn't exist; the caller should prompt to finalize
// scores and call CompleteGame instead.
func AdvanceSeason(g *Game) bool {
	next, ok := NextSeason(g.Season)
	if !ok {
		return true
	}
	if g.Trackers.Markers == nil {
		g.Trackers.Markers = map[string]int{}
	}
	g.Trackers.Markers["sun"] = 0
	g.Trackers.Markers["moon"] = 0
	RandomizeSkunks(g)
	for i := range g.Players {
		if g.Players[i].Resources == nil {
			g.Players[i].Resources = map[string]int{}
		}
		g.Players[i].Resources["worker"] = StartingWorkers
	}
	g.Season = next
	return false
}
