# 04: Trainingstermin-, Rechnung- & Vereinsdaten-Formulare: Gruppenboxen

**What to build:** Die drei übrigen Eingabeformulare (Trainingstermin
anlegen/bearbeiten, Rechnung erstellen, Vereinsdaten pflegen) folgen
demselben Gruppenboxen-Muster wie das Mitglied-Formular aus Ticket 03.

**Blocked by:** 01 (Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite),
03 (Mitglied-Formular: Gruppenboxen)

**Status:** ready-for-agent

- [ ] Alle drei Formulare zeigen ihre Felder in getönten, beschrifteten
      Gruppen, visuell konsistent mit dem Mitglied-Formular aus Ticket 03
- [ ] Keine Feld- oder Validierungsänderung an einem der drei Formulare
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei
- [ ] Manueller Smoke-Test in `wails dev`: jedes der drei Formulare einmal
      durchgespielt (Trainingstermin anlegen/bearbeiten, Rechnung erstellen,
      Vereinsdaten speichern)
- [ ] Bereits vorhandene Umsetzung (Fold-Commit `f3a19fa`, Branch
      `prototype/vier-restbereiche`) ist gegen diese Kriterien geprüft,
      nicht neu gebaut

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Die drei Formulare
wurden mechanisch identisch behandelt — deshalb ein gebündeltes Ticket statt
dreier separater.
