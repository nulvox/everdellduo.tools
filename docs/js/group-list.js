(function () {
  const GAME_TYPE_LABELS = {
    challenge: "🏆 Challenge",
    duel: "⚔️ Duel",
    coop: "🤝 Co-op",
    campaign: "📖 Campaign",
  };

  function renderGroupList(root) {
    const view = document.getElementById("tpl-group-list").content.cloneNode(true);
    const list = view.querySelector("#group-list");
    const empty = view.querySelector("#group-list-empty");
    const groups = EDStorage.listGroups().sort((a, b) => (a.updatedAt < b.updatedAt ? 1 : -1));

    if (groups.length === 0) {
      empty.hidden = false;
    } else {
      groups.forEach((entry) => {
        const card = document.getElementById("tpl-group-card").content.cloneNode(true);
        card.querySelector(".game-type-badge").textContent = GAME_TYPE_LABELS[entry.gameType] || entry.gameType;
        card.querySelector(".group-name").textContent = entry.name;
        card.querySelector(".group-updated").textContent = new Date(entry.updatedAt).toLocaleString();
        card.querySelector(".btn-open").addEventListener("click", () => {
          const group = EDStorage.getGroup(entry.id);
          if (group.gameType === "campaign") {
            window.EDApp.showCampaign(group.id);
          } else {
            window.EDApp.showGame(group.id);
          }
        });
        card.querySelector(".btn-export").addEventListener("click", () => {
          const group = EDStorage.getGroup(entry.id);
          EDImportExport.exportGroup(group);
        });
        card.querySelector(".btn-delete").addEventListener("click", () => {
          if (confirm(`Delete "${entry.name}"? This cannot be undone (export a backup first if unsure).`)) {
            EDStorage.deleteGroup(entry.id);
            window.EDApp.showGroupList();
          }
        });
        list.appendChild(card);
      });
    }

    view.querySelector("#btn-new-group").addEventListener("click", () => {
      window.EDApp.showNewGroupForm();
    });

    root.innerHTML = "";
    root.appendChild(view);
  }

  function renderNewGroupForm(root) {
    const view = document.getElementById("tpl-new-group").content.cloneNode(true);
    view.querySelector("#btn-cancel-new-group").addEventListener("click", () => window.EDApp.showGroupList());

    const difficultyFieldset = view.querySelector(".difficulty-picker");
    view.querySelectorAll('input[name="gameType"]').forEach((radio) => {
      radio.addEventListener("change", () => {
        difficultyFieldset.hidden = radio.value !== "campaign";
      });
    });

    view.querySelector("#form-new-group").addEventListener("submit", (ev) => {
      ev.preventDefault();
      const form = new FormData(ev.target);
      const name = form.get("name").trim();
      const gameType = form.get("gameType");
      const difficulty = form.get("difficulty") || "normal";
      const players = form
        .get("players")
        .split(",")
        .map((p) => p.trim())
        .filter(Boolean);
      if (!name || !gameType || players.length === 0) return;

      const group = EDWasm.newGroup(name, gameType, players, difficulty);
      EDStorage.saveGroup(group);
      if (gameType === "campaign") {
        window.EDApp.showCampaign(group.id);
      } else {
        window.EDApp.showGame(group.id);
      }
    });

    root.innerHTML = "";
    root.appendChild(view);
  }

  window.EDGroupList = { renderGroupList, renderNewGroupForm, GAME_TYPE_LABELS };
})();
