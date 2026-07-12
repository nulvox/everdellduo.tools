(function () {
  const RESOURCE_META = {
    worker: { icon: "👷", label: "Workers" },
    berry: { icon: "🍓", label: "Berry" },
    twig: { icon: "🪵", label: "Twig" },
    resin: { icon: "🔶", label: "Resin" },
    pebble: { icon: "🪨", label: "Pebble" },
  };

  const SEASON_META = {
    winter: { icon: "❄️", label: "Winter" },
    spring: { icon: "🌱", label: "Spring" },
    summer: { icon: "☀️", label: "Summer" },
    autumn: { icon: "🍂", label: "Autumn" },
  };

  const SCORE_CATEGORIES = ["cards", "events", "journey", "prosperity", "stopwatches"];
  // Journey scoring doesn't apply to campaign chapters (always coop).
  const CAMPAIGN_SCORE_CATEGORIES = SCORE_CATEGORIES.filter((c) => c !== "journey");

  const SKUNK_ZONE_LABELS = {
    events: "📜 Events",
    meadow: "🌾 Meadow",
    farm: "🚜 Farm",
  };

  // Renders a live game tracker into `root`.
  // opts: { title, game, onChange(updatedGame), onBack(), onComplete(updatedGame), hideComplete, isCampaign }
  function renderGameTracker(root, opts) {
    let game = opts.game;
    const scoreCategories = opts.isCampaign ? CAMPAIGN_SCORE_CATEGORIES : SCORE_CATEGORIES;

    function persist(newGame) {
      game = newGame;
      opts.onChange(game);
      draw();
    }

    function mutate(op, payload) {
      persist(EDWasm.mutateGame(game, op, payload));
    }

    function draw() {
      const view = document.getElementById("tpl-game-tracker").content.cloneNode(true);
      view.querySelector(".group-title").textContent = opts.title;
      view.querySelector(".btn-back").addEventListener("click", () => opts.onBack());

      const season = game.season || "winter";
      const seasonMeta = SEASON_META[season] || { icon: "🌀", label: season };
      view.querySelector(".season-badge").textContent = `${seasonMeta.icon} ${seasonMeta.label}`;

      const completeGame = () => {
        const updated = EDWasm.mutateGame(game, "completeGame", {});
        game = updated;
        opts.onChange(game);
        if (opts.onComplete) opts.onComplete(game);
        draw();
      };

      const completeBtn = view.querySelector(".btn-complete-game");
      if (game.status === "completed") {
        completeBtn.textContent = "✅ Completed";
        completeBtn.disabled = true;
      }
      if (opts.hideComplete) {
        completeBtn.hidden = true;
      }
      completeBtn.addEventListener("click", () => {
        if (!confirm("Mark this game complete? Scores will be locked into the record.")) return;
        completeGame();
      });

      const newSeasonBtn = view.querySelector(".btn-new-season");
      if (game.status === "completed") {
        newSeasonBtn.disabled = true;
      }
      newSeasonBtn.addEventListener("click", () => {
        if (season === "autumn") {
          if (
            !confirm(
              "Autumn just ended — the game is over! Double-check each player's score breakdown " +
                `(${scoreCategories.join(", ")}) before finalizing. Ready to record the game?`
            )
          ) {
            return;
          }
          completeGame();
          return;
        }
        if (
          !confirm(
            "Start a new season? This resets Sun/Moon to 0, re-randomizes the skunks, and resets everyone's workers to 3."
          )
        ) {
          return;
        }
        mutate("advanceSeason", {});
      });

      // Markers (sun / moon)
      view.querySelectorAll(".marker-row[data-marker]").forEach((row) => {
        const name = row.dataset.marker;
        const value = (game.trackers.markers && game.trackers.markers[name]) || 0;
        row.querySelector(".marker-value").textContent = value;
        row.querySelector(".marker-dec").addEventListener("click", () => mutate("adjustMarker", { name, delta: -1 }));
        row.querySelector(".marker-inc").addEventListener("click", () => mutate("adjustMarker", { name, delta: 1 }));
      });

      // Skunks
      const skunkList = view.querySelector(".skunk-list");
      (game.trackers.skunks || []).forEach((skunk, i) => {
        const li = document.createElement("li");
        li.className = "skunk-row";

        const label = document.createElement("span");
        label.textContent = "🦨";
        li.appendChild(label);

        const zoneSelect = document.createElement("select");
        zoneSelect.className = "skunk-zone-select";
        Object.keys(SKUNK_ZONE_LABELS).forEach((zone) => {
          const option = document.createElement("option");
          option.value = zone;
          option.textContent = SKUNK_ZONE_LABELS[zone];
          if (zone === skunk.zone) option.selected = true;
          zoneSelect.appendChild(option);
        });
        zoneSelect.addEventListener("change", () => {
          try {
            mutate("setSkunkZone", { index: i, zone: zoneSelect.value });
          } catch (e) {
            alert(e.message);
          }
        });

        const posValue = document.createElement("span");
        posValue.className = "skunk-value";
        posValue.textContent = skunk.position;

        const dec = document.createElement("button");
        dec.className = "btn btn-tiny";
        dec.textContent = "-";
        dec.addEventListener("click", () => {
          try {
            mutate("setSkunkPosition", { index: i, position: skunk.position - 1 });
          } catch (e) {
            alert(e.message);
          }
        });
        const inc = document.createElement("button");
        inc.className = "btn btn-tiny";
        inc.textContent = "+";
        inc.addEventListener("click", () => {
          try {
            mutate("setSkunkPosition", { index: i, position: skunk.position + 1 });
          } catch (e) {
            alert(e.message);
          }
        });
        const rm = document.createElement("button");
        rm.className = "btn btn-tiny btn-danger";
        rm.textContent = "✕";
        rm.addEventListener("click", () => mutate("removeSkunk", { index: i }));
        li.append(zoneSelect, dec, posValue, inc, rm);
        skunkList.appendChild(li);
      });
      const addSkunkZoneSelect = view.querySelector(".add-skunk-zone");
      view.querySelector(".btn-add-skunk").addEventListener("click", () => {
        mutate("addSkunk", { zone: addSkunkZoneSelect.value });
      });

      // Team stopwatches (campaign chapters only - shared, not per-player)
      if (opts.isCampaign) {
        const stopwatchBlock = view.querySelector(".team-stopwatch-block");
        stopwatchBlock.hidden = false;
        const stopwatchValue = game.teamStopwatches || 0;
        stopwatchBlock.querySelector(".stopwatch-value").textContent = stopwatchValue;
        stopwatchBlock
          .querySelector(".stopwatch-dec")
          .addEventListener("click", () => mutate("adjustTeamStopwatches", { delta: -1 }));
        stopwatchBlock
          .querySelector(".stopwatch-inc")
          .addEventListener("click", () => mutate("adjustTeamStopwatches", { delta: 1 }));

        const starting = game.startingTeamStopwatches || 0;
        const used = Math.max(0, starting - stopwatchValue);
        const bonusKept = starting >= 1 && stopwatchValue >= 1;
        const parts = [`Started with ${starting}.`];
        if (used > 0) parts.push(`Used ${used} so far.`);
        if (bonusKept) parts.push("🎁 Bonus stopwatch kept if the chapter ends now.");
        stopwatchBlock.querySelector(".stopwatch-summary").textContent = parts.join(" ");
      }

      // Events
      const eventList = view.querySelector(".event-list");
      (game.events || []).forEach((event, i) => {
        const li = document.createElement("li");
        li.className = "event-row" + (event.achieved ? " achieved" : "");
        const label = document.createElement("label");
        const checkbox = document.createElement("input");
        checkbox.type = "checkbox";
        checkbox.checked = event.achieved;
        checkbox.addEventListener("change", () => {
          const achievedBy = event.achieved ? "" : game.players[0] ? game.players[0].name : "";
          mutate("setEventAchieved", { index: i, achieved: checkbox.checked, achievedBy });
        });
        const text = document.createElement("span");
        text.textContent = event.name + (event.achieved && event.achievedBy ? ` (${event.achievedBy})` : "");
        label.append(checkbox, text);
        const rm = document.createElement("button");
        rm.className = "btn btn-tiny btn-danger";
        rm.textContent = "✕";
        rm.addEventListener("click", () => mutate("removeEvent", { index: i }));
        li.append(label, rm);
        eventList.appendChild(li);
      });
      view.querySelector(".form-add-event").addEventListener("submit", (ev) => {
        ev.preventDefault();
        const input = ev.target.elements.eventName;
        const name = input.value.trim();
        if (!name) return;
        mutate("addEvent", { name });
      });

      // Players
      const playerCards = view.querySelector(".player-cards");
      (game.players || []).forEach((player) => {
        const card = document.getElementById("tpl-player-card").content.cloneNode(true);
        card.querySelector(".player-name").textContent = player.name;

        const resourceRows = card.querySelector(".resource-rows");
        Object.keys(RESOURCE_META).forEach((resource) => {
          const meta = RESOURCE_META[resource];
          const value = (player.resources && player.resources[resource]) || 0;
          const row = document.getElementById("tpl-resource-row").content.cloneNode(true);
          row.querySelector(".resource-icon").textContent = meta.icon;
          row.querySelector(".resource-label").textContent = meta.label;
          row.querySelector(".resource-value").textContent = value;
          row.querySelector(".res-dec").addEventListener("click", () =>
            mutate("adjustResource", { player: player.name, resource, delta: -1 })
          );
          row.querySelector(".res-inc").addEventListener("click", () => {
            try {
              mutate("adjustResource", { player: player.name, resource, delta: 1 });
            } catch (e) {
              alert(e.message);
            }
          });
          resourceRows.appendChild(row);
        });

        const breakdown = card.querySelector(".score-breakdown");
        scoreCategories.forEach((category) => {
          const value = (player.scoreBreakdown && player.scoreBreakdown[category]) || 0;
          const row = document.getElementById("tpl-score-row").content.cloneNode(true);
          row.querySelector(".score-category-label").textContent = category;
          const input = row.querySelector(".score-category-input");
          input.value = value;
          input.addEventListener("change", () => {
            mutate("setScoreBreakdown", { player: player.name, category, value: parseInt(input.value, 10) || 0 });
          });
          breakdown.appendChild(row);
        });

        card.querySelector(".score-value").textContent = player.score || 0;
        playerCards.appendChild(card);
      });

      const notes = view.querySelector(".game-notes");
      notes.value = game.notes || "";
      notes.addEventListener("change", () => {
        game = { ...game, notes: notes.value };
        opts.onChange(game);
      });

      root.innerHTML = "";
      root.appendChild(view);
    }

    draw();
  }

  window.EDGameTracker = { renderGameTracker, RESOURCE_META };
})();
