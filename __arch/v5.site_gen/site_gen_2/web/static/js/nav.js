(function () {
  var header = document.querySelector("[data-nav]");
  var btn = document.querySelector("[data-nav-toggle]");
  if (!header || !btn) return;

  function set(open) {
    header.setAttribute("data-nav", open ? "open" : "closed");
    btn.setAttribute("aria-expanded", open ? "true" : "false");
    btn.setAttribute("aria-label", open ? "Close menu" : "Open menu");
  }

  btn.addEventListener("click", function () {
    set(header.getAttribute("data-nav") !== "open");
  });

  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape" && header.getAttribute("data-nav") === "open") {
      set(false);
      btn.focus();
    }
  });

  document.addEventListener("click", function (e) {
    if (header.getAttribute("data-nav") === "open" && !header.contains(e.target)) {
      set(false);
    }
  });

  header.addEventListener("click", function (e) {
    if (header.getAttribute("data-nav") === "open" && e.target.closest("#main-nav a")) {
      set(false);
    }
  });

  var mq = window.matchMedia("(min-width: 768px)");
  var onChange = function (e) {
    if (e.matches) set(false);
  };
  if (mq.addEventListener) mq.addEventListener("change", onChange);
  else if (mq.addListener) mq.addListener(onChange);
})();
