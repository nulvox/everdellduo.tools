(function () {
  function download(filename, text) {
    const blob = new Blob([text], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  }

  function exportAll() {
    const groups = EDStorage.allGroupsFull();
    const json = EDWasm.exportBackup(groups);
    const stamp = new Date().toISOString().replace(/[:.]/g, "-");
    download(`everdellduo-backup-${stamp}.json`, json);
  }

  function exportGroup(group) {
    const json = EDWasm.exportBackup([group]);
    const safeName = group.name.replace(/[^a-z0-9]+/gi, "-").toLowerCase();
    download(`everdellduo-${safeName}.json`, json);
  }

  function importFile(file, onDone) {
    const reader = new FileReader();
    reader.onload = () => {
      try {
        const backup = EDWasm.importBackup(reader.result);
        EDStorage.replaceAll(backup.groups || []);
        onDone(null);
      } catch (e) {
        onDone(e);
      }
    };
    reader.onerror = () => onDone(reader.error);
    reader.readAsText(file);
  }

  window.EDImportExport = { exportAll, exportGroup, importFile };

  document.addEventListener("DOMContentLoaded", () => {
    const exportBtn = document.getElementById("btn-export-all");
    if (exportBtn) exportBtn.addEventListener("click", exportAll);

    const importInput = document.getElementById("input-import-all");
    if (importInput) {
      importInput.addEventListener("change", (ev) => {
        const file = ev.target.files[0];
        if (!file) return;
        const confirmed = confirm(
          "Importing a backup replaces ALL groups currently stored in this browser. Continue?"
        );
        if (!confirmed) {
          ev.target.value = "";
          return;
        }
        importFile(file, (err) => {
          ev.target.value = "";
          if (err) {
            alert("Import failed: " + err.message);
            return;
          }
          window.EDApp.showGroupList();
        });
      });
    }
  });
})();
