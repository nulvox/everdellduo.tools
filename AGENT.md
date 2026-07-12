# Agent Brief: Everdellduo.tools Development

You are building a browser-based Everdell Duo group/campaign/live-game tracker. Read `CLAUDE.md` thoroughly before starting — it contains the full spec, architecture, data model, and mandatory development rules.

## Your Development Process

You MUST follow this exact loop for every task in `TODO.md`:

```
1. Read the current task from TODO.md
2. Write a failing test for the task's acceptance criteria
3. Run the test — confirm it fails for the expected reason
4. Write the minimal production code to pass the test
5. Run ALL tests — confirm everything passes
6. Lint ALL changed files
7. Verify the feature works (for UI tasks: a headless-browser smoke test, or describe what to check manually)
8. Mark the task [x] in TODO.md
9. Commit with a message referencing the task
10. Move to the next task
```

**Never write production code without a failing test first.**
**Never mark a task done if any test fails or any lint error exists.**

## Linting Commands

```bash
GOOS=js GOARCH=wasm golangci-lint run ./...
gofmt -l .   # must print nothing
```

## Build Commands

```bash
make build   # GOOS=js GOARCH=wasm go build -o docs/main.wasm ./cmd/wasm/, copies wasm_exec.js + web/ assets into docs/
make serve   # build + python3 -m http.server 8080 --directory docs
make test    # go test ./internal/...
```

## Testing Go WASM Code

`internal/model`, `internal/scorecard`, and `internal/store` are pure Go with **no** `syscall/js` imports, so `go test ./internal/...` runs directly, no browser needed. Only `cmd/wasm/main.go` imports `syscall/js`; it has no direct unit tests (it cannot compile for the host GOOS/GOARCH) — verify it via `make build` + a browser smoke test instead.

## Architecture Guidance

Keep the same separation as `fate4.tools`:

```
internal/  — Pure Go. No JS dependencies. Fully testable with `go test`.
cmd/wasm/main.go — Bridge layer. Imports syscall/js. Registers JSON-envelope
                   functions as JS globals. Never contains game logic itself.
web/js/    — Frontend JS. Calls the WASM-exposed functions, renders DOM from
             <template> elements in index.html, persists to localStorage.
```

### Data Flow

```
User clicks a stepper / toggles a checkbox / submits a form
  → JS event handler reads current Game/Campaign JSON from in-memory state
  → JS calls the relevant everdellduo* WASM function
  → Go validates and returns the updated JSON (or an error)
  → JS re-renders the affected view from the returned state
  → JS persists the updated Group to localStorage
```

### Key Design Decisions

1. **Go WASM as pure logic engine**: all mutation and validation (resource clamping, stopwatch cap, scorecard computation, backup versioning) happens in Go. JS never mutates game data directly.
2. **JSON as the bridge format**: every WASM call takes/returns JSON strings wrapped in an `{ok, data|error}` envelope.
3. **localStorage, not cookies**: one key per group (`everdellduo_group_{id}`), plus a manifest key (`everdellduo_manifest`) for listing without loading every group.
4. **Backups are the durable recovery path**: `localStorage` can be cleared by the browser at any time. The header "Export backup" button and per-group export buttons are not optional nice-to-haves — treat backup/restore round-tripping as core functionality with test coverage in `internal/store`.
5. **Flexible data model for future formats**: resources are a `map[string]int`, not fixed struct fields; `GameType` is a string, not a closed enum; trackers use a markers map plus a slice of zoned skunks. When adding a new Everdell format, prefer extending these generic containers over adding new fixed fields.
6. **GitHub Pages via `docs/`**: `make build` outputs everything needed to `docs/`, which is what GitHub Pages serves from `main`.

## Season Mechanics

- A game tracks a `Season` (`winter` -> `spring` -> `summer` -> `autumn`, always starting at `winter`). The "New season" button calls `AdvanceSeason`, which: resets sun/moon to 0, re-randomizes all skunk positions within their zones (1-6, never 0, nudging/wrapping on collision), resets every player's `worker` resource to `StartingWorkers` (3), and advances to the next season.
- If the game is already in `autumn` when "New season" is clicked, there is no fifth season to advance to — this is the game-over trigger. The frontend prompts the player to review score-breakdown entries, then calls `completeGame` directly instead of `advanceSeason`.

## Everdell Duo Reference Notes

- Game types offered when creating a group: `challenge`, `duel`, `coop`, `campaign`.
- Only `campaign` groups have multiple games; all other types are exactly one game per group.
- Campaigns are chapters; each chapter can be (re)played as one or more game attempts; only **completed** attempts count toward the running scorecard. Chapters and individual game attempts can be deleted (`RemoveChapter`/`RemoveChapterGame`), removing them from the scorecard too; groups are deleted from the group list.
- Campaign chapters are **always co-op** — there's no game-type picker when starting a chapter game.
- A campaign `Group` has a `Difficulty` (`normal`/`hard`), and each `Chapter` has a `BonusStopwatch` flag the player sets (which chapters grant a bonus varies by campaign/box, so it isn't hardcoded). Stopwatches in a campaign chapter are a **shared team pool**, not per-player: `ComputeStartingStopwatches` gives Normal a 1-stopwatch baseline plus any chapter bonus, Hard only the chapter bonus. `AdjustTeamStopwatches` changes the live count; `StopwatchesUsed`/`StopwatchBonusKept` derive usage/bonus-kept for display (started with ≥ 1 and still ≥ 1) — the app records/shows this but does not invent a point value, since the exact bonus value isn't something we can verify.
- `journey` is not a valid score-breakdown category for campaign chapter games (co-op doesn't use it); it's still offered for standalone Duel/Challenge/Co-op groups.
- Skunk markers are specific to Challenge mode and roam three zones — `events`, `meadow`, and `farm` — with a position that only has to be unique within its own zone, hence they're modeled as a slice of `{zone, position}`, not a fixed field or bare int.
- A skunk parked in the `meadow` zone blocks that space for the sun/moon markers: `AdjustMarker` skips over occupied meadow positions one logical step at a time, and the sun/moon track is hard-capped at position 6 (`MaxMarkerPosition`).
- Exact board-marker step counts vary by print run; the app lets players freely increment/decrement rather than hardcoding rulebook specifics it can't verify.

## What "Done" Looks Like

The finished application:
1. Loads in a browser from static files (GitHub Pages compatible)
2. Lets a user start a new group, choose a game type, and name players
3. Tracks exactly one game for challenge/duel/coop groups
4. Tracks multiple chapters and multiple game attempts per chapter for campaign groups, with an automatically computed running scorecard
5. Provides a live tracker for resources, events (with achieved state), sun/moon/skunk positions, and per-player score breakdowns
6. Auto-saves to localStorage on every edit
7. Exports/imports full backups (and single-group exports) as downloadable/uploadable `.json` files
8. Has fun, legible woodland theming with a day/night toggle
9. All Go code has test coverage, all linting passes
