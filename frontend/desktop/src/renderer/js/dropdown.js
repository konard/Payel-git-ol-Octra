/* Tiny shared dropdown controller: tracks the single open dropdown and a backdrop
   so any open menu closes on outside click or Escape. */
(function (root) {
  let openEl = null;
  let openTrigger = null;
  const backdrop = () => document.querySelector('.backdrop');

  function close() {
    if (openEl) openEl.setAttribute('data-open', 'false');
    if (openTrigger) openTrigger.setAttribute('aria-expanded', 'false');
    const b = backdrop();
    if (b) b.setAttribute('data-open', 'false');
    openEl = null;
    openTrigger = null;
  }

  // Open `el` anchored under `trigger`. Re-clicking the same trigger toggles it.
  function open(el, trigger) {
    if (openEl === el) {
      close();
      return;
    }
    close();
    openEl = el;
    openTrigger = trigger;
    el.setAttribute('data-open', 'true');
    if (trigger) {
      trigger.setAttribute('aria-expanded', 'true');
      // Anchor the dropdown's left edge to the trigger, clamped to the viewport.
      const rect = trigger.getBoundingClientRect();
      const left = Math.min(rect.left, window.innerWidth - el.offsetWidth - 8);
      el.style.left = Math.max(4, left) + 'px';
    }
    const b = backdrop();
    if (b) b.setAttribute('data-open', 'true');
  }

  function isOpen(el) {
    return openEl === el;
  }

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') close();
  });

  root.OctraDropdown = { open, close, isOpen, current: () => openEl };
})(window);
