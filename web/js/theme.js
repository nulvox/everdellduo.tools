(function () {
  const KEY = "everdellduo_theme";

  function apply(theme) {
    document.documentElement.setAttribute("data-theme", theme);
  }

  function current() {
    return localStorage.getItem(KEY) || "day";
  }

  function toggle() {
    const next = current() === "day" ? "night" : "day";
    localStorage.setItem(KEY, next);
    apply(next);
  }

  apply(current());

  document.addEventListener("DOMContentLoaded", () => {
    const btn = document.getElementById("btn-theme-toggle");
    if (btn) btn.addEventListener("click", toggle);
  });
})();
