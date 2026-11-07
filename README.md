# Everdellduo.tools

A browser-based group, campaign, and live-game tracker for [Everdell Duo](https://boardgamegeek.com/boardgame/425005/everdell-duo), built with an open mind toward tracking other Everdell formats down the line. Built with Go compiled to WebAssembly for core logic, with a vanilla HTML/CSS/JS frontend. No server required — runs entirely in the browser.

**Live site:** [https://nulvox.github.io/everdellduo.tools](https://nulvox.github.io/everdellduo.tools)

## Features

- Start a **group** and choose its game type: Challenge, Duel, Co-op, or Campaign
- Challenge/Duel/Co-op groups track exactly one game
- Campaign groups track multiple chapters, each played as one or more games, with a running scorecard rolled up automatically from completed games
- Live in-game tracker: per-player resources (berry, twig, resin, pebble, stopwatch), events with achieved status, sun/moon marker positions, roaming skunk markers, and per-player score breakdowns
- Stopwatches are a scored resource capped at 0-2 per game/chapter (not a countdown timer) and feed into the score breakdown
- JSON export/import for durable backups — your data lives in `localStorage`, so back it up!
- Fun woodland theming, with a day/night toggle

## Architecture

The application is intentionally static and local-first. Pure Go packages own
the game model, validation, scorecard calculations, and versioned backup
format; they compile to WebAssembly for use in the browser. A small vanilla
JavaScript layer renders the UI and persists groups to `localStorage`. There is
no application server and selected game data never leaves the browser.

GitHub Pages serves the generated `docs/` directory. JSON import/export is the
durable backup and transfer path because browser storage can be cleared.

## Build and test

Requires Go 1.23+ and `golangci-lint`.

```sh
make build    # Compile WASM and assemble docs/ for deployment
make test     # Run Go tests
make lint     # Run golangci-lint
make serve    # Build and serve locally at http://localhost:8080
```

Pure Go tests cover the model, mutations, campaign scorecards, and backup
round-trips. `make lint` checks both the host-testable packages and the WASM
build target.

## Usage

Open `docs/index.html` in a browser (or use `make serve`). Click **Start a new group**, pick a game type, and add your players. All data is saved locally in your browser's `localStorage` — use the **Export backup** button regularly to keep a durable copy.

## License

This project is licensed under the [MIT License](LICENSE).

## Disclaimer

This is an unofficial fan-made tool. It is not affiliated with or endorsed by Starling Games.
