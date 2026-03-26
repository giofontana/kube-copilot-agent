/* ============================================
   KubeCopilot Website — JS
   ============================================ */

document.addEventListener("DOMContentLoaded", function () {
  // --- Mobile nav toggle ---
  var toggle = document.getElementById("nav-toggle");
  var links = document.getElementById("nav-links");

  if (toggle && links) {
    toggle.addEventListener("click", function () {
      links.classList.toggle("open");
    });

    // Close menu on link click
    links.querySelectorAll("a").forEach(function (link) {
      link.addEventListener("click", function () {
        links.classList.remove("open");
      });
    });
  }

  // --- Navbar background on scroll ---
  var navbar = document.getElementById("navbar");
  if (navbar) {
    window.addEventListener("scroll", function () {
      if (window.scrollY > 60) {
        navbar.style.background = "rgba(10, 14, 26, 0.95)";
      } else {
        navbar.style.background = "rgba(10, 14, 26, 0.85)";
      }
    });
  }

  // --- Screenshot tabs ---
  var tabButtons = document.querySelectorAll(".tab-btn");
  var screenshotImg = document.getElementById("screenshot-img");

  var screenshotMap = {
    "main-ui": { src: "images/main-ui.png", alt: "KubeCopilot Chat Interface" },
    "settings-model": { src: "images/settings-model.png", alt: "Model Selection Settings" },
    "settings-skills": { src: "images/settings-skills.png", alt: "Skills Management" },
    "settings-agents": { src: "images/settings-agents.png", alt: "Custom Agents Configuration" },
    "settings-byok": { src: "images/settings-byok.png", alt: "BYOK Provider Configuration" }
  };

  tabButtons.forEach(function (btn) {
    btn.addEventListener("click", function () {
      tabButtons.forEach(function (b) { b.classList.remove("active"); });
      btn.classList.add("active");

      var tab = btn.getAttribute("data-tab");
      var info = screenshotMap[tab];
      if (info && screenshotImg) {
        screenshotImg.src = info.src;
        screenshotImg.alt = info.alt;
      }
    });
  });

  // --- Smooth scroll for anchor links ---
  document.querySelectorAll('a[href^="#"]').forEach(function (anchor) {
    anchor.addEventListener("click", function (e) {
      var target = document.querySelector(this.getAttribute("href"));
      if (target) {
        e.preventDefault();
        target.scrollIntoView({ behavior: "smooth", block: "start" });
      }
    });
  });
});
