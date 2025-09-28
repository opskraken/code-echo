# ✅ CodeEcho TODOS

A running list of upcoming tasks, improvements, and ideas for the CodeEcho CLI.

---

## 🔹 Stage 1 – Core Scanning (DONE ✅)
- [x] Basic CLI with Cobra
- [x] `scan` command with XML/JSON/Markdown output
- [x] Exclude default directories (`.git`, `node_modules`, etc.)
- [x] Include/exclude file extensions
- [x] File processing options:
  - [x] Remove comments
  - [x] Strip empty lines
  - [x] Compress whitespace
- [x] Directory tree generation
- [x] `--no-content` flag (structure only)
- [x] Richer metadata in XML/JSON/Markdown
- [x] Auto-named output files
- [x] `version` command

---

## 🔹 Stage 2 – Quality & Polish (IN PROGRESS 🛠️)
- [ ] Better language detection (beyond file extension)
- [ ] Smarter binary/text detection
- [ ] More resilient error handling (skip unreadable files gracefully)
- [ ] Configurable defaults via `.codeecho.yaml`
- [ ] Improved CLI output (progress indicators, better summaries)

---

## 🔹 Stage 3 – Documentation Helpers (UPCOMING 📖)
- [ ] `doc` command (currently stubbed)
  - [ ] Generate `README.md`
  - [ ] Generate `OVERVIEW.md`
  - [ ] Generate `API.md`
- [ ] Dependency detection (`package.json`, `go.mod`, `requirements.txt`, etc.)
- [ ] Insert dependency summary into scan output
- [ ] Support project badges (language breakdown, file counts, etc.)

---

## 🔹 Stage 4 – Ecosystem Integration (PLANNED 🚀)
- [ ] Export to LLM-friendly formats (e.g. OpenAI JSONL, Anthropic format)
- [ ] VSCode extension integration
- [ ] GitHub Action: auto-generate repo context file on push
- [ ] NPM/Go module detection for dependency graphs

---

## 🔹 Stage 5 – Stretch Goals (IDEAS 💡)
- [ ] Syntax-aware comment stripping (AST-based, safer than regex)
- [ ] File similarity detection / duplicate finder
- [ ] Code complexity metrics
- [ ] Incremental scans (cache + only changed files)
- [ ] Remote repo scanning (GitHub/GitLab API)

---

## 📝 Notes
- Current release: **v0.2.0 (planned)**
- Priorities right now:
  - Polish `scan` output
  - Finalize config system
  - Implement real `doc` command
- Feedback-driven development: issues/discussions welcome!
