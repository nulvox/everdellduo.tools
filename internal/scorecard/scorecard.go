// Package scorecard computes running campaign scorecards from a
// model.Campaign's recorded chapter games. It is pure Go and fully
// testable.
package scorecard

import "everdellduo.tools/internal/model"

// ChapterResult summarizes one completed game's contribution to the
// scorecard.
type ChapterResult struct {
	ChapterName string         `json:"chapterName"`
	GameID      string         `json:"gameId"`
	Scores      map[string]int `json:"scores"`
}

// Scorecard is the running total across a campaign, plus the per-game
// history that produced it.
type Scorecard struct {
	Totals  map[string]int  `json:"totals"`
	History []ChapterResult `json:"history"`
}

// Compute walks every chapter's games, in order, and sums the score of each
// completed game per player name into a running total. In-progress games
// are ignored since they have not been recorded yet.
func Compute(c *model.Campaign) Scorecard {
	sc := Scorecard{Totals: map[string]int{}, History: []ChapterResult{}}
	if c == nil {
		return sc
	}
	for _, chapter := range c.Chapters {
		for _, game := range chapter.Games {
			if game.Status != model.StatusCompleted {
				continue
			}
			result := ChapterResult{ChapterName: chapter.Name, GameID: game.ID, Scores: map[string]int{}}
			for _, p := range game.Players {
				result.Scores[p.Name] = p.Score
				sc.Totals[p.Name] += p.Score
			}
			sc.History = append(sc.History, result)
		}
	}
	return sc
}
