# 04: Eintritt-Filter (Datumsbereich) an der Eintritt-Spalte

**What to build:** Das Zahnrad der Eintritt-Spalte bekommt ein Von-Bis-
Datumspaar, das die Mitgliederliste auf einen Eintrittszeitraum eingrenzt.
Von und Bis sind unabhängig voneinander optional nutzbar (nur Von, nur Bis,
oder beide). Beide Ränder zählen einschließend, wie an anderer Stelle im
Datenmodell üblich (siehe CONTEXT.md → Status).

`service.Suchfilter` bekommt zwei neue optionale Felder, `EintrittVon
*time.Time` und `EintrittBis *time.Time`. `nil` grenzt jeweils nicht ein.

**Blocked by:** 01 (Zahnrad-Grundgerüst)

**Status:** ready-for-human

- [x] `Suchfilter` hat neue Felder `EintrittVon`, `EintrittBis`
      (`*time.Time`, `nil` = unbegrenzt)
- [x] `Search` filtert korrekt: nur Von gesetzt, nur Bis gesetzt, beide
      gesetzt, beide `nil`
- [x] Randtage (Eintritt exakt gleich Von bzw. Bis) zählen einschließend mit
- [x] UI: Eintritt-Zahnrad zeigt Von/Bis-Datumsfelder, Zahnrad markiert
      aktiv, wenn ein Wert gesetzt ist
- [x] Neue Unit-Tests in `service/member_service_test.go` für alle
      Grenzfälle inklusive exaktem Randtag
- [x] `wails dev` und `wails build` laufen weiterhin auf Windows und Linux
      (`wails build` lokal auf Linux geprüft; Go-Code zusätzlich mit
      `GOOS=windows go build ./...` cross-kompiliert)

## Comments

Übersetzung der neuen Zahnrad-Beschriftungen (`filter_eintritt_von_aria`,
`filter_eintritt_bis_aria`) bleibt wie bei 02/03 vorerst Deutsch-only —
Englisch kommt gebündelt in Ticket 05.
