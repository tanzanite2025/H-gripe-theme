# Admin Dark Theme Guide

The admin app includes a dark theme stylesheet at:

```text
src/styles/theme-dark.css
```

## Enable Globally

Import the stylesheet in the admin app entrypoint and set the theme attribute:

```js
import './styles/theme-dark.css'

document.body.setAttribute('data-theme', 'dark')
```

## Enable With State

Prefer a Pinia store or app-level setting if the theme should be user-selectable.

```js
document.body.setAttribute('data-theme', selectedTheme)
localStorage.setItem('admin_theme', selectedTheme)
```

## Maintenance Rules

- Keep theme variables centralized in `theme-dark.css`.
- Avoid hard-coded colors in page components when a theme variable exists.
- Test Element Plus forms, tables, dialogs, and menus after changing theme tokens.
- Keep accessibility contrast readable in both light and dark modes.
