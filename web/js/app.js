(function () {
  const root = document.getElementById("view-root");

  function showGroupList() {
    EDGroupList.renderGroupList(root);
  }

  function showNewGroupForm() {
    EDGroupList.renderNewGroupForm(root);
  }

  function showGame(groupId) {
    const group = EDStorage.getGroup(groupId);
    if (!group) return showGroupList();
    EDGameTracker.renderGameTracker(root, {
      title: group.name,
      game: group.game,
      onChange: (updatedGame) => {
        group.game = updatedGame;
        group.updatedAt = new Date().toISOString();
        EDStorage.saveGroup(group);
      },
      onBack: showGroupList,
    });
  }

  function showCampaign(groupId) {
    const group = EDStorage.getGroup(groupId);
    if (!group) return showGroupList();
    EDCampaignView.renderCampaignView(root, {
      group,
      onSave: (updatedGroup) => EDStorage.saveGroup(updatedGroup),
      onBack: showGroupList,
      onOpenChapterGame: (chapterIndex, gameIndex) => showChapterGame(groupId, chapterIndex, gameIndex),
    });
  }

  function showChapterGame(groupId, chapterIndex, gameIndex) {
    const group = EDStorage.getGroup(groupId);
    if (!group) return showGroupList();
    const chapter = group.campaign.chapters[chapterIndex];
    const game = chapter.games[gameIndex];
    EDGameTracker.renderGameTracker(root, {
      title: `${group.name} — ${chapter.name}`,
      game,
      onChange: (updatedGame) => {
        group.campaign.chapters[chapterIndex].games[gameIndex] = updatedGame;
        group.updatedAt = new Date().toISOString();
        EDStorage.saveGroup(group);
      },
      onBack: () => showCampaign(groupId),
    });
  }

  window.EDApp = { showGroupList, showNewGroupForm, showGame, showCampaign, showChapterGame };

  EDWasm.ready()
    .then(() => {
      document.getElementById("app-loading").remove();
      showGroupList();
    })
    .catch((err) => {
      const loading = document.getElementById("app-loading");
      loading.innerHTML = `<p>Failed to load: ${err.message}</p>`;
    });
})();
