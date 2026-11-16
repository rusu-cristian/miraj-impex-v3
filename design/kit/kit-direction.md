# Impex Line Panel: direction, copy rules, assets (from the Design Kit)

## 2. Direction (Unslop)

- **Product:** a daily record of each breeder-flock series through its farm cycles (growth → transfer → laying → sale), compared with each farm's guide.
- **Character:** alert, exact, mechanical.
- **World:** the hall's climate controller. Light steel panels, engraved uppercase captions, graphite keys, status lamps, IBM Plex Sans and Mono.
- **Colour has three jobs:**
  1. **Record type:** a fixed icon + colour, shared across the product.
  2. **Status:** a lamp + a word.
  3. **Steel neutrals** for everything else.
- **Moments:**
  - dark Readouts for key figures
  - guide charts (bright green actual, dashed guide, dim band)
  - the full cycle track
- **Signature: the cycle track.**
  - One lane per farm, each cycle a block.
  - Transfer links in violet, placement and sale marks at the ends.
  - The green "azi" line, with guide-week readouts.
- **Theme and density:** light only. Mixed density. No blinking, no animated charts.
- **Anti-references:** the ledger design (paper, serif, barn red), generic SaaS dashboards, glows, gradients, neon.
- **Template test:** passes. A generic farm app wouldn't have farm lanes with cycles, lamp-plus-word status or a fixed record-type code.

## 3. Copy rules

- **Language:** Romanian with full diacritics (ș ț with a comma below), informal "tu".
- **Buttons are verbs:** Salvează, Mută, Înregistrează vânzarea, Salvează prețurile, Începe ciclu nou.
- **Errors** name the section, the number and the fix.
- **No** hype, exclamation marks, greeting headlines, emoji or unverified claims.
- **Counts:** hen before rooster, with thousand spaces and the true minus.
- **Weeks:** the guide week of the farm's cycle.
- **Unknown** is `–` with a reason. Unpriced feed is "fără preț".
- **Sample data** comes from the Seria 2026 workbook, or is labelled sample. Names of people (Dr. Ceban, Ion Rusu) and dates after the workbook are illustrative.

**Claims to verify**

| Claim | Source | Status |
|---|---|---|
| Placement H1 ♀ 14 080, H2 ♀ 14 000, H5-1 ♂ 3 510 on 07.10.2025 | Telita, row 3 | confirmed |
| Delivery 07.10.2025, 8 000 kg | Hrana Telita, row 9 | confirmed (price now entered later by the owner) |
| Transfer to Hîrbovăț on 10.02.2026, so the laying cycle starts in guide week 1 | Hrana Hîrbovăț starts 10.02.2026 | inferred; confirm |
| Hîrbovăț sections H1.1, H1.2, H3.1, H3.2, H4.1, H4.2 | DataHirbovat | confirmed |
| Total eggs = Cat. 1 + 2 + 3 + Incubabile, "Jos" kept separate | import date sheet | inferred; confirm |
| Guide names (e.g. "Ross 308 ouat") | none | illustrative; replace with the real guides |

## 4. Assets

- **Exists:** `Seria_2026.xlsx` (the user's file, the source of sample data).
- **To supply:**
  - **Logo:** none; use the wordmark only, and don't generate a logo.
  - **Fonts:** from Google Fonts, self-hosted as woff2.
  - **Icons:** Lucide-style line icons. The record icons are fixed in RecordCode.
  - **Guide tables:** from the user, per guide type.
- **No images, so no image prompts.**

