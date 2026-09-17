# 03: Mitglied-Formular: Gruppenboxen

**What to build:** Das Formular zum Anlegen/Bearbeiten eines Mitglieds zeigt
seine Felder in vier beschrifteten, optisch getönten Gruppen: Person,
Kontakt & Anschrift, Mitgliedschaft, Sonstiges.

**Blocked by:** 01 (Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite)

**Status:** ready-for-agent

- [ ] Alle bisherigen Felder sind vorhanden und funktionieren wie zuvor, nur
      neu gruppiert — keine Feld- oder Validierungsänderung
- [ ] Vier Gruppen sind visuell erkennbar (getönte Box, eigene Beschriftung)
      und decken Person / Kontakt & Anschrift / Mitgliedschaft / Sonstiges ab
- [ ] Fehlerliste bei ungültiger Eingabe bleibt wie gehabt rot und oberhalb
      der Gruppen sichtbar
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei,
      insbesondere `service/member_service_test.go` unverändert grün, da
      keine Fachlogik betroffen ist
- [ ] Manueller Smoke-Test in `wails dev`: Mitglied anlegen und bearbeiten,
      inklusive eines absichtlichen Validierungsfehlers
- [ ] Bereits vorhandene Umsetzung (Fold-Commit `8fa5778`) ist gegen diese
      Kriterien geprüft, nicht neu gebaut

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Etabliert das
Gruppenboxen-Muster, das Ticket 04 auf die übrigen Formulare überträgt.
