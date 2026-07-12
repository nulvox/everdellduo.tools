(function () {
  // Mirrors internal/model.ComputeStartingStopwatches: Normal gets a
  // baseline stopwatch plus any chapter bonus; Hard only gets the bonus.
  function computeStartingStopwatches(difficulty, chapterBonus) {
    let starting = difficulty === "hard" ? 0 : 1;
    if (chapterBonus) starting++;
    return starting;
  }

  function summarizeGame(game) {
    const totals = (game.players || []).map((p) => `${p.name}: ${p.score || 0}`).join(", ");
    const parts = [totals || "(no players)"];
    if (typeof game.startingTeamStopwatches === "number") {
      const starting = game.startingTeamStopwatches;
      const current = game.teamStopwatches || 0;
      const used = Math.max(0, starting - current);
      const bonusKept = starting >= 1 && current >= 1;
      parts.push(`⏱️ started ${starting}${used > 0 ? `, used ${used}` : ""}${bonusKept ? ", 🎁 bonus kept" : ""}`);
    }
    return parts.join(" — ");
  }

  // Renders the campaign overview: scorecard + chapters + game history.
  // opts: { group, onSave(group), onBack(), onOpenChapterGame(chapterIndex, gameIndex) }
  function renderCampaignView(root, opts) {
    let group = opts.group;

    function persist(newGroup) {
      group = newGroup;
      group.updatedAt = new Date().toISOString();
      opts.onSave(group);
      draw();
    }

    function draw() {
      const view = document.getElementById("tpl-campaign-view").content.cloneNode(true);
      view.querySelector(".group-title").textContent = group.name + " (Campaign)";
      const difficulty = group.difficulty === "hard" ? "hard" : "normal";
      view.querySelector(".difficulty-badge").textContent = difficulty === "hard" ? "🔥 Hard" : "🌿 Normal";
      view.querySelector(".btn-back").addEventListener("click", () => opts.onBack());

      const scorecard = EDWasm.computeScorecard(group.campaign);
      const scorecardBody = view.querySelector(".scorecard-body");
      const totals = scorecard.totals || {};
      const names = Object.keys(totals).length ? Object.keys(totals) : group.players;
      names.forEach((name) => {
        const tr = document.createElement("tr");
        tr.innerHTML = `<td>${name}</td><td>${totals[name] || 0}</td>`;
        scorecardBody.appendChild(tr);
      });

      const chapterList = view.querySelector(".chapter-list");
      (group.campaign.chapters || []).forEach((chapter, chapterIndex) => {
        const card = document.getElementById("tpl-chapter-card").content.cloneNode(true);
        card.querySelector(".chapter-name").textContent = chapter.name;

        const bonusCheckbox = card.querySelector(".chapter-bonus-stopwatch");
        bonusCheckbox.checked = !!chapter.bonusStopwatch;
        bonusCheckbox.addEventListener("change", () => {
          const updatedCampaign = EDWasm.setChapterBonusStopwatch(group.campaign, chapterIndex, bonusCheckbox.checked);
          persist({ ...group, campaign: updatedCampaign });
        });

        card.querySelector(".btn-start-game").addEventListener("click", () => {
          // Campaign chapters are always co-op.
          const startingStopwatches = computeStartingStopwatches(difficulty, chapter.bonusStopwatch);
          const updatedCampaign = EDWasm.startChapterGame(
            group.campaign,
            chapterIndex,
            "coop",
            group.players,
            startingStopwatches
          );
          persist({ ...group, campaign: updatedCampaign });
        });
        card.querySelector(".btn-delete-chapter").addEventListener("click", () => {
          const gameCount = (chapter.games || []).length;
          const warning =
            gameCount > 0
              ? `Delete "${chapter.name}"? This also deletes its ${gameCount} game(s) and removes them from the scorecard. This cannot be undone.`
              : `Delete "${chapter.name}"? This cannot be undone.`;
          if (!confirm(warning)) return;
          const updatedCampaign = EDWasm.removeChapter(group.campaign, chapterIndex);
          persist({ ...group, campaign: updatedCampaign });
        });

        const gameList = card.querySelector(".chapter-game-list");
        (chapter.games || []).forEach((game, gameIndex) => {
          const row = document.getElementById("tpl-chapter-game-row").content.cloneNode(true);
          row.querySelector(".game-status").textContent = game.status === "completed" ? "✅" : "▶️";
          row.querySelector(".game-summary").textContent = summarizeGame(game);
          row.querySelector(".btn-resume").addEventListener("click", () =>
            opts.onOpenChapterGame(chapterIndex, gameIndex)
          );
          row.querySelector(".btn-delete-game").addEventListener("click", () => {
            const label =
              game.status === "completed" ? "this completed game (it will be removed from the scorecard)" : "this game";
            if (!confirm(`Delete ${label}? This cannot be undone.`)) return;
            const updatedCampaign = EDWasm.removeChapterGame(group.campaign, chapterIndex, gameIndex);
            persist({ ...group, campaign: updatedCampaign });
          });
          gameList.appendChild(row);
        });

        chapterList.appendChild(card);
      });

      view.querySelector(".form-add-chapter").addEventListener("submit", (ev) => {
        ev.preventDefault();
        const input = ev.target.elements.chapterName;
        const name = input.value.trim();
        if (!name) return;
        const bonusStopwatch = ev.target.elements.bonusStopwatch.checked;
        const updatedCampaign = EDWasm.addChapter(group.campaign, name, bonusStopwatch);
        ev.target.elements.bonusStopwatch.checked = false;
        persist({ ...group, campaign: updatedCampaign });
      });

      root.innerHTML = "";
      root.appendChild(view);
    }

    draw();
  }

  window.EDCampaignView = { renderCampaignView };
})();
