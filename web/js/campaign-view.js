(function () {
  function summarizeGame(game) {
    const totals = (game.players || []).map((p) => `${p.name}: ${p.score || 0}`).join(", ");
    return totals || "(no players)";
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
        const gameTypeSelect = card.querySelector(".chapter-game-type");
        card.querySelector(".btn-start-game").addEventListener("click", () => {
          const gameType = gameTypeSelect.value;
          const updatedCampaign = EDWasm.startChapterGame(group.campaign, chapterIndex, gameType, group.players);
          persist({ ...group, campaign: updatedCampaign });
        });

        const gameList = card.querySelector(".chapter-game-list");
        (chapter.games || []).forEach((game, gameIndex) => {
          const row = document.getElementById("tpl-chapter-game-row").content.cloneNode(true);
          row.querySelector(".game-status").textContent = game.status === "completed" ? "✅" : "▶️";
          row.querySelector(".game-summary").textContent = `${game.gameType}: ${summarizeGame(game)}`;
          row.querySelector(".btn-resume").addEventListener("click", () =>
            opts.onOpenChapterGame(chapterIndex, gameIndex)
          );
          gameList.appendChild(row);
        });

        chapterList.appendChild(card);
      });

      view.querySelector(".form-add-chapter").addEventListener("submit", (ev) => {
        ev.preventDefault();
        const input = ev.target.elements.chapterName;
        const name = input.value.trim();
        if (!name) return;
        const updatedCampaign = EDWasm.addChapter(group.campaign, name);
        persist({ ...group, campaign: updatedCampaign });
      });

      root.innerHTML = "";
      root.appendChild(view);
    }

    draw();
  }

  window.EDCampaignView = { renderCampaignView };
})();
