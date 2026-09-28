# 04: Eintritt-Filter (Datumsbereich) an der Eintritt-Spalte

**What to build:** Das Zahnrad der Eintritt-Spalte bekommt ein Von-Bis-
Datumspaar, das die Mitgliederliste auf einen Eintrittszeitraum eingrenzt.
Von und Bis sind unabhängig voneinander optional nutzbar (nur Von, nur Bis,
oder beide). Beide Ränder zählen einschließend, wie an anderer Stelle im
Datenmodell üblich (siehe CONTEXT.md → Status).

`service.Suchfilter` bekommt zwei neue optionale Felder, `EintrittVon
*time.Time` und `EintrittBis *time.Time`. `nil` grenzt jeweils nicht ein.

**Blocked by:** 01 (Zahnrad-Grundgerüst)

**Status:** ready-for-agent

- [ ] `Suchfilter` hat neue Felder `EintrittVon`, `EintrittBis`
      (`*time.Time`, `nil` = unbegrenzt)
- [ ] `Search` filtert korrekt: nur Von gesetzt, nur Bis gesetzt, beide
      gesetzt, beide `nil`
- [ ] Randtage (Eintritt exakt gleich Von bzw. Bis) zählen einschließend mit
- [ ] UI: Eintritt-Zahnrad zeigt Von/Bis-Datumsfelder, Zahnrad markiert
      aktiv, wenn ein Wert gesetzt ist
- [ ] Neue Unit-Tests in `service/member_service_test.go` für alle
      Grenzfälle inklusive exaktem Randtag
- [ ] `wails dev` und `wails build` laufen weiterhin auf Windows und Linux
