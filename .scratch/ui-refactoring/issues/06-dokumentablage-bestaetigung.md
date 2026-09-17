# 06: Dokumentablage: Bestätigung unverändert

**What to build:** Kein Code-Umbau — dieses Ticket dokumentiert und
bestätigt formal, dass die Dokumentablage (Vertrag) bei ihrer bestehenden
Statuskarte bleibt, nachdem Prototyp-Runde 4 das geprüft und bestätigt hat.

**Blocked by:** 01 (Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite)

**Status:** ready-for-agent

- [ ] Dokumentablage zeigt weiterhin die bestehende Statuskarte mit
      Datei-Aktionen (Upload/Entfernen/Export), keine strukturelle Änderung
- [ ] Trägt die neue Akzentfarbe aus Ticket 01 konsistent (nur Farbtoken,
      keine Struktur)
- [ ] Verweis auf die Verdict-Datei bzw. den Prototyp-Branch
      `prototype/dokument-ablage` ist in den Ticket-Comments festgehalten,
      damit ein späterer Leser nicht erneut über eine Umgestaltung
      nachdenkt
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei
- [ ] Manueller Smoke-Test in `wails dev`: Dokumentablage an einer
      Mitgliedschaft mit und ohne abgelegten Vertrag ansehen

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Prototyp-Runde 4
(Branch `prototype/dokument-ablage`, Commits `5fb7a4d`/`2c9469a`) hat drei
Varianten für `templates/dokument.html` verglichen und **Variante A (die
bestehende Statuskarte)** bestätigt — kein Fold nötig, `main` blieb dadurch
unverändert. Dieses Ticket macht diese Entscheidung im Tracker explizit,
statt sie nur in einer Verdict-Datei auf einem Branch stehen zu lassen.
