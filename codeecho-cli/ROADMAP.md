# 🚀 CodeEcho Roadmap

_"Let your code speak back."_

CodeEcho is evolving in clear stages. Each milestone adds new capabilities while keeping the CLI simple and reliable.

---

## ✅ Stage 1 – Core Scanning

- [x] CLI scaffold with Cobra.
- [x] `scan` command to walk repos.
- [x] Output in `xml`, `json`, or `markdown`.
- [x] Flags: `--out`, `--format`, `--exclude`.

---

## ✅ Stage 2 – Expanded Scanning (Current)

- [x] File stats (size, modified time, line count).
- [x] Repo tree view in Markdown + XML.
- [x] `--no-content` flag to output structure only.
- [x] Language detection + syntax hinting.
- [x] Content processing options (`--compress-code`, `--remove-comments`, `--remove-empty-lines`).

---

## 🔹 Stage 3 – Documentation Helpers

- [ ] `doc` command to auto-generate:
  - README.md
  - OVERVIEW.md
  - API.md
- [ ] Dependency detection (`package.json`, `go.mod`, etc.).

---

## 🔹 Stage 4 – AI-Ready Boosts

- [ ] `--chunk` to split large files.
- [ ] `--compress` to strip comments/whitespace.
- [ ] Syntax highlighting in Markdown outputs.

---

## 🔹 Stage 5 – Productivity Enhancements

- [ ] Respect `.gitignore`.
- [ ] `--extensions` flag for file filtering.
- [ ] `--summary` command for high-level project overview.
- [ ] `version` command with build info.

---

## 🔹 Stage 6 – Advanced Features

- [ ] Secret/key detection.
- [ ] Auto changelog generator from commits.
- [ ] Knowledge Pack export (JSON + Markdown).
- [ ] Plugin system for custom parsers.

---

## 🔹 Stage 7 – Future Vision

- [ ] AI integration for instant docs/refactors.
- [ ] Continuous mode (update docs on every commit).
- [ ] Web dashboard with history + collaboration.
- [ ] Community marketplace for knowledge packs.
