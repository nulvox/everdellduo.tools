// Command wasm is the syscall/js bridge between the browser and the pure
// Go logic in internal/model, internal/scorecard, and internal/store. All
// data crosses the boundary as JSON strings. Every exposed function returns
// a JSON envelope: {"ok": true, "data": ...} or {"ok": false, "error": "..."}.
package main

import (
	"encoding/json"
	"syscall/js"
	"time"

	"everdellduo.tools/internal/model"
	"everdellduo.tools/internal/scorecard"
	"everdellduo.tools/internal/store"
)

func ok(data interface{}) string {
	b, err := json.Marshal(struct {
		OK   bool        `json:"ok"`
		Data interface{} `json:"data"`
	}{true, data})
	if err != nil {
		return `{"ok":false,"error":"failed to marshal response"}`
	}
	return string(b)
}

func fail(err error) string {
	b, _ := json.Marshal(struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}{false, err.Error()})
	return string(b)
}

func failMsg(msg string) string {
	b, _ := json.Marshal(struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}{false, msg})
	return string(b)
}

func arg(args []js.Value, i int) string {
	if i >= len(args) {
		return ""
	}
	return args[i].String()
}

func wrap(fn func(args []js.Value) string) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return fn(args)
	})
}

// newGroup(name, gameType, playersJSON, difficulty) -> Group
func newGroup(args []js.Value) string {
	name := arg(args, 0)
	gameType := model.GameType(arg(args, 1))
	var players []string
	if err := json.Unmarshal([]byte(arg(args, 2)), &players); err != nil {
		return fail(err)
	}
	difficulty := model.Difficulty(arg(args, 3))
	id := time.Now().UTC().Format("20060102T150405.000000000")
	g := model.NewGroup(id, name, gameType, players, difficulty)
	return ok(g)
}

// startChapterGame(campaignJSON, chapterIndex, gameType, playersJSON, startingStopwatches) -> Campaign
func startChapterGame(args []js.Value) string {
	var campaign model.Campaign
	if err := json.Unmarshal([]byte(arg(args, 0)), &campaign); err != nil {
		return fail(err)
	}
	var chapterIndex int
	if err := json.Unmarshal([]byte(arg(args, 1)), &chapterIndex); err != nil {
		return fail(err)
	}
	gameType := model.GameType(arg(args, 2))
	var players []string
	if err := json.Unmarshal([]byte(arg(args, 3)), &players); err != nil {
		return fail(err)
	}
	var startingStopwatches int
	if err := json.Unmarshal([]byte(arg(args, 4)), &startingStopwatches); err != nil {
		return fail(err)
	}
	id := time.Now().UTC().Format("20060102T150405.000000000")
	if _, err := model.StartChapterGame(&campaign, chapterIndex, id, gameType, players, startingStopwatches); err != nil {
		return fail(err)
	}
	return ok(campaign)
}

// addChapter(campaignJSON, name, bonusStopwatch) -> Campaign
func addChapter(args []js.Value) string {
	var campaign model.Campaign
	if err := json.Unmarshal([]byte(arg(args, 0)), &campaign); err != nil {
		return fail(err)
	}
	var bonusStopwatch bool
	if err := json.Unmarshal([]byte(arg(args, 2)), &bonusStopwatch); err != nil {
		return fail(err)
	}
	model.AddChapter(&campaign, arg(args, 1), bonusStopwatch)
	return ok(campaign)
}

// setChapterBonusStopwatch(campaignJSON, chapterIndex, bonusStopwatch) -> Campaign
func setChapterBonusStopwatch(args []js.Value) string {
	var campaign model.Campaign
	if err := json.Unmarshal([]byte(arg(args, 0)), &campaign); err != nil {
		return fail(err)
	}
	var chapterIndex int
	if err := json.Unmarshal([]byte(arg(args, 1)), &chapterIndex); err != nil {
		return fail(err)
	}
	var bonusStopwatch bool
	if err := json.Unmarshal([]byte(arg(args, 2)), &bonusStopwatch); err != nil {
		return fail(err)
	}
	if err := model.SetChapterBonusStopwatch(&campaign, chapterIndex, bonusStopwatch); err != nil {
		return fail(err)
	}
	return ok(campaign)
}

// removeChapter(campaignJSON, chapterIndex) -> Campaign
func removeChapter(args []js.Value) string {
	var campaign model.Campaign
	if err := json.Unmarshal([]byte(arg(args, 0)), &campaign); err != nil {
		return fail(err)
	}
	var chapterIndex int
	if err := json.Unmarshal([]byte(arg(args, 1)), &chapterIndex); err != nil {
		return fail(err)
	}
	if err := model.RemoveChapter(&campaign, chapterIndex); err != nil {
		return fail(err)
	}
	return ok(campaign)
}

// removeChapterGame(campaignJSON, chapterIndex, gameIndex) -> Campaign
func removeChapterGame(args []js.Value) string {
	var campaign model.Campaign
	if err := json.Unmarshal([]byte(arg(args, 0)), &campaign); err != nil {
		return fail(err)
	}
	var chapterIndex, gameIndex int
	if err := json.Unmarshal([]byte(arg(args, 1)), &chapterIndex); err != nil {
		return fail(err)
	}
	if err := json.Unmarshal([]byte(arg(args, 2)), &gameIndex); err != nil {
		return fail(err)
	}
	if err := model.RemoveChapterGame(&campaign, chapterIndex, gameIndex); err != nil {
		return fail(err)
	}
	return ok(campaign)
}

// mutateGame applies a named operation to a Game and returns the updated Game.
// mutateGame(gameJSON, op, argsJSON) -> Game
func mutateGame(args []js.Value) string {
	var game model.Game
	if err := json.Unmarshal([]byte(arg(args, 0)), &game); err != nil {
		return fail(err)
	}
	op := arg(args, 1)
	payload := arg(args, 2)

	switch op {
	case "adjustResource":
		var p struct {
			Player   string `json:"player"`
			Resource string `json:"resource"`
			Delta    int    `json:"delta"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		if err := model.AdjustResource(&game, p.Player, p.Resource, p.Delta); err != nil {
			return fail(err)
		}
	case "setScoreBreakdown":
		var p struct {
			Player   string `json:"player"`
			Category string `json:"category"`
			Value    int    `json:"value"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		if err := model.SetScoreBreakdownEntry(&game, p.Player, p.Category, p.Value); err != nil {
			return fail(err)
		}
	case "addEvent":
		var p struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		model.AddEvent(&game, p.Name)
	case "removeEvent":
		var p struct {
			Index int `json:"index"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		model.RemoveEvent(&game, p.Index)
	case "setEventAchieved":
		var p struct {
			Index      int    `json:"index"`
			Achieved   bool   `json:"achieved"`
			AchievedBy string `json:"achievedBy"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		model.SetEventAchieved(&game, p.Index, p.Achieved, p.AchievedBy)
	case "adjustMarker":
		var p struct {
			Name  string `json:"name"`
			Delta int    `json:"delta"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		model.AdjustMarker(&game, p.Name, p.Delta)
	case "addSkunk":
		var p struct {
			Zone string `json:"zone"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		if err := model.AddSkunk(&game, model.SkunkZone(p.Zone)); err != nil {
			return fail(err)
		}
	case "setSkunkPosition":
		var p struct {
			Index    int `json:"index"`
			Position int `json:"position"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		if err := model.SetSkunkPosition(&game, p.Index, p.Position); err != nil {
			return fail(err)
		}
	case "setSkunkZone":
		var p struct {
			Index int    `json:"index"`
			Zone  string `json:"zone"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		if err := model.SetSkunkZone(&game, p.Index, model.SkunkZone(p.Zone)); err != nil {
			return fail(err)
		}
	case "removeSkunk":
		var p struct {
			Index int `json:"index"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		model.RemoveSkunk(&game, p.Index)
	case "advanceSeason":
		model.AdvanceSeason(&game)
	case "adjustTeamStopwatches":
		var p struct {
			Delta int `json:"delta"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return fail(err)
		}
		model.AdjustTeamStopwatches(&game, p.Delta)
	case "completeGame":
		model.CompleteGame(&game)
	default:
		return failMsg("unknown operation: " + op)
	}
	return ok(game)
}

// computeScorecard(campaignJSON) -> Scorecard
func computeScorecard(args []js.Value) string {
	var campaign model.Campaign
	if err := json.Unmarshal([]byte(arg(args, 0)), &campaign); err != nil {
		return fail(err)
	}
	sc := scorecard.Compute(&campaign)
	return ok(sc)
}

// exportBackup(groupsJSON) -> string (JSON document)
func exportBackup(args []js.Value) string {
	var groups []model.Group
	if err := json.Unmarshal([]byte(arg(args, 0)), &groups); err != nil {
		return fail(err)
	}
	backup := store.NewBackup(groups)
	data, err := store.MarshalBackup(backup)
	if err != nil {
		return fail(err)
	}
	return ok(data)
}

// importBackup(backupJSON) -> Backup
func importBackup(args []js.Value) string {
	backup, err := store.ParseBackup(arg(args, 0))
	if err != nil {
		return fail(err)
	}
	return ok(backup)
}

func main() {
	js.Global().Set("everdellduoNewGroup", wrap(newGroup))
	js.Global().Set("everdellduoAddChapter", wrap(addChapter))
	js.Global().Set("everdellduoSetChapterBonusStopwatch", wrap(setChapterBonusStopwatch))
	js.Global().Set("everdellduoRemoveChapter", wrap(removeChapter))
	js.Global().Set("everdellduoRemoveChapterGame", wrap(removeChapterGame))
	js.Global().Set("everdellduoStartChapterGame", wrap(startChapterGame))
	js.Global().Set("everdellduoMutateGame", wrap(mutateGame))
	js.Global().Set("everdellduoComputeScorecard", wrap(computeScorecard))
	js.Global().Set("everdellduoExportBackup", wrap(exportBackup))
	js.Global().Set("everdellduoImportBackup", wrap(importBackup))

	select {} // keep the WASM module alive
}
