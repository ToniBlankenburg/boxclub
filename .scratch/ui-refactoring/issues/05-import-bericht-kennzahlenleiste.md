# 05: Excel-Import-Bericht: Kennzahlenleiste

**What to build:** Der Ergebnisbericht nach einem Excel-Import zeigt
übernommene Sätze, Fehler und Hinweise als Kennzahlenleiste (dasselbe
Muster wie das Dashboard aus Ticket 02) statt als Fließtext.

**Blocked by:** 02 (Dashboard: Kennzahlenleiste)

**Status:** ready-for-agent

- [ ] Bericht zeigt seine Kennzahlen (übernommene Sätze/Fehler/Hinweise) als
      Leiste, visuell konsistent mit dem Dashboard-Muster aus Ticket 02
- [ ] Die detaillierte Fehler-/Hinweisliste selbst bleibt inhaltlich
      unverändert und weiterhin einsehbar — nur die zusammenfassenden
      Kennzahlen ändern die Darstellung
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei,
      insbesondere die `importer/`-Tests unverändert grün
- [ ] Manueller Smoke-Test in `wails dev`: einen Excel-Import mit mindestens
      einem Fehler und einem Hinweis durchführen
- [ ] Bereits vorhandene Umsetzung (Fold-Commit `f3a19fa`) ist gegen diese
      Kriterien geprüft, nicht neu gebaut

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md).
