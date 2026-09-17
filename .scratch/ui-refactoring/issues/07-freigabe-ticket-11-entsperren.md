# 07: Freigabe: Ticket 11 (Mac-Release-Build) entsperren

**What to build:** Bestätigt, dass alle UI-Refactoring-Tickets (01–06)
abgeschlossen sind, führt eine app-weite Abschlussprüfung durch und
aktualisiert den Blocker-Vermerk in `boxclub-v1/issues/11-mac-release-build.md`,
damit dieses Ticket starten darf.

**Blocked by:** 02 (Dashboard: Kennzahlenleiste), 03 (Mitglied-Formular:
Gruppenboxen), 04 (Trainingstermin-, Rechnung- & Vereinsdaten-Formulare:
Gruppenboxen), 05 (Excel-Import-Bericht: Kennzahlenleiste), 06
(Dokumentablage: Bestätigung unverändert) — 01 transitiv über alle fünf

**Status:** ready-for-agent

- [ ] Alle Tickets 01–06 stehen auf `ready-for-human` oder besser
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` und `wails build`
      laufen app-weit fehlerfrei
- [ ] Vollständiger manueller Durchlauf aller 6 Bereiche in `wails dev` ohne
      visuelle Inkonsistenz (z. B. vergessene Rot-Reste außerhalb der zwei
      bewusst roten Warnstellen)
- [ ] `boxclub-v1/issues/11-mac-release-build.md`: Blocker-Vermerk zum
      UI-Refactoring ist aktualisiert (z. B. "erledigt, siehe
      `.scratch/ui-refactoring/`" statt offen)

## Comments

### Kontext

Schließt das UI-Refactoring formal ab, siehe [Spec](../spec.md). Kein
eigener Code-Umbau — reine Verifikation und Freigabe.
