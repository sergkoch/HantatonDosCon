(function () {
  try {
    var stored = localStorage.getItem("theme");
    document.documentElement.dataset.theme =
      stored === "light" || stored === "dark" ? stored : "dark";
  } catch (e) {
    document.documentElement.dataset.theme = "dark";
  }
})();
