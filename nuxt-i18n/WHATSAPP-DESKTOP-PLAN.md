# WhatsApp Chat Modal – Desktop Layout Change Request

> **Scope**: Desktop breakpoint (`md:` and above) only. Mobile layout must remain untouched.

## Requirements

1. **Relocate action buttons**
   - Move the existing **WhatsApp** quick-action button and the **Transfer** button from the desktop header (right column) into the left agent list, directly following the selected agent’s name/email row.
   - Buttons should only appear for the currently selected agent.

2. **Adjust column widths without changing overall modal width**
   - Increase the left agent list column to **400 px**.
   - Reduce the right conversation pane’s width accordingly so the combined width equals the current container width (do not change the overall modal width or its percentage-based sizing).

3. **Remove redundant agent info in desktop header**
   - Since the left column already shows avatar/name/email, the desktop header row should no longer display this information. Keep only the close button (and any future controls that stay on the right).

4. **FAQ button placement**
   - Because the chat/share/products/orders tabs are shared (not per-agent instances), the desktop **FAQ** trigger can be placed once in the left column (e.g., above the email buttons), preventing duplicate controls per agent.

5. **Desktop-only chat search input**
   - Add a search box above the desktop agent list. It should visually match the modal styling (rounded corners, dark theme) and allow future filtering by user name or message content. For now, implement only the UI shell; wiring up the search logic can happen later. Reserve space below the input for displaying results or states.

## Additional Notes

- Ensure that moving elements or resizing columns does **not** affect the mobile layout (`md:hidden` / `hidden md:flex` sections must stay intact).
- Cross-check hover/focus states after relocation so the visual language remains consistent with existing gradient buttons.
- After implementing, verify that scrollbar behavior and responsiveness are preserved at widths down to the desktop breakpoint threshold.
