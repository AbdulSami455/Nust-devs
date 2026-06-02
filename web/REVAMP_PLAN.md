# NUST Devs Frontend Revamp Plan

This document outlines a comprehensive plan to revamp the NUST Devs frontend, drawing inspiration from the modern, data-rich, and community-focused design of [LebHub](https://lebhub.xyz/), while adapting it specifically for the NUST developer community.

## 1. Design Direction & Visual Identity

- **Inspiration:** LebHub's clean, dark-mode-first aesthetic, bento-box layouts, and data-heavy dashboard feel.
- **Branding:** NUST-branded color palette. Instead of generic monochrome, use a deep Navy Blue background for dark mode with Gold/Yellow accents for primary actions and highlights, reflecting NUST's traditional colors.
- **Typography:** Continue using `Geist` (already configured) for a sleek, modern, and highly legible interface.
- **Theme:** Implement full Dark/Light mode support (defaulting to Dark mode) using `next-themes`.

## 2. Information Architecture & Page Structure

The navigation will be simplified and centralized, with a focus on discovery and community metrics.

- **`/` (Home / Dashboard)**: The central hub. Overview stats, community activity graph, live feed, and spotlight.
- **`/developers` (Directory)**: Grid of all developers with advanced filtering.
- **`/developers/[username]` (Dev Profile)**: The "Dev Card" view, individual stats, and contribution graph.
- **`/leaderboard` (Rankings)**: Gamified rankings with a podium for the top 3.
- **`/projects` (Showcase)**: Trending and top repositories by NUST devs.
- **`/stats` (Analytics)**: Deep dive into languages and community growth.
- **`/admin`**: (Existing) Protected area for managing tracked developers.

## 3. Component Library Revamp

The project currently uses `shadcn/ui`. We will expand this significantly:

- **Add `Command` (cmdk)**: For a global `Cmd+K` search palette (search devs, repos, pages).
- **Add `Chart` (Recharts)**: For community activity graphs and language distribution.
- **Add `Skeleton`**: Replace text-based "Loading..." states with polished skeleton loaders.
- **Add `Avatar`**: Standardize profile pictures with fallbacks.
- **Add `Tabs`**: For switching views (e.g., Top 10 vs Top 100 on the leaderboard).
- **Add `Tooltip`**: Essential for contribution graphs (GitHub-style heatmaps).

## 4. New Features to Add

1. **Global Cmd+K Search**: Instantly find any developer or project from anywhere on the site.
2. **Dev Card Generator**: A visually appealing, shareable card on developer profiles that users can export as an image to share on social media (Twitter/LinkedIn).
3. **Live Activity Feed**: A scrolling feed of the latest commits, PRs, and issues opened by NUST developers, making the platform feel "alive".
4. **Community Activity Graph**: An aggregated chart showing total community contributions over the year/month.
5. **Developer Spotlight**: A featured section on the homepage highlighting a specific developer (random, top trending, or newly joined).
6. **Time-bound Stats**: "Last 24 Hours" and "Year to Date" metrics (Commits, PRs, Active Repos).

## 5. Page-by-Page Redesign Specs

### Home Page (`/`)
- **Hero**: "NUST Devs on GitHub" with a primary CTA: "Create your Dev Card".
- **Bento Grid Stats**: Year-to-Date metrics (Active Devs, Commits, PRs, Repos).
- **Recent Highlights**: Text summary (e.g., "This week: 45 PRs opened, React is trending").
- **Activity Graph**: A large bar or area chart showing community activity over the last 30 days.
- **Live Feed**: A sidebar or bottom section showing real-time events (e.g., "UserX pushed to RepoY").
- **Spotlight**: A featured developer card.

### Leaderboard (`/leaderboard`)
- **Podium UI**: Special visual treatment (larger avatars, gold/silver/bronze medals) for the top 3 developers.
- **Tabs**: Toggle between "Top 10", "Top 100", and "All".
- **Filters**: Sort by Activity Score (default), Commits, PRs, Stars, or Followers.
- **Trend Indicators**: Show if a developer moved up or down in rank compared to last week (requires backend historical data).

### Developers Directory (`/developers`)
- **Dev Cards**: Redesign the current basic cards into mini "Dev Cards" with better typography, tech stack badges, and contribution sparklines.
- **Advanced Filters**: Filter by primary language, campus/department (if added to schema), or verification status.

### Developer Profile (`/developers/[username]`)
- **The Dev Card**: A large, beautifully styled card containing their stats, avatar, and rank.
- **Share Button**: "Share to Twitter/X" or "Download Image".
- **Contribution Heatmap**: GitHub-style green/blue contribution squares.
- **Top Repos**: Grid of their best work.

## 6. Charts & Visualizations

- **Community Pulse (Home)**: Area chart showing aggregate contributions over time.
- **Contribution Heatmap (Profile)**: Calendar heatmap for individual developers.
- **Language Distribution (Stats)**: Donut chart showing the most popular languages across all NUST public repos.

## 7. UX Improvements

- **Loading States**: Implement `Skeleton` components for all data fetches to prevent layout shift.
- **Navigation**: Sticky blurred header. On mobile, implement a bottom navigation bar or a clean hamburger menu.
- **Dark Mode**: Seamless toggle with `next-themes` to prevent hydration mismatch flashes.
- **Empty States**: Beautifully illustrated empty states for searches with no results.

## 8. Phased Milestones & Priorities

- **Milestone 1: Foundation & Theming (High Priority)**
  - Install `next-themes`, configure NUST Dark/Light palettes.
  - Setup global `Cmd+K` search shell.
  - Upgrade basic UI components (`Avatar`, `Skeleton`).

- **Milestone 2: Dashboard & Home Page (High Priority)**
  - Build the Bento grid for YTD and 24h stats.
  - Integrate `recharts` and build the Community Activity Graph.
  - Build the Developer Spotlight component.

- **Milestone 3: Profiles & Dev Cards (Medium Priority)**
  - Redesign `/developers/[username]` to feature the new Dev Card.
  - Implement the GitHub-style contribution heatmap.
  - Add "Share/Export" functionality for the Dev Card.

- **Milestone 4: Leaderboard & Directory (Medium Priority)**
  - Build the Podium UI for `/leaderboard`.
  - Add Tabs and advanced sorting.
  - Redesign the `/developers` grid to use mini Dev Cards.

- **Milestone 5: Live Feed & Polish (Low Priority / Dependent)**
  - Implement the Live Activity Feed (requires new backend endpoint).
  - Finalize mobile responsiveness and animations (Framer Motion).

## 9. Tech Stack Recommendations

- **Theming**: `next-themes` for robust dark mode.
- **Charts**: `recharts` (via shadcn/ui) for responsive, accessible charts.
- **Search**: `cmdk` for the command palette.
- **Icons**: `lucide-react` (already standard with shadcn).
- **Dates**: `date-fns` for relative time formatting (e.g., "2 hours ago" in the activity feed).
- **Animations**: `framer-motion` for smooth layout transitions (e.g., sorting the leaderboard).

## 10. Backend API Requirements (To Support New UI)

To fully realize this LebHub-inspired revamp, the Go backend will need the following new or updated endpoints:

1. **`GET /api/v1/activity/recent`**: Returns a list of recent GitHub events (pushes, PRs, issues) across all tracked developers for the Live Activity Feed.
2. **`GET /api/v1/stats/community-activity`**: Returns aggregated contribution counts (e.g., grouped by day for the last 30 days or 12 months) for the Community Activity Graph.
3. **`GET /api/v1/stats/summary`**: Returns YTD and 24-hour stats (total commits, PRs, active devs) to populate the Bento grid.
4. **`GET /api/v1/developers/spotlight`**: Returns a featured developer (could be random, highest rising, or manually curated).
5. **Historical Rankings (Optional)**: For trend indicators on the leaderboard, the backend would need to return previous rank alongside current rank.