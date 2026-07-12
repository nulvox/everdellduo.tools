// Package model defines the core data types for tracking Everdell Duo
// groups, games, and campaigns. It is pure Go with no syscall/js
// dependencies so it can be tested with `go test` directly.
package model

import "time"

// GameType identifies which Everdell Duo (or future Everdell variant) mode
// a group is tracking. It is a plain string rather than a fixed enum so new
// formats can be added without breaking existing saved data.
type GameType string

const (
	GameTypeChallenge GameType = "challenge"
	GameTypeDuel      GameType = "duel"
	GameTypeCoop      GameType = "coop"
	GameTypeCampaign  GameType = "campaign"
)

// KnownGameTypes lists the built-in modes offered when starting a new
// group. Custom values are still accepted elsewhere in the model.
var KnownGameTypes = []GameType{GameTypeChallenge, GameTypeDuel, GameTypeCoop, GameTypeCampaign}

// Event represents a single event card / challenge tile being tracked
// during a game, and whether it has been achieved.
type Event struct {
	Name       string `json:"name"`
	Achieved   bool   `json:"achieved"`
	AchievedBy string `json:"achievedBy,omitempty"`
}

// SkunkZone identifies which board zone a skunk marker occupies. Everdell
// Duo's Challenge mode roams skunks across three zones; a skunk's position
// is only meaningful relative to other skunks in the same zone.
type SkunkZone string

const (
	SkunkZoneEvents SkunkZone = "events"
	SkunkZoneMeadow SkunkZone = "meadow"
	SkunkZoneFarm   SkunkZone = "farm"
)

// KnownSkunkZones lists the valid zones a skunk can be placed in.
var KnownSkunkZones = []SkunkZone{SkunkZoneEvents, SkunkZoneMeadow, SkunkZoneFarm}

// Skunk is one roaming skunk marker. Position is unique among skunks that
// share the same Zone (two skunks may share a position only if they are in
// different zones).
type Skunk struct {
	Zone     SkunkZone `json:"zone"`
	Position int       `json:"position"`
}

// MinMarkerPosition and MaxMarkerPosition bound the sun/moon track. The
// track tops out at 6 and does not go negative.
const (
	MinMarkerPosition = 0
	MaxMarkerPosition = 6
)

// Trackers holds the position of board markers. Markers is a free-form map
// so different formats can track whatever named markers they need (e.g.
// "sun", "moon"), clamped to [MinMarkerPosition, MaxMarkerPosition]. Skunks
// occupy one of three zones (events, meadow, farm); a meadow skunk blocks
// that space for the sun/moon markers, which must skip over it.
type Trackers struct {
	Markers map[string]int `json:"markers"`
	Skunks  []Skunk        `json:"skunks"`
}

// NewTrackers returns a Trackers with the default sun/moon markers at 0 and
// no skunks. Callers can add/remove skunks and markers freely.
func NewTrackers() Trackers {
	return Trackers{
		Markers: map[string]int{"sun": 0, "moon": 0},
		Skunks:  []Skunk{},
	}
}

// DefaultResourceOrder is the display order for the standard Everdell
// resources. Players/games are free to add other named resources too.
var DefaultResourceOrder = []string{"berry", "twig", "resin", "pebble", "stopwatch"}

// NewDefaultResources returns the standard starting resource map, all at
// zero. "stopwatch" is a scored resource capped at 0-2 per chapter, not a
// live countdown timer.
func NewDefaultResources() map[string]int {
	r := make(map[string]int, len(DefaultResourceOrder))
	for _, name := range DefaultResourceOrder {
		r[name] = 0
	}
	return r
}

// StopwatchCap is the maximum number of stopwatches a player may hold.
const StopwatchCap = 2

// PlayerState is one player's live state within a Game.
type PlayerState struct {
	Name           string         `json:"name"`
	Resources      map[string]int `json:"resources"`
	Score          int            `json:"score"`
	ScoreBreakdown map[string]int `json:"scoreBreakdown,omitempty"`
}

// NewPlayerState creates a player with default zeroed resources.
func NewPlayerState(name string) PlayerState {
	return PlayerState{
		Name:      name,
		Resources: NewDefaultResources(),
	}
}

// GameStatus tracks whether the live game is still being played or has
// been recorded/completed.
type GameStatus string

const (
	StatusInProgress GameStatus = "in_progress"
	StatusCompleted  GameStatus = "completed"
)

// Game is the live/current game state: players, events, board trackers.
type Game struct {
	ID          string        `json:"id"`
	GameType    GameType      `json:"gameType"`
	Players     []PlayerState `json:"players"`
	Events      []Event       `json:"events"`
	Trackers    Trackers      `json:"trackers"`
	Status      GameStatus    `json:"status"`
	StartedAt   time.Time     `json:"startedAt"`
	CompletedAt *time.Time    `json:"completedAt,omitempty"`
	Notes       string        `json:"notes,omitempty"`
}

// NewGame creates a fresh Game for the given type and player names.
func NewGame(id string, gameType GameType, playerNames []string) Game {
	players := make([]PlayerState, 0, len(playerNames))
	for _, n := range playerNames {
		players = append(players, NewPlayerState(n))
	}
	return Game{
		ID:        id,
		GameType:  gameType,
		Players:   players,
		Events:    []Event{},
		Trackers:  NewTrackers(),
		Status:    StatusInProgress,
		StartedAt: time.Now().UTC(),
	}
}

// Chapter is one chapter of a Campaign. A chapter may be played (and
// re-played) multiple times; each completed attempt is appended to Games.
type Chapter struct {
	Name  string `json:"name"`
	Games []Game `json:"games"`
}

// Campaign tracks progress across multiple chapters, each of which is
// played as one or more games whose results roll up into a running
// scorecard (computed from the chapters, see internal/scorecard).
type Campaign struct {
	Chapters []Chapter `json:"chapters"`
}

// Group is the top-level durable artifact: a single tracked group, created
// as one specific game type. For all types except campaign there is
// exactly one Game. For campaign type, Campaign is populated instead.
type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	GameType  GameType  `json:"gameType"`
	Players   []string  `json:"players"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Game      *Game     `json:"game,omitempty"`
	Campaign  *Campaign `json:"campaign,omitempty"`
}

// NewGroup creates a new Group of the given type. For campaign groups, an
// empty Campaign is attached; otherwise a fresh single Game is attached.
func NewGroup(id, name string, gameType GameType, playerNames []string) Group {
	now := time.Now().UTC()
	g := Group{
		ID:        id,
		Name:      name,
		GameType:  gameType,
		Players:   playerNames,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if gameType == GameTypeCampaign {
		g.Campaign = &Campaign{Chapters: []Chapter{}}
	} else {
		game := NewGame(id+"-game", gameType, playerNames)
		g.Game = &game
	}
	return g
}
