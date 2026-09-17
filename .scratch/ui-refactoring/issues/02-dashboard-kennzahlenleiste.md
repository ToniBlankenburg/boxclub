# 02: Dashboard: Kennzahlenleiste

**What to build:** Das Dashboard zeigt Monatssoll und Mitgliederzahlen als
durchgehende horizontale Kennzahlenleiste statt des bisherigen
Kartenrasters.

**Blocked by:** 01 (Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite)

**Status:** ready-for-human

- [x] Dashboard zeigt seine Kennzahlen als eine zusammenhängende Leiste,
      kein Kartenraster mehr
- [x] Totes Kartenraster-Teil-Template ("dashboard-kachel"), Go-Typ
      `kachelDaten` und Template-Funktion `kachel` sind entfernt — keine
      toten Reste
- [x] Zahlen bleiben korrekt: Monatssoll und Mitgliederzahlen stimmen
      weiterhin mit den bisherigen Werten überein (keine
      Berechnungsänderung, nur Darstellung)
- [x] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei
- [x] Manueller Smoke-Test in `wails dev`: Dashboard öffnen, Werte mit
      einem bekannten Datensatz gegenprüfen (siehe Comments: ohne
      GUI-Umgebung per Wegwerf-`httptest` gegen den echten Handler mit
      einem bekannten Datensatz geprüft statt in `wails dev` selbst)
- [x] Bereits vorhandene Umsetzung (Fold-Commit `8fa5778`) ist gegen diese
      Kriterien geprüft, nicht neu gebaut

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Etabliert das
Kennzahlenleisten-Muster, das Ticket 05 (Excel-Import-Bericht) übernimmt.

### Verifikation (2026-09-17)

Alle Kriterien waren durch den Prototyp-Fold-Commit `8fa5778` bereits
erfüllt, es gab nichts zu härten:

- `templates/dashboard.html` zeigt eine durchgehende, `divide-x`-getrennte
  Leiste (Monatssoll, davon ruhend, im Rückstand, Neu/Ausgetreten) plus eine
  zweite Leiste für die Mitgliederzahlen nach Zustand — kein Kartenraster
  (kein `grid-cols` im Markup).
- Grep über `templates/` und `app/` findet keine Spur mehr von
  `dashboard-kachel`, `kachelDaten` oder der Template-Funktion `kachel`.
- `app/dashboard.go`: `dashboardDaten` trägt `service.Monatsuebersicht`
  unverändert weiter, keine eigene Berechnung im Handler — die
  Kennzahlenberechnung selbst ist bereits in `service/dashboard_test.go`
  unit-getestet (die richtige Seam laut `CLAUDE.md`).
- Zusätzlich per Wegwerf-`httptest` gegen den echten Handler geprüft
  (`app.New` + `httptest.NewServer`): zwei Mitglieder mit bekanntem Beitrag
  angelegt, `svc.Monatsuebersicht()` direkt aufgerufen und mit dem
  gerenderten HTML abgeglichen — Monatssoll, Gesamt-/Aktiv-/Neu-Zahlen,
  Rückstandsanzahl und Neueintritte stimmen exakt überein; kein
  `dashboard-kachel`-Rest im Markup. Datei danach wieder entfernt, wie in
  der Spec vorgesehen.
- `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei.
- Keine Code-Änderung war nötig — das Ticket schließt als reine
  Verifikation.
