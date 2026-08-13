# AGENTS.md — Instructions for AI Coding Agents

This file applies to **every AI agent** working in this repository — Claude Code, Cursor, OpenCode/Hermes, Antigravity, or any other tool, and **every model tier**, from free/low-cost models (e.g. DeepSeek V4 Flash) up to Claude Opus 5. Read this file fully before writing any code, every session, even if you believe you remember it from a previous session.

If anything in this file conflicts with a user instruction mid-task, **stop and flag the conflict** rather than silently choosing one.

---

## 0. Read Order (mandatory, every session)

1. `AGENTS.md` (this file)
2. `rexio-city-v1.md` (or the current version's PRD) — the product spec
3. `DESIGN.md` — visual/design system
4. `DECISIONS.md` — why past architecture choices were made (do not re-litigate or silently reverse these)
5. `WORKLOGS.md` — read the **last 5 entries minimum** to understand what the previous session left in-progress or broken
6. `.env.example` for any service you will touch

Do not start writing code before completing this read order.

---

## 1. Single-Agent-at-a-Time Rule

Only one agent works on this repository at any given time. You do not need to handle concurrent-edit conflicts. You do need to assume a **different agent, possibly a much weaker model**, will pick up work after you — write code and notes accordingly (clear, well-commented, no cleverness that only you would understand).

---

## 2. Branching Rules — Non-Negotiable

- **Always push to `dev`. Never push directly to `main`.**
- `main` is only updated via a manual pull request that the human project owner (Sijan) reviews and merges himself.
- Before pushing, run all applicable checks locally if possible (build, vet, lint, test) — do not rely solely on CI to catch basic errors.

---

## 3. Secrets — Absolute Rules

- **Never commit `Private-Info.md`.** Never commit any `.env` file. Both must remain in `.gitignore` at all times — if you ever notice either is missing from `.gitignore`, add it immediately as your first action, before any other work.
- Never print, log, or paste raw secret values into commit messages, code comments, `WORKLOGS.md`, or any other tracked file.
- If a task requires a new secret/API key, do **not** invent a placeholder value and hardcode it. Add the variable name (with a dummy value) to the relevant `.env.example`, and write a note in `WORKLOGS.md`: "New env var required: `X` — needs real value from project owner."
- If you ever discover a secret already committed in git history, stop, do not try to fix it yourself by force-pushing history rewrites, and flag it clearly in `WORKLOGS.md` as urgent.

---

## 4. Architecture Rules — Do Not Violate

These are binding decisions from `DECISIONS.md` / the PRD. Do not "improve" or bypass them without an explicit new instruction from the project owner:

1. **The `frontend/` (Next.js) app never calls Supabase directly.** All data access goes through Next.js API routes, which call the Go backend.
2. **The `admin/` (Vite) app never calls Supabase directly either.** It calls its own separate admin backend (`backend/go/cmd/admin`).
3. **The Go backend (`cmd/api`) is the only thing that talks to the main Supabase project and to Cloudflare R2.**
4. **The admin backend (`cmd/admin`) is the only thing that talks to the separate admin Supabase project.**
5. **CORS on the main backend only allows the Next.js server origin** (plus whitelisted `localhost` ports in dev). Never widen this to `*` or to arbitrary origins "to make testing easier" — fix the actual origin instead.
6. **No Supabase Auth.** Auth is fully custom (argon2id passwords, custom OAuth2 for Google/GitHub, custom JWT + refresh token sessions). No Supabase default email verification either — email goes through Brevo.
7. **DMs use a custom Go WebSocket server, not Supabase Realtime.**
8. **DM encryption is at-rest (AES-256-GCM) only.** Do not implement or claim end-to-end encryption in V1 — this would misrepresent the product to users.
9. Prefer **DB-driven configuration** (a `settings`/`config` table) over hardcoded `.env` values for business logic (feature flags, rate limits, feed-scoring weights). `.env` is reserved for genuinely static/secret values, and can act as a fallback default.
10. Media URLs follow the structured, non-guessable pattern in the PRD (e.g. `cdn-city.rexio.pro/post/{random-id}`) — never expose sequential/guessable IDs in public media URLs.

---

## 5. Code Style

- Write code as a competent human engineering team would — clear naming, no unnecessary abstraction, no AI-generated boilerplate filler. This is a real engineering standard, not an attempt to disguise authorship (this project is openly built with AI agents — see `BRANDING.md`).
- **Complex or non-obvious logic must have English comments** explaining the *why*, not just the *what*. Assume the next reader (human or a weaker model) has no memory of this session.
- Go: run `gofmt`/`go vet`/`golangci-lint` before considering a task done.
- Frontend/Admin: run `eslint` and `tsc --noEmit` before considering a task done.
- Never commit commented-out dead code "just in case" — delete it, git history preserves it.

---

## 6. Testing

- Backend (Go): critical-path logic — auth, feed scoring, follow system, DM handling — requires accompanying `_test.go` unit tests. A PR that adds this kind of logic without tests is incomplete.
- Run `go test ./...` and ensure it passes before ending your session.
- Frontend: no mandatory E2E in V1. Manually verify the flow you changed and note what you checked in `WORKLOGS.md`.
- See `TESTING.md` for the manual QA checklist for frontend flows.

---

## 7. WORKLOGS.md — Mandatory Session Log

At the **start** of a session, append an entry:

```
## [YYYY-MM-DD HH:MM, timezone] — Agent: <tool name> — Model: <model name>
### Picking up:
- <what you read in WORKLOGS.md that's relevant>
### Plan for this session:
- <bullet list>
```

At the **end** of a session (or whenever you stop, even mid-task), append:

```
### Done:
- <what was completed>
### Left incomplete / blocked:
- <anything not finished, and why>
### Notes for next agent:
- <anything a following session needs to know>
```

Never skip this. A session with no `WORKLOGS.md` entry is considered undocumented work and should be treated with suspicion by the next agent.

---

## 8. Commits

- Conventional commit style: `feat:`, `fix:`, `chore:`, `refactor:`, `test:`, `docs:`.
- No fabricated human co-authors. Do not add `Co-authored-by:` lines for people who did not actually contribute to that specific change.
- Commit messages describe what changed and why, in plain English.

---

## 9. When Uncertain

If a task is ambiguous, or seems to require violating any rule in Section 4, **do not guess and proceed silently**. Write the ambiguity into `WORKLOGS.md` under "Notes for next agent" / "Left incomplete" and stop that portion of the work, rather than making an architectural decision that isn't yours to make.
