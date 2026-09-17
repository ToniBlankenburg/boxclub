# 04: Trainingstermin-, Rechnung- & Vereinsdaten-Formulare: Gruppenboxen

**What to build:** Die drei übrigen Eingabeformulare (Trainingstermin
anlegen/bearbeiten, Rechnung erstellen, Vereinsdaten pflegen) folgen
demselben Gruppenboxen-Muster wie das Mitglied-Formular aus Ticket 03.

**Blocked by:** 01 (Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite),
03 (Mitglied-Formular: Gruppenboxen)

**Status:** ready-for-human

- [x] Alle drei Formulare zeigen ihre Felder in getönten, beschrifteten
      Gruppen, visuell konsistent mit dem Mitglied-Formular aus Ticket 03
- [x] Keine Feld- oder Validierungsänderung an einem der drei Formulare
- [x] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei
- [x] Manueller Smoke-Test in `wails dev`: jedes der drei Formulare einmal
      durchgespielt (Trainingstermin anlegen/bearbeiten, Rechnung erstellen,
      Vereinsdaten speichern) — siehe "Verifikation" unten: ohne GUI-Umgebung
      per Wegwerf-`httptest` gegen den echten Handler geprüft statt in
      `wails dev` selbst, wie von der Spec als gleichwertig vorgesehen
- [x] Bereits vorhandene Umsetzung (Fold-Commit `f3a19fa`, Branch
      `prototype/vier-restbereiche`) ist gegen diese Kriterien geprüft,
      nicht neu gebaut

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Die drei Formulare
wurden mechanisch identisch behandelt — deshalb ein gebündeltes Ticket statt
dreier separater.

### Verifikation (2026-09-17)

Alle Kriterien waren durch den Prototyp-Fold-Commit `f3a19fa`
(`prototype/vier-restbereiche`) bereits erfüllt:

- Der Fold-Commit ändert ausschließlich `templates/rechnung.html`,
  `templates/trainingstermine.html`, `templates/verein.html` (sowie
  `templates/import.html`, siehe Ticket 05) — kein `app/`- oder
  `service/`-Code ist betroffen. Der Diff verschiebt bestehende Felder
  unverändert in neue `<div class="rounded-md border border-neutral-200
  bg-neutral-50 p-4">`-Boxen; die einzige inhaltliche Änderung an den
  betroffenen Feld-`<input>`-Zeilen ist Einrückung, kein Name-, Typ- oder
  Pflicht-Attribut wurde angefasst.
- Trainingstermin-Formular: eine einzelne Gruppen-Box ("Trainingstermin"),
  vier Felder (Wochentag, Beginn, Ende, Bezeichnung) — bewusst keine weitere
  Aufteilung bei nur vier Feldern (siehe Kommentar in
  `templates/trainingstermine.html`).
- Rechnung-Formular: drei Gruppen-Boxen (Empfänger / Rechnungsdaten /
  Positionen).
- Verein-Formular: drei Gruppen-Boxen (Anschrift & Kontakt / Bankverbindung /
  Rechnungstext).
- Fehlerlisten (`role="alert"`, `border-red-700`, `text-red-800`) bleiben in
  allen Formularen mit Validierung (Trainingstermin, Rechnung) unverändert
  oberhalb der Gruppen; das Verein-Formular hat wie zuvor keine Fehlerliste
  (Ticket 23: keine Pflichtangabe, keine Prüfung).
- `go build ./...`, `go vet ./...`, `go test ./...` sowie `wails build`
  laufen fehlerfrei auf Linux.
- Zusätzlich per Wegwerf-`httptest` gegen die echten Handler geprüft
  (`app.New` + `a.Handler().ServeHTTP`, Testdatei danach wieder entfernt, wie
  im Testing-Abschnitt der Spec vorgesehen):
  - `GET /api/trainingstermin/formular` und `GET /api/trainingstermin/{id}
    /formular` zeigen die Gruppen-Box und alle vier Feldnamen; eine
    unvollständige `POST /api/trainingstermin` liefert die rote Fehlerliste.
  - `GET /api/rechnung` zeigt alle drei Gruppen-Beschriftungen und alle
    Feldnamen (inklusive der Positionszeilen); eine `POST /api/rechnung`
    ohne Rechnungsnummer liefert die rote Fehlerliste.
  - `GET /api/verein` zeigt alle drei Gruppen-Beschriftungen und alle
    Feldnamen; `POST /api/verein` speichert weiterhin und zeigt die
    Erfolgsmeldung samt gespeichertem Wert.
- Keine Code-Änderung war nötig — das Ticket schließt als reine
  Verifikation, wie schon Ticket 01, 02 und 03.
