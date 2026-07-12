// Loads the Go WASM module and exposes a thin, JSON-envelope-aware wrapper
// around the exported Go functions. All calls are synchronous once loaded
// (Go's js.FuncOf callbacks run to completion immediately in this app),
// but EDWasm.ready() returns a Promise so the app can wait for load.
(function () {
  let readyPromise = null;

  function callEnvelope(fnName, ...args) {
    const raw = window[fnName](...args);
    let envelope;
    try {
      envelope = JSON.parse(raw);
    } catch (e) {
      throw new Error(`${fnName}: invalid response from WASM`);
    }
    if (!envelope.ok) {
      throw new Error(envelope.error || `${fnName} failed`);
    }
    return envelope.data;
  }

  function ready() {
    if (readyPromise) return readyPromise;
    readyPromise = new Promise((resolve, reject) => {
      const go = new Go();
      WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
        .then((result) => {
          go.run(result.instance);
          resolve();
        })
        .catch(reject);
    });
    return readyPromise;
  }

  window.EDWasm = {
    ready,
    newGroup(name, gameType, players, difficulty) {
      return callEnvelope("everdellduoNewGroup", name, gameType, JSON.stringify(players), difficulty);
    },
    addChapter(campaign, name, bonusStopwatch) {
      return callEnvelope("everdellduoAddChapter", JSON.stringify(campaign), name, JSON.stringify(bonusStopwatch));
    },
    setChapterBonusStopwatch(campaign, chapterIndex, bonusStopwatch) {
      return callEnvelope(
        "everdellduoSetChapterBonusStopwatch",
        JSON.stringify(campaign),
        JSON.stringify(chapterIndex),
        JSON.stringify(bonusStopwatch)
      );
    },
    removeChapter(campaign, chapterIndex) {
      return callEnvelope("everdellduoRemoveChapter", JSON.stringify(campaign), JSON.stringify(chapterIndex));
    },
    removeChapterGame(campaign, chapterIndex, gameIndex) {
      return callEnvelope(
        "everdellduoRemoveChapterGame",
        JSON.stringify(campaign),
        JSON.stringify(chapterIndex),
        JSON.stringify(gameIndex)
      );
    },
    startChapterGame(campaign, chapterIndex, gameType, players, startingStopwatches) {
      return callEnvelope(
        "everdellduoStartChapterGame",
        JSON.stringify(campaign),
        JSON.stringify(chapterIndex),
        gameType,
        JSON.stringify(players),
        JSON.stringify(startingStopwatches)
      );
    },
    mutateGame(game, op, payload) {
      return callEnvelope("everdellduoMutateGame", JSON.stringify(game), op, JSON.stringify(payload));
    },
    computeScorecard(campaign) {
      return callEnvelope("everdellduoComputeScorecard", JSON.stringify(campaign));
    },
    exportBackup(groups) {
      return callEnvelope("everdellduoExportBackup", JSON.stringify(groups));
    },
    importBackup(backupText) {
      return callEnvelope("everdellduoImportBackup", backupText);
    },
  };
})();
