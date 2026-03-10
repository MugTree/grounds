document.addEventListener("DOMContentLoaded", function () {
  const deposit = document.querySelector(".deposit-trigger");
  if (deposit) {
    deposit.addEventListener("click", () => toggle());
  }

  function toggle() {
    var cont = document.getElementById("deposit");
    cont.style.display = cont.style.display == "none" ? "block" : "none";
  }

  let tabs = document.querySelectorAll(".tabs a");
  let relatedBoxes = document.querySelectorAll(".quickbox");
  let galleries = document.getElementById("galleries");

  for (const [index, element] of tabs.entries()) {
    element.addEventListener("click", (e) => {
      tabs.forEach((t) => t.classList.remove("tab-active"));
      tabs[index].classList.add("tab-active");
      relatedBoxes.forEach((t) => t.classList.add("hidden"));
      relatedBoxes[index].classList.remove("hidden");
    });
  }

  // Links to gallery
  document.querySelectorAll(".bolt-link").forEach((e) => {
    e.addEventListener("click", (e) => {
      var name = e.target.dataset.bolt;
      tabs.forEach((t) => {
        if (!t.classList.contains("tab-" + name)) {
          t.classList.remove("tab-active");
        } else {
          t.classList.add("tab-active");
        }
      });
      relatedBoxes.forEach((r) => {
        if (!r.classList.contains("quickbox-" + name)) {
          r.classList.add("hidden");
        } else {
          r.classList.remove("hidden");
        }
      });

      galleries.scrollIntoView({
        behavior: "smooth",
      });
    });
  });
});
