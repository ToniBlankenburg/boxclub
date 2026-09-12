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

**SQLite schema (three tables):** nothing is seeded — a fresh database is empty.

- `mitglied(id, vorname, nachname, geburtsdatum, adresse, postleitzahl, ort, email, telefon, iban, geschlecht, google_bewertung, digital, rueckstand, rueckstand_notiz)` — person master data; one row per person even across re-entries. The postal address is three separate columns (`adresse` is street + house number) — see `CONTEXT.md` → Anschrift
- `mitgliedschaft(id, mitglied_id, anmeldedatum NULL, eintritt, kuendigungsdatum NULL, austritt NULL, anmeldegebuehr_cents, beitrag_monatlich_cents, ruhend)` — time-bound membership period; `austritt IS NULL` means currently active. `anmeldedatum` is the day the application was handed in and usually precedes `eintritt`; it is **not** the same day (see `CONTEXT.md` → Anmeldedatum)
- `trainingstermin(id, wochentag, beginn, ende NULL, bezeichnung, archiviert)` — the club's weekly schedule, maintained in its own view. Entries exist on their own, independent of any member; an entry that is no longer offered is **archived, never deleted** (ADR-0008)
- `mitgliedschaft_trainingstermin(mitgliedschaft_id, trainingstermin_id)` — a membership is signed up for zero to three schedule entries. Replaces the former `trainingsslot` free-text table
- `dokument(id, mitglied_id NULL, mitgliedschaft_id NULL, art, dateiname, inhalt BLOB, erstellt_am)` — PDFs stored in the database, never as files beside it (ADR-0007). Exactly one of the two foreign keys is set: a *Vertrag* hangs on the **membership**, a *Rechnung* on the **person**. `inhalt` must never appear in a query that returns more than one row
- `vereinsdaten(...)` — a single row: the club's own name, address, bank details and invoice footer. The only data in the app that is not about a member

The **Trainingsfrequenz** (1×/2×/3× per week) is never stored: it is the number of schedule entries a membership is signed up for, derived on every read — zero mean "no frequency", not "1×". A rejoin starts with none; the old membership keeps its own (see `CONTEXT.md` → Trainingstermin). An **archived** entry still counts while the signup stands: tidying the schedule must not silently change what members agreed to.

A **Rechnung has no table**. Recipient, line items, tax rate and number are entered, rendered into a PDF and kept only as a `dokument` row — there is no invoice record, no status, no payment (ADR-0009). The invoice number is typed by the user, not issued by the app. For a recipient who is not a member, the PDF is handed out instead of stored.

The Excel importer **never creates schedule entries**. It matches the free text of `Training - 1/2/3` against existing ones and reports the rest through the error report from Ticket 09 (ADR-0008).

`iban` is stored as plain text: no validation, no format check, no SEPA export (ADR-0006). `geschlecht` is free text — the Excel value list is a typing aid, not a constraint. `google_bewertung` is two-valued (`CONTEXT.md` → Google-Bewertung). `digital` carries the Excel column of the same name **verbatim and without meaning in the model** — its semantics are unknown, so it gets no check, no dropdown and no glossary entry until they are.

The **Anmeldegebühr** is a one-off historical value in cents: it records what was actually paid at the time and is never recalculated. Like the fee it hangs on the membership, so a rejoin does not carry it over — a new period starts with neither Anmeldedatum nor Anmeldegebühr, just as it starts without training slots.

The monthly fee hangs on the **membership**, not on the person, and is individually agreed — there are no fee classes (see ADR-0005). 0 € is a valid fee, not a missing one. Euro input is converted to cents in `service.BeitragAusEuro`; there is no migration mechanism, so a schema change means deleting the dev database.

A **Kündigung** is recorded as two dates on the membership: `kuendigungsdatum` (the day it was declared) and `austritt` (the day it takes effect). The second is **never derived** from the first when storing — `SetKuendigung` writes exactly what it is given — and either may be absent: a declared cancellation without a date yet, or a historical exit whose declaration day nobody wrote down. `austritt` may lie in the future; that is the normal case during the Kündigungsfrist. Both are written together in `MemberService.SetKuendigung`, which replaces both columns at once. The **form** does propose one: `service.RegulaererAustritt` computes the club's regular notice period (`KuendigungsfristMonate` = 3 months, rounded up to month end) to prefill the empty field. That is a suggestion and binds nothing — no validation checks a stored `austritt` against it, and the field stays editable and clearable.

`ruhend` hangs on the **membership**, because only a running period can be suspended (`MemberService.SetRuhend`, `ErrNichtAktiv` otherwise). It is a flag beside the lifecycle, not a state in it — a suspended member is an active one nobody collects from. It is independent of the Rückstand: no direct debit goes out for a suspended member, so no *new* arrears can arise, but an existing one stays.

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

**Rechnungen und Finanzüberblick sind in v1** — aber nur in einem eng gezogenen Zuschnitt, siehe [ADR-0009](docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md). Eine *Rechnung* ist ein erzeugtes PDF über eine Leistung neben dem Beitrag (Einzeltraining), abgelegt als Dokument am Mitglied: ohne Rechnungsstatus, ohne Zahlungseingang, ohne Mahnung, mit von Hand eingetippter Nummer. Der *Finanzüberblick* ist das **Monatssoll** des laufenden Monats plus Mitgliederzahlen — eine Ableitung aus vorhandenen Beiträgen, kein Verlauf, weil es keine Beitragshistorie gibt.

**Das Zahlungsmodell bleibt ausgeschlossen.** [ADR-0006](docs/adr/0006-rueckstand-statt-bezahlt-bis.md) ist dadurch nicht abgelöst, sondern präzisiert: v1 hält den *Rückstand* als handgepflegtes Kennzeichen und kennt weder Zahlungen noch offene Posten noch eine Historie. Wer einer Rechnung ansehen will, ob sie bezahlt ist, schaut aufs Konto — dieselbe Grenze wie beim Beitrag.

Die Grenze zu *attendance tracking* verläuft dabei nicht dort, wo man sie vermutet: ein **gepflegter Katalog von Trainingsterminen**, dem Mitgliedschaften zugeordnet werden, ist v1-tauglich und löst den früheren Freitext in `trainingsslot.bezeichnung` ab ([ADR-0008](docs/adr/0008-trainingstermine-als-wochenplan.md)). Ausgeschlossen ist nur, **pro Termin festzuhalten, wer da war** — das ist die Anwesenheit, und die bringt eine Historie mit, die v1 nicht führt.
