// Durable localStorage persistence for groups. Each group is stored under
// its own key so large campaigns don't force a full manifest rewrite on
// every edit; a lightweight manifest key tracks what exists for listing.
(function () {
  const MANIFEST_KEY = "everdellduo_manifest";
  const GROUP_KEY_PREFIX = "everdellduo_group_";

  function readManifest() {
    try {
      const raw = localStorage.getItem(MANIFEST_KEY);
      return raw ? JSON.parse(raw) : [];
    } catch (e) {
      console.error("Failed to read manifest", e);
      return [];
    }
  }

  function writeManifest(manifest) {
    localStorage.setItem(MANIFEST_KEY, JSON.stringify(manifest));
  }

  function groupKey(id) {
    return GROUP_KEY_PREFIX + id;
  }

  function listGroups() {
    return readManifest();
  }

  function getGroup(id) {
    const raw = localStorage.getItem(groupKey(id));
    return raw ? JSON.parse(raw) : null;
  }

  function saveGroup(group) {
    localStorage.setItem(groupKey(group.id), JSON.stringify(group));
    const manifest = readManifest();
    const idx = manifest.findIndex((g) => g.id === group.id);
    const entry = {
      id: group.id,
      name: group.name,
      gameType: group.gameType,
      updatedAt: group.updatedAt,
    };
    if (idx >= 0) {
      manifest[idx] = entry;
    } else {
      manifest.push(entry);
    }
    writeManifest(manifest);
  }

  function deleteGroup(id) {
    localStorage.removeItem(groupKey(id));
    writeManifest(readManifest().filter((g) => g.id !== id));
  }

  function allGroupsFull() {
    return readManifest()
      .map((entry) => getGroup(entry.id))
      .filter(Boolean);
  }

  function replaceAll(groups) {
    // Used by import: wipes existing groups and manifest, then writes the
    // imported set fresh.
    readManifest().forEach((entry) => localStorage.removeItem(groupKey(entry.id)));
    writeManifest([]);
    groups.forEach((g) => saveGroup(g));
  }

  window.EDStorage = {
    listGroups,
    getGroup,
    saveGroup,
    deleteGroup,
    allGroupsFull,
    replaceAll,
  };
})();
