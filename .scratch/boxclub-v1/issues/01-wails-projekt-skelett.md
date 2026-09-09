Status: ready-for-agent

# 01: Wails-Projekt-Skelett

**What to build:** Ein lauffähiges Wails-v2-Projekt mit dem Vanilla-Template, das per `wails dev` unter Windows **und** Linux fehlerfrei startet. Kein SQLite, keine Fachlogik. Dieses Ticket verifiziert allein die Toolchain, bevor irgendein Feature-Ticket losläuft.

**Blocked by:** None (kann sofort starten)

## Acceptance Criteria

- [ ] Wails-Projekt initialisiert (Vanilla-Template) und im Repository eingecheckt
- [ ] `wails dev` startet ohne Fehler unter Windows und zeigt den Default-Screen im WebView
- [ ] `wails dev` startet ohne Fehler unter Linux und zeigt den Default-Screen im WebView
- [ ] `wails build` erzeugt auf mindestens einer Dev-Plattform ein natives Binary (Rauchtest)
- [ ] `.gitignore` schließt Build-Artefakte, `frontend/node_modules`, `build/bin` etc. aus
- [ ] Repository-Root enthält weiterhin die bestehenden Artefakte (`AGENTS.md`, `CONTEXT.md`, `docs/`, `.scratch/`) unverändert

## Notes

Siehe [CONTEXT.md](../../../CONTEXT.md) für Vokabular und [ADR-0001](../../../docs/adr/0001-go-wails-fuer-desktop-gui.md) für die Stack-Begründung. Dokumentation der Prereqs erfolgt separat in Ticket 10 — hier reicht ein funktionierendes Skelett.
