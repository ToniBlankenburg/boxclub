# 01: Zahnrad-Grundgerüst für Spaltenfilter + ADR

**What to build:** Das alte Filterpanel (Trichter-Button) und das alte
"Spalten"-Menü der Mitgliederliste entfallen. Jede heute schon ausblendbare
Spaltenüberschrift (alle außer Nr. und Name) bekommt stattdessen ein eigenes
Zahnrad-Icon, das ein kleines Menü direkt unter dieser Spalte öffnet. Die
bestehenden Filter aus dem alten Panel (Rückstand, Frequenz, Termin,
Geschlecht) ziehen unverändert in ihre passende Spalte um — Rückstand in die
Rückstand-Spalte, Frequenz und Termin gemeinsam in die Training-Spalte,
Geschlecht in die Name-Spalte (da Geschlecht keine eigene Spalte hat). Die
Aktion "Spalte ausblenden" zieht aus dem alten Spalten-Menü ebenfalls ins
jeweilige Zahnrad. Ein neues Fallback-Element am Ende der Kopfzeile listet
ausgeblendete Spalten zum Wiedereinblenden auf. Ein eigener Reset-Icon-Button
setzt alle aktiven Filter auf einmal zurück. Jedes Zahnrad-Icon zeigt visuell
an, ob auf seiner Spalte gerade gefiltert wird. Nur ein Zahnrad-Menü ist
gleichzeitig offen.

Vor der Umsetzung ist eine ADR zu schreiben, die ADR-0015 (und den
Filterpanel-Teil davor) ablöst — Vorgabe aus der Spec (`.scratch/spaltenfilter/spec.md`).

Dieses Ticket ändert `service.Suchfilter` nicht (keine neuen Filterdimensionen,
nur Umzug bestehender Filter in die neue UI). Die Ehemalige-Checkbox aus dem
alten Panel entfällt in diesem Ticket ersatzlos — ihr Ersatz (Status-Filter)
kommt in Ticket 02, das direkt auf diesem aufbaut.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] Neue ADR geschrieben, referenziert ADR-0015 als abgelöst
- [ ] Zahnrad-Icon an jeder ausblendbaren Spalte (Status, Anschrift, Training,
      Beitrag, Rückstand, Eintritt); Nr. und Name bleiben ohne Zahnrad
- [ ] Rückstand-Filter funktioniert identisch zu vorher, jetzt aus dem
      Rückstand-Zahnrad heraus
- [ ] Frequenz- und Termin-Filter funktionieren identisch zu vorher, jetzt
      gemeinsam aus dem Training-Zahnrad heraus
- [ ] Geschlecht-Filter funktioniert identisch zu vorher, jetzt aus dem
      Name-Zahnrad heraus
- [ ] Altes Filterpanel (Trichter-Button) und altes Spalten-Menü sind entfernt
- [ ] Spalte ausblenden funktioniert wie bisher über das jeweilige Zahnrad
      (Sichtbarkeit weiterhin per localStorage persistiert)
- [ ] Fallback-Element am Ende der Kopfzeile zeigt ausgeblendete Spalten und
      erlaubt Wiedereinblenden
- [ ] Eigener Reset-Icon-Button setzt alle aktiven Filter zurück
- [ ] Zahnrad zeigt visuell an, ob auf seiner Spalte gerade gefiltert wird
- [ ] Nur ein Zahnrad-Menü gleichzeitig offen
- [ ] Bestehende Sortierung per Klick auf den Spaltentitel bleibt unverändert
      nutzbar
- [ ] Bestehende Service-Tests (`service/member_service_test.go`) bleiben grün
      (keine Änderung an `Suchfilter` in diesem Ticket)
- [ ] `wails dev` und `wails build` laufen weiterhin auf Windows und Linux
