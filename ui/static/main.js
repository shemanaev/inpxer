document.addEventListener("DOMContentLoaded", () => {
  // DROPDOWNS
  const $clickableDropdowns = document.querySelectorAll(
    ".dropdown:not(.is-hoverable)",
  );

  if ($clickableDropdowns.length > 0) {
    $clickableDropdowns.forEach(($dropdown) => {
      if (!$dropdown.querySelector("button")) {
        return;
      }

      $dropdown.querySelector("button").addEventListener("click", (event) => {
        event.stopPropagation();
        $dropdown.classList.toggle("is-active");
      });
    });

    document.addEventListener("click", () => {
      closeDropdowns();
    });

    document.addEventListener('keydown', (event) => {
      let e = event || window.event;
      if (e.key === 'Esc' || e.key === 'Escape') {
        closeDropdowns();
      }
    });
  }

  function closeDropdowns() {
    $clickableDropdowns.forEach(($el) => {
      $el.classList.remove("is-active");
    });
  }
});
