# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Boxclub Mitgliederverwaltung — a local macOS desktop app replacing an Excel spreadsheet for managing ~50–200 club members, built as a learning project in Go. Production target: macOS only. Development runs on Windows and Linux.

## Commands

```bash
# Install Wails CLI (once)
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Run in dev mode (hot reload, opens WebView)
wails dev

# Production build
wails build

# Run Go tests only (no Wails required)
go test ./...

# Run a single package's tests
go test ./service/...

# Run a single test function
go test ./service/... -run TestMemberService_Create
```

## Architecture

**Stack:** Go backend + Wails v2 (native WebView bridge) + Vanilla HTML/htmx/Tailwind frontend. No SPA framework, no custom npm pipeline beyond Wails' built-in step. SQLite via `modernc.org/sqlite` (pure Go, no CGo — keeps cross-compilation trivial).

**Two primary seams:**

- `service/` — `MemberService`: all business logic (CRUD, search/filter, payment status, membership lifecycle). Talks directly to `database/sql` — no repository interface, no separate domain/persistence models.
- `importer/` — `ExcelImporter`: parses `.xlsx` files via `excelize/v2`, returns rows + error report. Does **not** call `MemberService` itself; a thin orchestrator in the Wails layer wires them together.

**Adapter layers (not primary seams):**

- `app/` — HTTP handlers for the htmx frontend: thin 1:1 wrappers over `MemberService`/`ExcelImporter` that return HTML fragments. They hang off `assetserver.Options.Handler`; Wails bindings are **not** used (see ADR-0002).
- `templates/` — Go `html/template` files rendering htmx fragments (member rows, forms, error lists).

**SQLite schema (two tables):** nothing is seeded — a fresh database is empty.

- `mitglied(id, vorname, nachname, geburtsdatum, adresse, email, telefon, rueckstand, rueckstand_notiz)` — person master data; one row per person even across re-entries
- `mitgliedschaft(id, mitglied_id, eintritt, austritt NULL, beitrag_monatlich_cents)` — time-bound membership period; `austritt IS NULL` means currently active

The monthly fee hangs on the **membership**, not on the person, and is individually agreed — there are no fee classes (see ADR-0005). 0 € is a valid fee, not a missing one. Euro input is converted to cents in `service.BeitragAusEuro`; there is no migration mechanism, so a schema change means deleting the dev database.

The club collects by SEPA direct debit, so there is no payment status to track — only **Rückstand**: a two-valued flag (`in Ordnung` / `im Rückstand`) plus a free-text note, both hand-maintained and both on the **person**, so an exit does not clear a debt (see ADR-0006). There is no neutral third state, and no payment history in v1.

The database file lives in the user config dir (`~/Library/Application Support/Boxclub/boxclub.db` on macOS); `BOXCLUB_DB` overrides the path for development.

## Testing

Tests hit a real SQLite database (`t.TempDir()`-based temp file per test) — **no mocking**. Test via the service/importer API only; never reach into SQL directly.

- `service/member_service_test.go` is the reference test file — all subsequent tests mirror its setup pattern.
- `app/` handlers and htmx templates are not unit-tested; manual smoke test suffices.
- After any significant change: verify `wails dev` and `wails build` succeed on both Windows and Linux. A build failure on a dev platform blocks the ticket just like a failing test.

## Domain vocabulary

Use terms exactly as defined in `CONTEXT.md`. Key distinction: **Mitglied** (the person, permanent record) vs. **Mitgliedschaft** (a time-bounded active period). See `CONTEXT.md` for the full glossary.

Before working in an area, read the relevant ADR(s) in `docs/adr/`.

## Issue tracker

Issues and specs are local markdown files in `.scratch/<feature-slug>/`:

- Spec: `.scratch/<feature-slug>/spec.md`
- Tickets: `.scratch/<feature-slug>/issues/<NN>-<slug>.md` (numbered from `01`, one file per ticket)
- Each ticket has a `Status:` line near the top (`needs-triage`, `ready-for-agent`, `ready-for-human`, `wontfix`, etc.)
- Comments append under a `## Comments` heading at the bottom of the file

## Out of scope for v1

MoneyMoney CSV import/auto-matching, multi-user/login, cloud/web deployment, Windows/Linux release builds, UI test automation, attendance tracking, boxing-specific fields (license, weight class).
