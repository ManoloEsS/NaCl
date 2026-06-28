# CSS Styling Cheat Sheet — NaCl

**Target audience:** Backend developers who know basic CSS (classes, colors, fonts) but are not frontend specialists.
**Goal:** Understand every CSS pattern used in the NaCl frontend so you can add new pages, tweak layouts, and fix styling without guessing.

---

## 1. How CSS works in this project

NaCl uses **plain CSS in a single file**, located at `nacl_frontend/src/index.css`. There is no Tailwind, no CSS modules, no styled-components — just one stylesheet imported at the app entry point (`main.tsx`).

**Key difference from backend templating:** In React, you use `className` instead of `class`:

```tsx
// HTML: <div class="card">
// React: <div className="card">
```

Every class name in `index.css` is available globally. If you add `.card { ... }` to the file, any component can use `className="card"`.

---

## 2. Selectors — how to target elements

A **selector** tells CSS which elements to style. These are the selectors used in this project:

| Selector | What it targets | Example from NaCl |
|---|---|---|
| `.class-name` | Any element with that class | `.card` targets `<div className="card">` |
| `element` | All elements of that HTML tag | `body` targets the `<body>` tag |
| `.parent .child` | `.child` anywhere inside `.parent` (any nesting depth) | `.login-card h1` targets the `<h1>` inside the login card |
| `.parent > .child` | `.child` that is a **direct** child of `.parent` | `.credential-card > div` targets only the immediate `<div>` children, not nested ones deeper inside |
| `element:hover` | When the mouse is over the element | `.btn-primary:hover` applies when hovering a button |
| `element:focus` | When the element is focused (tab key or click into an input) | `.form-group input:focus` styles the blue outline ring |
| `element:disabled` | When a button/input is disabled | `.btn-primary:disabled` dims the button |
| `element:last-child` | The last child among its siblings | `.credential-card > div:last-child` removes bottom margin on the last line |
| `element + element` | An element that comes **directly after** another (adjacent sibling) | `.nav-links a + a::before` adds a separator before every link except the first |
| `element::before` | A decorative element inserted **before** the content | `.nav-links a + a::before` creates the vertical line separator |
| `element::after` | A decorative element inserted **after** the content | `.nav-links a.active::after` creates the active link underline |
| `element::-webkit-scrollbar` | The scrollbar (WebKit browsers only) | `.scrollable-list::-webkit-scrollbar` styles the thin dark scrollbar |

**Selector specificity quick rule:** A class (`.card`) beats an element (`div`). Two classes (`.credential-card > .credential-service`) beat one class + one element (`.credential-card > div`). When two rules conflict, the one with higher specificity wins. When specificity is equal, the rule that comes **later** in the file wins.

---

## 3. The Box Model — every element is a rectangle

Every element on the page is a rectangle with three layers of space around its content:

```
┌───────────────────────────────────┐
│          MARGIN (outside)         │
│   ┌───────────────────────────┐   │
│   │        BORDER             │   │
│   │   ┌───────────────────┐   │   │
│   │   │     PADDING       │   │   │
│   │   │   ┌───────────┐   │   │   │
│   │   │   │  CONTENT  │   │   │   │
│   │   │   └───────────┘   │   │   │
│   │   └───────────────────┘   │   │
│   └───────────────────────────┘   │
└───────────────────────────────────┘
```

**Properties used in this project:**

| Property | What it does | Example |
|---|---|---|
| `margin` | Space **outside** the element (pushes other elements away) | `.card` has no margin; `.form-group` has `margin-bottom: 1rem` to separate fields |
| `border` | Visible edge around the element | `.card` has `border: 1px solid #3b4261` |
| `padding` | Space **inside** the element (pushes content inward) | `.card` has `padding: 1.5rem` so text doesn't touch the edges |
| `width` | How wide the element is | `.btn-full` has `width: 100%` |
| `max-width` | Maximum width (element can be narrower but not wider) | `.main-content` has `max-width: 960px` so it shrinks on small screens |
| `min-width` | Minimum width (element can be wider but not narrower) | `.credential-card strong` has `min-width: 10rem` so all labels align |

**`margin: auto` trick:** Setting `margin-left: auto` and `margin-right: auto` on a block element centers it horizontally. NaCl uses this pattern: `.main-content { margin: 1.5rem auto; }` centers the content area.

---

## 4. Layout with Flexbox

Flexbox is the main tool for arranging items in rows or columns. You turn a container into a flexbox with `display: flex`, then control how its children behave.

### 4a. Basic flex row

```css
.container {
  display: flex;
  align-items: center;    /* vertical alignment */
  gap: 1.5rem;            /* space between children */
}
```

```
┌──────────────────────────────────────────────┐
│  [item 1]  ───1.5rem───  [item 2]  ───1.5rem───  [item 3]  │
└──────────────────────────────────────────────┘
                    all vertically centered
```

### 4b. Common flex properties

| Property | Values | What it does |
|---|---|---|
| `display: flex` | — | Turns element into a flex container |
| `flex-direction` | `row` (default), `column` | Direction items flow |
| `align-items` | `center`, `flex-start`, `stretch` | Vertical alignment (or horizontal if `flex-direction: column`) |
| `justify-content` | `center`, `space-between`, `flex-end`, `flex-start` | Horizontal distribution |
| `gap` | any size (e.g. `1.5rem`) | Space between children |
| `flex: 1` | — | Child grows to fill remaining space |
| `flex-shrink: 0` | — | Child refuses to shrink |

### 4c. Practical examples from NaCl

**Navbar (items spread across the row):**
```css
.navbar {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}
.nav-links { flex: 1; }  /* pushes .nav-user to the right */
```

```
┌────────────────────────────────────────────────────────┐
│  [Brand]  [Vault] | [New] | [Account]    [user] [Logout]  │
│                         ↑ flex: 1 pushes this gap       │
└────────────────────────────────────────────────────────┘
```

**Two-column layout (form + info panel):**
```css
.form-with-info {
  display: flex;
  gap: 1.5rem;
  align-items: flex-start;  /* top-align both columns */
}
.form-with-info .card { flex: 1; }  /* card takes remaining space */
.info-panel { width: 280px; flex-shrink: 0; }  /* fixed width, won't shrink */
```

```
┌──────────────────────────────┬──────────────┐
│                              │              │
│     Form card (flex: 1)      │  Info panel  │
│     takes remaining space    │  280px wide  │
│                              │              │
└──────────────────────────────┴──────────────┘
         ←──── gap: 1.5rem ────→
```

**Row with label on left, button on right:**
```css
.credential-copy-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}
```

```
┌──────────────────────────────────┐
│  Username: myuser    [Copy]      │
│  ↑ left               ↑ right    │
└──────────────────────────────────┘
       space-between
```

**Button row:**
```css
.credential-btn-line {
  display: flex;
  gap: 2.5rem;
}
```

---

## 5. Positioning — controlling where elements sit

By default, elements flow top-to-bottom in the order they appear in HTML (`position: static`). The other position values let you break out of that flow:

| Value | How it behaves | Used for |
|---|---|---|
| `static` (default) | Normal document flow | Most elements |
| `relative` | Same as static, but you can offset it with `top`/`left`/`right`/`bottom` | Container for `absolute` children |
| `absolute` | Removed from normal flow. Positioned relative to the nearest positioned ancestor (or the page) | Pseudo-elements (`::before`, `::after`) inside a `position: relative` parent |
| `fixed` | Removed from normal flow. Stays in the same place on screen even when scrolling | Toast notifications stay at bottom-right always |
| `sticky` | Normal flow until you scroll past it, then it "sticks" in place | Navbar sticks to top when scrolling; info panel sticks below the navbar |

**Sticky navbar in NaCl:**
```css
.navbar {
  position: sticky;
  top: 0;         /* sticks when top of navbar reaches top of viewport */
  z-index: 100;   /* stays above other content */
}
```

**Fixed toast container:**
```css
.toast-container {
  position: fixed;
  bottom: 1rem;   /* 16px from bottom of screen */
  right: 1rem;    /* 16px from right of screen */
  z-index: 1000;  /* above everything */
}
```

**Absolute positioning for pseudo-element separators:**
```css
.nav-links a { position: relative; }  /* anchor point for the ::before */

.nav-links a + a::before {
  content: '';
  position: absolute;
  left: 0;               /* at the left edge of the link */
  top: 50%;              /* halfway down */
  transform: translateY(-50%);  /* adjust to truly center vertically */
  width: 1px;
  height: 1.2em;
  background: #3b4261;
}
```

**`z-index` rules of thumb:**
- Higher number = closer to the viewer (in front)
- Navbar: `100`, info panel: sticky at `5rem` (no z-index needed), toasts: `1000`
- Only works on positioned elements (`relative`, `absolute`, `fixed`, `sticky`)

---

## 6. The color palette — Tokyo Night Storm

NaCl uses a dark blue theme inspired by the Tokyo Night Storm editor theme. These 8 colors are reused everywhere:

| Name | Hex | Looks like | Used for |
|---|---|---|---|
| Background | `#24283b` | Dark navy | Page background, input background |
| Card/Nav bg | `#1f2335` | Slightly darker navy | Cards, navbar, panels |
| Input bg | `#1a1b26` | Deepest navy | Form input backgrounds |
| Text | `#c0caf5` | Light blue-white | Headings, body text, labels that need emphasis |
| Muted | `#565f89` | Muted gray-blue | Secondary text, labels, placeholders, subtext |
| Blue accent | `#7aa2f7` | Bright blue | Links, active nav links, focus rings, info toasts |
| Orange | `#e0af68` | Warm orange | Primary buttons (Encrypt and Save, Update) |
| Red | `#f7768e` | Pink-red | Danger buttons (Delete), error messages |
| Green | `#9ece6a` | Soft green | Success toasts |
| Border | `#3b4261` | Dark gray-blue | Card borders, input borders, separators |

**How to read `#7aa2f7`:**
- `#` = this is a hex color
- `7a` = amount of **red** (0 = none, ff = full)
- `a2` = amount of **green**
- `f7` = amount of **blue**
- So `#7aa2f7` has a lot of blue, some green, and a little red — making it a bright blue

**Adding a new color:** If you need a color not in the palette, add a CSS custom property at the top of `index.css`:

```css
:root {
  --my-new-color: #somehex;
}
```

Then use it as `color: var(--my-new-color)`. This keeps colors consistent across the app.

---

## 7. Typography — text styling

```css
body {
  font-family: 'Satoshi', -apple-system, BlinkMacSystemFont, sans-serif;
  font-size: 16px;  /* default, set by browser */
}
```

**Key properties:**

| Property | What it does | Example |
|---|---|---|
| `font-family` | Which font to use (first available wins) | `'Satoshi', sans-serif` — tries Satoshi first, then any sans-serif fallback |
| `font-size` | How big the text is | `2rem`, `1.1rem`, `0.85rem` |
| `font-weight` | How bold | `400` = normal, `500` = medium, `600` = semi-bold, `700` = bold |
| `line-height` | Space between lines of text | `1.6` = 1.6x the font size (good for paragraphs) |
| `text-align` | Left, center, or right | `center` for titles |
| `text-transform` | Force upper/lower case | `uppercase` for table headers |
| `text-decoration` | Underline or none | `none` on links by default, `underline` on hover |

**Font sizes used in NaCl (from largest to smallest):**

| Size | Where used |
|---|---|
| `1.7rem` (~27px) | Service name in credential card |
| `1.75rem` (~28px) | Login page title "NaCl" |
| `1.5rem` (~24px) | Navbar brand "NaCl" |
| `1.1rem` (~18px) | Card titles, section headings |
| `1rem` (~16px) | Body text, credential card lines |
| `0.95rem` (~15px) | Form inputs, buttons |
| `0.9rem` (~14px) | Descriptions, secondary text |
| `0.85rem` (~14px) | Form labels, small buttons |
| `0.75rem` (~12px) | Copy buttons, info icons |
| `0.7rem` (~11px) | `?` info circle icon |

---

## 8. Transitions — smooth hover effects

A **transition** makes a property change gradually instead of instantly. Without it, hovering a button would snap instantly to the new color. With it, the color fades smoothly over 0.2 seconds.

```css
.btn-primary {
  background: #e0af68;
  transition: background 0.2s ease;  /* animate background changes */
}
.btn-primary:hover { background: #e6c07a; }  /* new value */
```

**The rules:**
1. Put `transition` on the **base state** (`.btn-primary`), not on `:hover`
2. The first value is the property to animate (`background`, `color`, `border-color`, `box-shadow`)
3. The second value is duration (`0.2s` = 200 milliseconds)
4. The third value is timing function (`ease` = starts fast, ends slow)

**Common transitions in NaCl:**
```css
transition: color 0.2s ease;                    /* link hover color changes */
transition: background 0.2s ease;               /* button hover bg changes */
transition: border-color 0.2s ease, box-shadow 0.2s ease;  /* input focus changes */
```

---

## 9. Key reusable patterns

### 9a. Card pattern

Cards are dark rounded rectangles that group related content. They are the most common layout element.

```css
.card {
  background: #1f2335;      /* dark card background */
  padding: 1.5rem;          /* space inside the card */
  border-radius: 12px;      /* rounded corners */
  border: 1px solid #3b4261; /* subtle border */
}
.card-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: #c0caf5;
}
.card-text {
  font-size: 0.9rem;
  line-height: 1.6;
  color: #565f89;
}
.card-narrow { max-width: 660px; }  /* optional width limit */
```

**Usage in components:**
```tsx
<div className='card'>
  <h3 className='card-title'>Why encrypt your own passwords?</h3>
  <p className='card-text'>NaCl encrypts everything on your device...</p>
</div>

<div className='card card-narrow'>
  {/* narrower card, good for forms */}
  <form>...</form>
</div>
```

### 9b. Button pattern

Buttons use a base set of styles with modifier classes for different colors/sizes.

```css
.btn-primary { background: #e0af68; color: #1a1b26; padding: 0.65rem 1rem; font-weight: 700; cursor: pointer; }
.btn-primary:hover { background: #e6c07a; }
.btn-primary:disabled { opacity: 0.4; cursor: not-allowed; }

.btn-danger  { background: #f7768e; color: #1a1b26; padding: 0.25rem 0.75rem; font-weight: 700; }
.btn-small   { font-size: 0.85rem; padding: 0.3rem 0.6rem; }  /* smaller text + padding */
.btn-full    { width: 100%; }                                   /* full-width button */
```

**Usage:**
```tsx
<button className='btn-primary'>Save</button>
<button className='btn-primary btn-full'>Full Width Button</button>
<button className='btn-danger btn-small'>Delete</button>
<button className='btn-small'>Cancel</button>
```

### 9c. Form group pattern

Every form field follows the same structure: a wrapper with a label and an input.

```css
.form-group { margin-bottom: 1rem; }
.form-group label {
  display: block;               /* label on its own line above input */
  margin-bottom: 0.35rem;
  font-weight: 500;
  font-size: 0.85rem;
  color: #565f89;
}
.form-group input[type="text"],
.form-group input[type="password"],
.form-group select,
.form-group textarea {
  width: 100%;                  /* fill the container */
  padding: 0.65rem 0.75rem;
  border: 1px solid #3b4261;
  border-radius: 8px;
  font-size: 0.95rem;
  background: #1a1b26;
  color: #c0caf5;
}
.form-group input:focus {
  outline: 2px solid #7aa2f7;   /* blue ring on focus */
  outline-offset: 1px;           /* tiny gap between ring and input */
  border-color: transparent;     /* hide the border when focused */
}
```

**Usage in components (together with React Hook Form):**
```tsx
<div className='form-group'>
  <label htmlFor='email'>Email</label>
  <input id='email' type='email' {...register('email')} />
  {errors.email && <span className='field-error'>{errors.email.message}</span>}
</div>
```

**Variation with an info icon next to the label:**
```tsx
<div className='form-group'>
  <div className='form-label-row'>
    <label htmlFor='service'>Service</label>
    <span className='field-info'>?</span>
  </div>
  <input id='service' type='text' {...register('service')} />
</div>
```

```css
.form-label-row { display: flex; align-items: center; gap: 0.35rem; }
.field-info {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px; height: 18px;
  border-radius: 50%;
  border: 1px solid #565f89;
  color: #565f89;
  font-size: 0.7rem;
  cursor: default;
  transition: border-color 0.2s ease, color 0.2s ease;
}
.field-info:hover { border-color: #7aa2f7; color: #7aa2f7; }
```

### 9d. Two-column layout (form with info panel)

Used on the New Credential page and Account page.

```css
.form-with-info { display: flex; gap: 1.5rem; align-items: flex-start; }
.form-with-info .card { flex: 1; }                     /* left column fills space */
.info-panel { width: 280px; flex-shrink: 0; }          /* right column fixed width, sticky */
```

The `.info-panel` also uses `position: sticky; top: 5rem` so it follows you as you scroll through the form.

### 9e. Sticky navbar

```css
.navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 1.5rem;
  background: #1f2335;
  padding: 0.75rem 2rem;
  border-bottom: 1px solid #3b4261;
}
```

The navbar is a flex container with the brand on the left, links in the middle (pushed by `flex: 1`), and user info on the right. Active links get a blue underline via `::after`. Links have vertical separators via `::before` on `a + a`.

### 9f. Scrollable list

When you have a list that might be longer than the available space:

```css
.scrollable-list {
  max-height: 700px;      /* or whatever height you need */
  overflow-y: auto;       /* show scrollbar when content overflows */
}
```

Custom scrollbar styling (optional, only works in WebKit browsers like Chrome/Safari/Edge):
```css
.scrollable-list::-webkit-scrollbar       { width: 6px; }
.scrollable-list::-webkit-scrollbar-track { background: transparent; }
.scrollable-list::-webkit-scrollbar-thumb { background: #3b4261; border-radius: 3px; }
```

---

## 10. Units — px vs rem

| Unit | What it is | Example |
|---|---|---|
| `px` | One pixel on screen | `border: 1px solid ...`, `outline: 2px solid ...` |
| `rem` | Relative to the root font size (default 16px) | `1rem = 16px`, `1.5rem = 24px`, `0.85rem ≈ 14px` |

**Why use `rem` instead of `px`?**
- If a user has changed their browser's default font size (accessibility), `rem` scales with it. `px` does not.
- It's easier to maintain a consistent rhythm: if all sizes use `rem`, changing the root font size changes everything proportionally.

**When to use `px`:**
- Borders (`1px`, `2px`)
- Widths that must never change (custom scrollbar `6px`, active underline `2px`)
- `outline-offset` (`1px`)

---

## 11. Recipe section — "How do I..."

### How do I center something on a page?

```css
.parent {
  display: flex;
  justify-content: center;  /* horizontal */
  align-items: center;      /* vertical */
  min-height: 100vh;        /* full viewport height */
}
```

### How do I put two things side by side?

```css
.parent { display: flex; gap: 1.5rem; }
.left   { flex: 1; }         /* takes remaining space */
.right  { width: 280px; }    /* fixed width */
```

### How do I push a button to the right side of a row?

```css
.row { display: flex; justify-content: space-between; align-items: center; }
```

### How do I add space between items in a row?

Use `gap` on the flex container: `gap: 1rem`. Do not add individual margins to each child.

### How do I make an input look different when focused (tabbed into)?

```css
.form-group input:focus {
  outline: 2px solid #7aa2f7;
  outline-offset: 1px;
  border-color: transparent;
}
```

### How do I make something follow as I scroll (sticky)?

```css
.element {
  position: sticky;
  top: 5rem;  /* how far from the top of the viewport before it sticks */
}
```

### How do I add a vertical line between navigation links?

```css
.link + .link::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 1px;
  height: 1.2em;
  background: #3b4261;
}
```

The `a + a` selector targets every link that follows another link (skips the first one). The `::before` inserts a thin line. The `position: absolute` with `top: 50%; transform: translateY(-50%)` centers it vertically.

### How do I make a hover effect smooth?

```css
.element { transition: color 0.2s ease; }  /* on the base */
.element:hover { color: #newcolor; }        /* just the new value */
```

### How do I add a card to a page?

Add the HTML in your component and make sure the `.card` class exists in `index.css` (it already does).

```tsx
<div className='card'>
  <h3 className='card-title'>Your Title</h3>
  <p className='card-text'>Your content here.</p>
</div>
```

### How do I make a button that fills the full width of its container?

```tsx
<button className='btn-primary btn-full'>Submit</button>
```

`.btn-full` just has `width: 100%`.

---

## Quick reference: most-used properties

| Property | Typical value | What it does |
|---|---|---|
| `display` | `flex` | Enables flexbox layout |
| `align-items` | `center`, `flex-start` | Vertical alignment in flex |
| `justify-content` | `center`, `space-between` | Horizontal distribution in flex |
| `gap` | `1rem`, `1.5rem` | Space between flex children |
| `flex` | `1` | Grow to fill space |
| `padding` | `0.75rem 1rem` | Space inside element (top/bottom left/right) |
| `margin` | `1.5rem auto` | Space outside element |
| `border` | `1px solid #3b4261` | Edge around element |
| `border-radius` | `8px`, `12px` | Roundness of corners (0 = square) |
| `background` | `#1f2335` | Background color |
| `color` | `#c0caf5` | Text color |
| `font-size` | `0.95rem` | Text size |
| `font-weight` | `700` | Boldness |
| `cursor` | `pointer`, `default` | Mouse cursor style |
| `position` | `sticky`, `fixed`, `relative` | Positioning mode |
| `top` / `left` | `0`, `5rem` | Offset for positioned elements |
| `z-index` | `100`, `1000` | Stack order (higher = in front) |
| `transition` | `color 0.2s ease` | Smooth animation on property change |
| `outline` | `2px solid #7aa2f7` | Focus ring (does not affect layout) |
| `overflow-y` | `auto` | Show scrollbar when content overflows |
| `max-width` | `960px`, `660px` | Maximum width (shrinks on small screens) |
