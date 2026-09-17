# 03: Mitglied-Formular: Gruppenboxen

**What to build:** Das Formular zum Anlegen/Bearbeiten eines Mitglieds zeigt
seine Felder in vier beschrifteten, optisch getönten Gruppen: Person,
Kontakt & Anschrift, Mitgliedschaft, Sonstiges.

**Blocked by:** 01 (Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite)

**Status:** ready-for-human

- [x] Alle bisherigen Felder sind vorhanden und funktionieren wie zuvor, nur
      neu gruppiert — keine Feld- oder Validierungsänderung
- [x] Vier Gruppen sind visuell erkennbar (getönte Box, eigene Beschriftung)
      und decken Person / Kontakt & Anschrift / Mitgliedschaft / Sonstiges ab
- [x] Fehlerliste bei ungültiger Eingabe bleibt wie gehabt rot und oberhalb
      der Gruppen sichtbar
- [x] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei,
      insbesondere `service/member_service_test.go` unverändert grün, da
      keine Fachlogik betroffen ist
- [x] Manueller Smoke-Test in `wails dev`: Mitglied anlegen und bearbeiten,
      inklusive eines absichtlichen Validierungsfehlers (siehe "Verifikation"
      unten: ohne GUI-Umgebung per Wegwerf-`httptest` gegen den echten
      Handler geprüft statt in `wails dev` selbst, wie von der Spec als
      gleichwertig vorgesehen)
- [x] Bereits vorhandene Umsetzung (Fold-Commit `8fa5778`) ist gegen diese
      Kriterien geprüft, nicht neu gebaut

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Etabliert das
Gruppenboxen-Muster, das Ticket 04 auf die übrigen Formulare überträgt.

### Verifikation (2026-09-17)

Alle Kriterien waren durch den Prototyp-Fold-Commit `8fa5778`
(`prototype/stilrichtungen`) bereits erfüllt:

- Der Diff des Fold-Commits für `templates/mitglied_formular.html` verschiebt
  jedes bestehende `{{template "feld" ...}}` unverändert in eine von vier
  neuen `<div class="rounded-md border ... bg-neutral-50 p-4">`-Boxen
  (Person / Kontakt & Anschrift / Mitgliedschaft / Sonstiges) — kein Feld
  wurde hinzugefügt, entfernt oder in Name/Typ/Pflicht-Attribut verändert;
  einzige Nebenänderung ist der bereits in Ticket 01 verifizierte
  Farbtausch Rot → `var(--akzent)`.
- Die Fehlerliste (`role="alert"`, `border-red-700`, `text-red-800`) steht
  unverändert vor dem gruppierten Grid, nicht darin.
- `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei.
- Zusätzlich per Wegwerf-`httptest` gegen den echten Handler geprüft
  (`app.New` + `a.Handler().ServeHTTP`, Testdatei danach wieder entfernt,
  wie im Testing-Abschnitt der Spec vorgesehen):
  - Anlege-Formular (`GET /api/mitglied/formular`) enthält alle vier
    Gruppen-Beschriftungen, mindestens vier getönte Boxen und alle
    bisherigen Feldnamen.
  - Eine ungültige Eingabe (`POST /api/mitglied` ohne Nachname) liefert die
    rote Fehlerliste oberhalb der ersten Gruppen-Box.
  - Nach gültiger Anlage zeigt das Bearbeiten-Formular
    (`GET /api/mitglied/{id}/formular`) weiterhin alle vier Gruppen, die
    Mitglieds-ID in der Überschrift und die schreibgeschützten Felder
    (Geburtsdatum, Eintritt).
- Keine Code-Änderung war nötig — das Ticket schließt als reine
  Verifikation, wie schon Ticket 01 und 02.
