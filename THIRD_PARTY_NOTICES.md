# Third-Party Notices

This file documents third-party libraries downloaded at Docker build time and bundled into the container image.

## JavaScript Libraries

### HTMX
- **Version**: 1.9.12
- **License**: BSD 2-Clause
- **URL**: https://htmx.org/
- **Files**: `static/js/htmx.min.js`, `static/js/htmx-json-enc.js`

### Alpine.js
- **Version**: 3.14.3
- **License**: MIT
- **URL**: https://alpinejs.dev/
- **Files**: `static/js/alpine.min.js`

### Chart.js
- **Version**: 4.4.7
- **License**: MIT
- **URL**: https://www.chartjs.org/
- **Files**: `static/js/chart.min.js`

### TipTap
- **Version**: 3.5.1
- **License**: MIT
- **URL**: https://tiptap.dev/
- **Files**: `static/js/tiptap.min.js` (bundled with extensions)
- **Included Extensions**:
  - @tiptap/core
  - @tiptap/starter-kit
  - @tiptap/extension-placeholder
  - @tiptap/extension-table (+ row, cell, header)
  - @tiptap/extension-text-align
  - @tiptap/extension-text-style
  - @tiptap/extension-color
  - @tiptap/extension-highlight
  - @tiptap/extension-underline
  - @tiptap/extension-task-list
  - @tiptap/extension-task-item
  - @tiptap/extension-image

## CSS Libraries

### Font Awesome Free
- **Version**: 6.7.2
- **License**: 
  - Icons: CC BY 4.0 (https://creativecommons.org/licenses/by/4.0/)
  - Fonts: SIL OFL 1.1 (https://scripts.sil.org/OFL)
  - Code: MIT (https://opensource.org/licenses/MIT)
- **URL**: https://fontawesome.com/
- **Files**: 
  - `static/css/fontawesome/all.min.css`
  - `static/webfonts/fa-brands-400.woff2`
  - `static/webfonts/fa-regular-400.woff2`
  - `static/webfonts/fa-solid-900.woff2`

### Tailwind CSS
- **Version**: 3.4.x
- **License**: MIT
- **URL**: https://tailwindcss.com/
- **Files**: `static/css/output.css` (compiled at build time)

## Build-Time Tools (not included in runtime)

### esbuild
- **Version**: 0.25.x
- **License**: MIT
- **URL**: https://esbuild.github.io/
- **Usage**: Bundles TipTap extensions

## Updating Versions

Third-party asset versions are pinned in `Dockerfile`. To update:

1. Edit the version numbers in the `assets` stage of `Dockerfile`
2. Rebuild the containers: `make restart`
3. Test the application
4. Update this file with new version numbers
