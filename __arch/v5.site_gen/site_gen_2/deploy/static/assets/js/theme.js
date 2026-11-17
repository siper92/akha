(function () {
  var root = document.documentElement;
  var btn = document.querySelector("[data-theme-toggle]");
  if (!btn) return;

  function current() {
    return root.getAttribute("data-theme") === "light" ? "light" : "dark";
  }

  function sync() {
    var next = current() === "light" ? "dark" : "light";
    btn.setAttribute("aria-label", "Switch to " + next + " theme");
    btn.setAttribute("title", "Switch to " + next + " theme");
  }

  btn.addEventListener("click", function () {
    var next = current() === "light" ? "dark" : "light";
    root.setAttribute("data-theme", next);
    try {
      localStorage.setItem("theme", next);
    } catch (e) {}
    sync();
  });

  btn.hidden = false;
  sync();
})();
