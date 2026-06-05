/* Minimal inline-SVG icon set (Lucide-style strokes) used by the chrome, so the
   renderer has zero asset dependencies and works from file:// in Electron. */
(function (root) {
  const SVG = (paths, extra) =>
    `<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" ` +
    `stroke-width="2" stroke-linecap="round" stroke-linejoin="round" ${extra || ''}>${paths}</svg>`;

  const Icons = {
    menu: SVG('<line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>'),
    git: SVG('<circle cx="6" cy="6" r="3"/><circle cx="18" cy="18" r="3"/><path d="M6 9v6a3 3 0 0 0 3 3h6"/>'),
    branch: SVG('<line x1="6" y1="3" x2="6" y2="15"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/>'),
    check: SVG('<polyline points="20 6 9 17 4 12"/>'),
    folder: SVG('<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>'),
    file: SVG('<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>'),
    minimize: SVG('<line x1="5" y1="12" x2="19" y2="12"/>'),
    maximize: SVG('<rect x="5" y="5" width="14" height="14" rx="1"/>'),
    close: SVG('<line x1="6" y1="6" x2="18" y2="18"/><line x1="18" y1="6" x2="6" y2="18"/>'),
  };

  root.OctraIcons = Icons;
})(window);
