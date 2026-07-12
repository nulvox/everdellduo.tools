# TODO — Everdellduo.tools Incremental Build

Work through remaining tasks in order. Each task is a small, testable
increment. Do NOT skip ahead. Mark `[x]` only after all tests pass and
linting is clean.

---

## Phase 0: Project Skeleton — DONE

- [x] **0.1** Initialize Go module (`go mod init everdellduo.tools`), create directory structure (`cmd/wasm/`, `internal/model/`, `internal/scorecard/`, `internal/store/`, `web/css/`, `web/js/`).
- [x] **0.2** Create `Makefile` (build/test/lint/serve/clean), `.golangci.yml`, `.gitignore`, `LICENSE`, `.github/workflows/pages.yml` — all mirrored from `fate4.tools`.
- [x] **0.3** `make build` produces `docs/main.wasm`, `docs/wasm_exec.js`, and copies `web/` assets into `docs/`.
- [x] **0.4** `golangci-lint run ./...` passes cleanly; `gofmt -l .` prints nothing.

## Phase 1: Core Data Model (Pure Go, Fully Testable) — DONE

- [x] **1.1** `model.Group`, `model.Game`, `model.PlayerState`, `model.Event`, `model.Trackers` — see `internal/model/model.go`. Tests in `model_test.go`.
- [x] **1.2** `model.Campaign` / `model.Chapter` nesting for campaign-type groups.
- [x] **1.3** Mutation helpers: `AdjustResource` (0-floor, stopwatch capped at 2), `SetScoreBreakdownEntry` (auto-sums total), event add/remove/toggle, `AdjustMarker` (clamped to [0,6], skips occupied meadow-skunk positions), zoned skunk add/reposition/rezone/remove (`AddSkunk`, `SetSkunkPosition`, `SetSkunkZone`, `RemoveSkunk`, all uniqueness-checked per zone), `CompleteGame`, `AddChapter`, `StartChapterGame`. Tests in `mutate_test.go`.

## Phase 2: Scorecard & Backup (Pure Go, Fully Testable) — DONE

- [x] **2.1** `internal/scorecard`: `Compute(campaign)` sums each **completed** chapter game's player scores into a running total + history. Tests cover in-progress games being excluded.
- [x] **2.2** `internal/store`: `Backup`/`NewBackup`/`MarshalBackup`/`ParseBackup` with a `Version` field and deterministic (sorted) group ordering for diffable exports. Tests cover round-trip, garbage input, and future-version rejection.

## Phase 3: WASM Bridge — DONE

- [x] **3.1** `cmd/wasm/main.go` exposes `everdellduoNewGroup`, `everdellduoAddChapter`, `everdellduoStartChapterGame`, `everdellduoMutateGame` (op-dispatch: adjustResource, setScoreBreakdown, addEvent, removeEvent, setEventAchieved, adjustMarker, addSkunk, setSkunkPosition, setSkunkZone, removeSkunk, completeGame), `everdellduoComputeScorecard`, `everdellduoExportBackup`, `everdellduoImportBackup`. All return a `{ok, data|error}` JSON envelope.
- [x] **3.2** Verified via `GOOS=js GOARCH=wasm go build` and headless-Firefox smoke tests exercising create-group, resource steppers, stopwatch cap error, events, sun/moon markers (including capping at 6 and hopping over occupied meadow-skunk positions), zoned skunk add/reposition/rezone with collision alerts, campaign chapters, chapter game scoring, and the running scorecard.

## Phase 4: Frontend — DONE (MVP)

- [x] **4.1** `web/js/wasm-loader.js` — loads `main.wasm`, wraps every exported function with envelope unwrapping (throws on `{ok:false}`).
- [x] **4.2** `web/js/storage.js` — per-group `localStorage` keys + manifest key; `allGroupsFull()`/`replaceAll()` for backups.
- [x] **4.3** `web/js/group-list.js` — group list view (open/export/delete) + new-group form (name, game type picker, comma-separated players).
- [x] **4.4** `web/js/game-tracker.js` — live tracker: resource steppers, score-breakdown inputs (auto-totaled), events, sun/moon markers (capped at 6, hop over occupied meadow skunks), zoned skunk markers (zone selector + position steppers + add-skunk zone picker, collision errors surfaced as alerts), notes, complete-game action.
- [x] **4.5** `web/js/campaign-view.js` — running scorecard table, chapter list, add-chapter form, per-chapter start-game (mode picker), chapter game history with reopen.
- [x] **4.6** `web/js/import-export.js` + `web/js/theme.js` — header export/import backup buttons, day/night theme toggle.
- [x] **4.7** `web/css/style.css` — woodland/critter theme, day + night palettes via `[data-theme]`.

## Phase 4.5: Seasons & Workers — DONE

- [x] **4.8** `worker` resource (default 3, reset each season) added to the model, ordered first in `DefaultResourceOrder`; pebble/resin emoji updated to 🪨/🔶.
- [x] **4.9** `Game.Season` (winter/spring/summer/autumn) + `NextSeason`/`AdvanceSeason`/`RandomizeSkunks(WithRand)` in `internal/model`, with tests covering progression, nudge/wrap-on-collision randomization, never-on-zero placement, and the autumn "game over" signal.
- [x] **4.10** WASM `advanceSeason` op; frontend season badge + "New season" button (disabled once the game is completed) that either runs the reset sequence or, in autumn, prompts for a final score check and completes the game.

## Phase 5: Polish (Not Yet Started)

- [x] **5.1** Add delete affordances: `RemoveChapter`/`RemoveChapterGame` in `internal/model` (tested), `everdellduoRemoveChapter`/`everdellduoRemoveChapterGame` WASM ops, and 🗑 Delete chapter / ✕ delete-game buttons in the campaign view, each with a confirmation prompt. (Group deletion already existed via the group list.) Chapter *renaming* is still open.
- [ ] **5.2** Add player add/remove UI within an existing group (currently players are fixed at group creation).
- [ ] **5.3** Add a confirmation step before `Complete game` when required score-breakdown categories are still at 0, to catch accidental early completion.
- [ ] **5.4** Add a print-friendly campaign scorecard layout in `print.css` (chapter-by-chapter breakdown, not just the running total table).
- [ ] **5.5** Add an automated headless-browser smoke test script under `scripts/` (checked into the repo) so `make test` can optionally run a full UI round-trip, not just Go unit tests.
- [ ] **5.6** Consider adding richer default event/resource templates per game type once real box/rulebook data can be verified (current defaults are intentionally generic/adjustable rather than hardcoded to unverified specifics).
- [ ] **5.7** Push the initial commit to `github.com/nulvox/everdellduo.tools` and confirm GitHub Pages deployment via the `pages.yml` workflow.
