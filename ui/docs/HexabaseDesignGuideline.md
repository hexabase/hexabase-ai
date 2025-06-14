# Hexabase Design System

This guideline defines the "Hexabase-like design" for all Hexabase product operation screens. The goal is to create a unified, high-quality user experience across all products.

## 1. Design Philosophy

The Hexabase design system is built on three core principles that guide every design decision.

- **Sophistication:** Delivering a refined, professional, and uncluttered user experience. The design should feel intelligent and precise.
- **Machine:** Evoking a sense of efficiency, reliability, and power. The UI should be robust and perform like a well-oiled machine.
- **Ease of Use:** Ensuring that interfaces are intuitive, predictable, and accessible to all users, enabling them to accomplish their tasks with minimal friction.

---

## 2. Color Palette

The color palette is minimalist and clean, using a combination of neutral primary colors and vibrant accents for interactive elements.

### Brand Colors

| Role          | Color Name | Hex       | Swatch                                                          |
| :------------ | :--------- | :-------- | :-------------------------------------------------------------- |
| **Primary**   | Black      | `#000000` | ![#000000](https://placehold.co/100x40/000000/000000?text=+)    |
| **Primary**   | White      | `#FFFFFF` | ![#FFFFFF](https://placehold.co/100x40/FFFFFF/FFFFFF?text=+)    |
| **Secondary** | Gray       | `#CCCCCC` | ![#CCCCCC](https://placehold.co/100x40/CCCCCC/CCCCCC?text=+)    |
| **Accent 1**  | Hexa Green | `#00C6AB` | ![Hexa Green](https://placehold.co/100x40/00C6AB/00C6AB?text=+) |
| **Accent 2**  | Hexa Pink  | `#FF346B` | ![Hexa Pink](https://placehold.co/100x40/FF346B/FF346B?text=+)  |

### Color Gradients & Scales

#### **Hexa Green (Primary Accent)**

Used for primary calls-to-action (CTAs), success states, and active elements.

| Weight  | Hex           | Usage                 |
| :------ | :------------ | :-------------------- |
| 900     | `#00907D`     | Darkest               |
| 800     | `#00A08B`     |                       |
| 700     | `#00B29A`     |                       |
| **600** | **`#00C6AB`** | **Primary Color/CTA** |
| 500     | `#00DABC`     | Hover                 |
| 400     | `#00F0CF`     |                       |
| 300     | `#09FFDD`     |                       |
| 200     | `#AAFFF3`     |                       |
| 100     | `#D4FFF9`     | Lightest              |

#### **Hexa Pink (Secondary Accent)**

Used for secondary actions or to draw attention to specific features.

| Weight  | Hex           | Usage                   |
| :------ | :------------ | :---------------------- |
| 900     | `#DF003D`     | Darkest                 |
| 800     | `#F80044`     |                         |
| 700     | `#FF1555`     |                         |
| **600** | **`#FF346B`** | **Secondary Color/CTA** |
| 500     | `#FF5381`     | Hover                   |
| 400     | `#FF759A`     |                         |
| 300     | `#FF9AB5`     |                         |
| 200     | `#FFC3D3`     |                         |
| 100     | `#FFF0F4`     | Lightest                |

#### **System Colors (Gray, Text, Background, etc.)**

| Category         | Usage               | Hex       |
| :--------------- | :------------------ | :-------- |
| **Gray**         | Cancel Button       | `#9e9e9e` |
|                  | Cancel Button Hover | `#b5b5b5` |
|                  | Disabled Button     | `#e6e6e6` |
| **Text**         | Primary Text        | `#FFFFFF` |
|                  | Placeholder         | `#E6E6E6` |
| **Background**   | Default (Dark)      | `#28292D` |
|                  | Thin / Side Bar     | `#333336` |
| **Line/Border**  | Default             | `#555558` |
|                  | Hover               | `#656569` |
|                  | Disable             | `#3D3D3F` |
| **Input**        | Default             | `#38383B` |
|                  | Hover               | `#3A3A3E` |
|                  | Disable             | `#2B2B2E` |
| **Error/Delete** | Default             | `#FF7979` |
|                  | Hover               | `#FF9B9B` |

---

## 3. Typography

Clarity and readability are paramount. We use a clean, modern sans-serif typeface.

- **For Japanese:** Noto Sans JP
- **For English:** Inter

### **Headings**

Used for page titles, section titles, and menus.

- **Font Weight:** Bold (700)
- **Letter Spacing:** 0.05em

| Size Name         | Font Size |
| :---------------- | :-------- |
| Extra extra large | `24px`    |
| Extra large       | `21px`    |
| Large             | `18px`    |
| Base              | `16px`    |
| Medium            | `14px`    |
| Small             | `12px`    |
| Extra small       | `10px`    |

### **Body Text**

Used for all paragraphs and standard text content.

- **Font Weight:** Regular (400)
- **Letter Spacing:** 0.05em or 0.03em

| Size Name   | Font Size |
| :---------- | :-------- |
| Base        | `16px`    |
| Medium      | `14px`    |
| Small       | `12px`    |
| Extra small | `10px`    |

### **Line Height**

| Usage       | Line Height |
| :---------- | :---------- |
| Label       | `1.0`       |
| Description | `1.5`       |
| Note        | `1.8`       |
| Body        | `2.0`       |

---

## 4. Spacing

A consistent spacing scale based on a multiple of 8px is used for all layout, padding, and margin decisions to create visual rhythm.

| Name     | Spacing    | Pixels         |
| :------- | :--------- | :------------- |
| xs       | `2px`      | 2px            |
| sm       | `4px`      | 4px            |
| **md**   | **`8px`**  | **8px (Base)** |
| **st**   | **`16px`** | **16px**       |
| lg       | `24px`     | 24px           |
| xl       | `28px`     | 28px           |
| xxl      | `32px`     | 32px           |
| xxxl     | `40px`     | 40px           |
| plus     | `48px`     | 48px           |
| extended | `56px`     | 56px           |
| super    | `64px`     | 64px           |
| queen    | `72px`     | 72px           |
| king     | `80px`     | 80px           |

---

## 5. Elevation

Elevation is used to convey hierarchy and depth between surfaces. It is achieved through `box-shadow`.

| DPs | `box-shadow` CSS Rule                   | Common Use Cases                                   |
| :-- | :-------------------------------------- | :------------------------------------------------- |
| 0   | `none`                                  | Text buttons                                       |
| 1   | `0px 1px 2px 0px rgba(0, 0, 0, 0.16)`   | Search bar (static), Cards (static), Switch toggle |
| 2   | `0px 2px 3px 0px rgba(0, 0, 0, 0.16)`   | Contained button (static)                          |
| 4   | `0px 4px 5px 0px rgba(0, 0, 0, 0.16)`   | Top app bar                                        |
| 8   | `0px 8px 9px 0px rgba(0, 0, 0, 0.15)`   | Bottom navigation, Side sheets                     |
| 16  | `0px 16px 17px 0px rgba(0, 0, 0, 0.15)` | **Navigation Drawer**, **Modal Side/Bottom Sheet** |
| 24  | `0px 24px 25px 0px rgba(0, 0, 0, 0.04)` | **Dialogs**, **Pickers**                           |

![Elevation Scale](./images/elevation-scale.png)

---

## 6. Imagery

### Logo Marks

The Hexabase logo should be used consistently across all products. Both landscape and vertical orientations are available in black and white versions to ensure visibility on different backgrounds.

| Type          | Light Version                                                              | Dark Version                                                               |
| :------------ | :------------------------------------------------------------------------- | :------------------------------------------------------------------------- |
| **Landscape** | ![Hexabase Logo White](./images/hexabase-logo-white.png)                   | ![Hexabase Logo Black](./images/hexabase-logo-black.png)                   |
| **Vertical**  | ![Hexabase Logo Vertical White](./images/hexabase-logo-vertical-white.png) | ![Hexabase Logo Vertical Black](./images/hexabase-logo-vertical-black.png) |

---

## 7. Components

### Buttons

Buttons have clear states to provide user feedback.

![Button Component States](./images/button-component-states.png)

- **Primary Button:** Uses `Hexa Green`. For main confirmation actions (e.g., Save, Submit).
- **Cancel Button:** Uses the `Gray` scale. For secondary actions that close or cancel a flow.
- **Delete Button:** Uses the `Error/Delete` color. For destructive actions.
- **Secondary Button (Outline):** For alternative, non-primary actions.
- **States:** All buttons must have defined styles for `Default`, `Hover`, `Focused`, `Pressed`, and `Disabled`.

### Form Inputs (Input, Select, TextArea)

Form fields are designed for clarity and to guide the user.

![Input Field Component States](./images/input-field-component-states.png)

- **States:** All form fields must have defined styles for:
  - **Default:** The standard, inactive state.
  - **Focused:** When the user clicks into the field. Border becomes `Hexa Green`.
  - **Hover:** When the user mouses over the field.
  - **Error:** When validation fails. Border and error message use the `Error/Delete` color.
  - **Disabled:** The field is not interactive.
- **Labels:** Labels can be positioned vertically (above the input) or parallel (to the left of the input).

### Tags

Tags are used to highlight information, such as status or categories. Importance is conveyed through color.

![Tag Component Examples](./images/tag-component-examples.png)

| Importance | Color                  | Example |
| :--------- | :--------------------- | :------ |
| High       | Primary (`Hexa Green`) | `Admin` |
|            | Pink (`Hexa Pink`)     | `Admin` |
|            | Gray                   | `Admin` |
| Low        | Black                  | `Admin` |

### Dialogs

Dialogs are small modal windows that confirm user actions, often destructive ones. They require an explicit user choice to be dismissed.

![Dialog Component Examples](./images/dialog-component-examples.png)

### Modals

Modals are used for focused tasks like editing or creating complex data without navigating away from the current page.

![Modal Component Examples](./images/modal-component-examples.png)

#### **Sizing**

Modals and dialogs have defined sizes to ensure consistency.

| Type              | min-width | max-width | min-height | max-height |
| :---------------- | :-------- | :-------- | :--------- | :--------- |
| **Dialog Box**    | `400px`   | `670px`   | `160px`    | `60vh`     |
| **Modal/default** | `670px`   | `82vw`    | `230px`    | `60vh`     |
| **Modal/large**   | `82vw`    | -         | `60vh`     | `84vh`     |

#### **Padding**

Consistent internal padding is crucial for the readability and usability of modals.

- **Header:** `24px` padding on all sides.
- **Content Body:** `24px` padding top, left, and right.
- **Footer:** `32px` padding top and bottom, `24px` left and right.
- **Gutter between buttons:** `16px`.

![Modal Padding Specification](./images/modal-padding-specification.png)

### Icons

Icons are simple, clean, and immediately recognizable. Use icons from the defined library for consistency.

![Example Icons](./images/example-icons.png)

Common icons include:

- `icon_manage-search`
- `icon_bar-chart`
- `icon_insert-comment`
- `icon_arrow-left` / `icon_arrow-right`
- `icon_check` / `icon_close`
- `icon_calendar`
- `icon_db`
