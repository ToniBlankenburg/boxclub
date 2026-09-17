# 02: Dashboard: Kennzahlenleiste

**What to build:** Das Dashboard zeigt Monatssoll und Mitgliederzahlen als
durchgehende horizontale Kennzahlenleiste statt des bisherigen
Kartenrasters.

**Blocked by:** 01 (Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite)

**Status:** ready-for-agent

- [ ] Dashboard zeigt seine Kennzahlen als eine zusammenhängende Leiste,
      kein Kartenraster mehr
- [ ] Totes Kartenraster-Teil-Template ("dashboard-kachel"), Go-Typ
      `kachelDaten` und Template-Funktion `kachel` sind entfernt — keine
      toten Reste
- [ ] Zahlen bleiben korrekt: Monatssoll und Mitgliederzahlen stimmen
      weiterhin mit den bisherigen Werten überein (keine
      Berechnungsänderung, nur Darstellung)
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei
- [ ] Manueller Smoke-Test in `wails dev`: Dashboard öffnen, Werte mit
      einem bekannten Datensatz gegenprüfen
- [ ] Bereits vorhandene Umsetzung (Fold-Commit `8fa5778`) ist gegen diese
      Kriterien geprüft, nicht neu gebaut

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Etabliert das
Kennzahlenleisten-Muster, das Ticket 05 (Excel-Import-Bericht) übernimmt.
