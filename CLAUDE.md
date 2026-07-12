# Everdellduo.tools — Everdell Duo Group & Campaign Tracker

## Project Overview

A browser-based tracker for [Everdell Duo](https://boardgamegeek.com/boardgame/425005/everdell-duo) groups, campaigns, and live games. Built with Go compiled to WebAssembly for core logic, with an HTML/CSS/JS frontend. Hosted as static files on GitHub Pages. Designed with an open mind toward tracking other Everdell formats in the future — data types favor flexible maps/slices over rigid enums where the underlying rules might vary by format or edition.

## Tech Stack

- **Core logic**: Go 1.23+, compiled to `GOOS=js GOARCH=wasm`
- **Frontend**: Vanilla HTML5, CSS3, JavaScript (no framework)
- **Storage**: `localStorage` (JSON-serialized groups, one key per group) plus a lightweight manifest key for listing
- **Backups**: JSON import/export of the full group set (or a single group) — this is the durable recovery path, since `localStorage` alone is not backed up anywhere
- **Build**: Makefile; output is a `docs/` directory for GitHub Pages
- **Testing**: Go standard `testing` package for all pure-logic packages
- **Linting**: `golangci-lint` for Go; browser-console-clean JS (no lint warnings)

## Architecture

Go WASM serves as the logic backend. It exposes functions to JavaScript via `syscall/js`, each returning a JSON envelope `{"ok":true,"data":...}` or `{"ok":false,"error":"..."}`:

- `everdellduoNewGroup(name, gameType, playersJSON)` — create a new Group
- `everdellduoAddChapter(campaignJSON, name)` — append a chapter to a campaign
- `everdellduoStartChapterGame(campaignJSON, chapterIndex, gameType, playersJSON)` — start a new game attempt within a chapter
- `everdellduoMutateGame(gameJSON, op, payloadJSON)` — apply one of a fixed set of mutations to a live Game (adjustResource, setScoreBreakdown, addEvent, removeEvent, setEventAchieved, adjustMarker, addSkunk, setSkunkPosition, setSkunkZone, removeSkunk, advanceSeason, completeGame)
- `everdellduoComputeScorecard(campaignJSON)` — compute the running per-player scorecard from a campaign's completed games
- `everdellduoExportBackup(groupsJSON)` / `everdellduoImportBackup(backupJSON)` — durable backup serialization

The JS frontend handles all DOM rendering, event binding, and `localStorage` persistence. It calls into Go WASM for all data mutations and derived computations (scorecards, backups) so the rules logic has a single, tested source of truth.

## Development Workflow

### MANDATORY: Test-Driven Development

1. **Write the test first.** No production code without a failing test.
2. **Run the test, confirm it fails** for the right reason.
3. **Write the minimal code** to make the test pass.
4. **Refactor** only after green.
5. **Lint after every change** to a `.go`, `.js`, `.html`, or `.css` file.

### MANDATORY: Incremental TODO-Driven Progress

- Work from `TODO.md`. Each task is a small, testable increment.
- Mark tasks `[x]` only after ALL tests pass AND linting is clean.
- Never skip ahead. Complete tasks in order.

### MANDATORY: Definition of Done

A task is NOT done until:
- [ ] All existing tests pass (`go test ./internal/...`)
- [ ] New tests cover the new behavior
- [ ] `golangci-lint run` passes with zero issues
- [ ] The feature works in a browser (manual verification, ideally via a headless-browser smoke test)
- [ ] `TODO.md` is updated

## Package Management

- Use `uv` for any Python tooling needs (never `pip`).
- Use standard Go modules (`go mod tidy`).

## Project Structure

```
everdellduo.tools/
  CLAUDE.md          # This file
  AGENT.md           # Agent development prompt/brief
  TODO.md            # Incremental task list
  go.mod
  cmd/
    wasm/
      main.go        # WASM entry point, JS bridge (JSON envelope in/out)
  internal/
    model/
      model.go       # Core data types: Group, Game, Campaign, Chapter, PlayerState, Event, Trackers
      model_test.go
      mutate.go       # Mutation helpers (resources, events, markers, skunks, chapters)
      mutate_test.go
    scorecard/
      scorecard.go    # Computes running campaign scorecards from completed games
      scorecard_test.go
    store/
      store.go        # Backup/manifest serialization format + versioning
      store_test.go
  web/
    index.html        # Main HTML shell + <template> elements for every view
    css/
      style.css        # Woodland theme (day/night via [data-theme])
      print.css
    js/
      wasm-loader.js   # Loads main.wasm, wraps exported functions with envelope handling
      storage.js       # localStorage manifest + per-group persistence
      theme.js         # Day/night toggle
      import-export.js # Backup download/upload
      group-list.js    # Group list + "new group" form views
      game-tracker.js  # Live game tracker view (resources, events, trackers, scores)
      campaign-view.js # Campaign overview: scorecard + chapters + chapter games
      app.js           # Router tying views together
  docs/               # Built output for GitHub Pages
  Makefile
```

## Data Model

- **Group**: top-level durable artifact — id, name, gameType (`challenge`/`duel`/`coop`/`campaign`, but stored as a plain string so other formats can be added later), players, timestamps, and either a `Game` (non-campaign) or a `Campaign` (campaign type).
- **Game**: live/current game state — players (each with a resource map and score/score breakdown), events, board trackers, status, timestamps, notes.
- **Resources**: a `map[string]int` per player, not a fixed struct, so different formats can use different resource sets. Defaults: `worker`, `berry`, `twig`, `resin`, `pebble`, `stopwatch`.
  - `stopwatch` is a **scored resource capped at 0-2**, not a live countdown timer. It is recorded into the score breakdown like any other category.
  - `worker` starts at `StartingWorkers` (3) and is reset to 3 for every player each time `AdvanceSeason` runs — it represents workers available to send out that season, not a scored total.
- **Events**: `{name, achieved, achievedBy}` — free-form list, add/remove/toggle at will.
- **Trackers**: `{markers: map[string]int, skunks: []Skunk}`. `markers` defaults to `sun` and `moon` at position 0 and is clamped to `[0,6]` (`internal/model.MinMarkerPosition`/`MaxMarkerPosition`). Each `Skunk` is `{zone, position}` where `zone` is one of `events`/`meadow`/`farm`; a skunk's position only has to be unique among other skunks in the *same* zone. Meadow skunks are special: `AdjustMarker` moves the sun/moon one logical step at a time and skips over any position occupied by a meadow skunk, so the marker "hops" them. `AddSkunk`/`SetSkunkZone` auto-place at the next free position in the target zone; `SetSkunkPosition` rejects a move onto an already-occupied position in the same zone (`ErrSkunkPositionTaken`).
- **Campaign**: an ordered list of `Chapter`s, each with a name and a list of `Game` attempts. The running scorecard is *computed*, not stored, by summing each **completed** game's player scores across all chapters (`internal/scorecard`). In-progress attempts are ignored until completed.
- **Season**: a `Game` tracks a `Season` (`winter`/`spring`/`summer`/`autumn`, always starting at `winter`). `AdvanceSeason` performs the full "new season" sequence: reset sun/moon markers to 0, re-randomize every skunk's position within its zone (`RandomizeSkunks`, positions 1-6, meadow-blocking rules from `Trackers` still apply to markers, not to the randomization itself), reset every player's `worker` resource to `StartingWorkers`, and advance to the next season. If the game is already in `autumn`, `AdvanceSeason` makes no changes and returns `true` ("game over") — the frontend uses this to prompt the player to double-check score-breakdown entries before calling `completeGame` instead of starting a nonexistent fifth season.

## Key UX Requirements

- **Group list**: the landing view. Shows all groups with type badge, last-updated time, and Open/Export/Delete actions.
- **New group form**: name, game type picker (Challenge/Duel/Co-op/Campaign), comma-separated player list.
- **Live game tracker**: per-player resource +/- steppers, event add/remove/achieve-toggle, sun/moon marker +/- steppers, skunk add/remove/reposition, per-player score-breakdown inputs (auto-summed into total), free-form notes, and a "Complete game" action that locks in the result.
- **Campaign view**: running scorecard table, chapter list with "add chapter" and per-chapter "start game" (choosing the mode for that chapter), and a list of that chapter's game attempts (status + summary) that can be reopened.
- **Backups**: header-level "Export backup" (all groups) and "Import backup" (replaces all groups, with a confirmation prompt) plus a per-group "Export" button on each group card.
- **Theming**: woodland/critter theme with a day/night toggle persisted in `localStorage`.
