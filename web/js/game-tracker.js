(function () {
  const RESOURCE_META = {
    berry: { icon: "🍓", label: "Berry" },
    twig: { icon: "🪵", label: "Twig" },
    resin: { icon: "🟤", label: "Resin" },
    pebble: { icon: "⚪", label: "Pebble" },
    stopwatch: { icon: "⏱️", label: "Stopwatch" },
  };

  const SCORE_CATEGORIES = ["cards", "events", "journey", "prosperity", "stopwatches"];

  // Renders a live game tracker into `root`.
  // opts: { title, game, onChange(updatedGame), onBack(), onComplete(updatedGame), hideComplete }
  function renderGameTracker(root, opts) {
    let game = opts.game;

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
        const updated = EDWasm.mutateGame(game, "completeGame", {});
        game = updated;
        opts.onChange(game);
        if (opts.onComplete) opts.onComplete(game);
        draw();
      });

      // Markers (sun / moon)
      view.querySelectorAll(".marker-row").forEach((row) => {
        const name = row.dataset.marker;
        const value = (game.trackers.markers && game.trackers.markers[name]) || 0;
        row.querySelector(".marker-value").textContent = value;
        row.querySelector(".marker-dec").addEventListener("click", () =>
          mutate("setMarker", { name, position: value - 1 })
        );
        row.querySelector(".marker-inc").addEventListener("click", () =>
          mutate("setMarker", { name, position: value + 1 })
        );
      });

      // Skunks
      const skunkList = view.querySelector(".skunk-list");
      (game.trackers.skunks || []).forEach((position, i) => {
        const li = document.createElement("li");
        li.className = "skunk-row";
        li.innerHTML = `🦨 <span class="skunk-value">${position}</span>`;
        const dec = document.createElement("button");
        dec.className = "btn btn-tiny";
        dec.textContent = "-";
        dec.addEventListener("click", () => mutate("setSkunkPosition", { index: i, position: position - 1 }));
        const inc = document.createElement("button");
        inc.className = "btn btn-tiny";
        inc.textContent = "+";
        inc.addEventListener("click", () => mutate("setSkunkPosition", { index: i, position: position + 1 }));
        const rm = document.createElement("button");
        rm.className = "btn btn-tiny btn-danger";
        rm.textContent = "✕";
        rm.addEventListener("click", () => mutate("removeSkunk", { index: i }));
        li.append(dec, inc, rm);
        skunkList.appendChild(li);
      });
      view.querySelector(".btn-add-skunk").addEventListener("click", () => mutate("addSkunk", { position: 0 }));

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
        SCORE_CATEGORIES.forEach((category) => {
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
