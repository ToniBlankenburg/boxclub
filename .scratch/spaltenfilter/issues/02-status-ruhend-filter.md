# 02: Status- und Ruhend-Filter an der Status-Spalte

**What to build:** Das Zahnrad der Status-Spalte bekommt zwei unabhängige
Filter: einen Status-Filter (alle / Neu / Aktiv / In Kündigungsfrist /
Ausgetreten) und einen Ruhend-Filter (alle / ruhend / laufend) — letzterer
eigenständig, weil Ruhend laut Datenmodell kein eigener Lebenszyklus-Status
ist, sondern eine Fahne daneben. Der Status-Filter übernimmt vollständig die
Funktion der alten Ehemalige-Checkbox: nur wer aktiv "Ausgetreten" wählt,
sieht ausgetretene Mitglieder; die Standardansicht (kein Statusfilter aktiv)
zeigt weiterhin keine Ausgetretenen.

`service.Suchfilter` verliert dafür das öffentliche Feld `AuchEhemalige` und
bekommt stattdessen `Status Statusfilter` (neuer Enum-Typ) sowie
`Ruhend Ruhendfilter` (neuer Enum-Typ, analog zu `Rueckstandsfilter`).
`Search` entscheidet intern anhand von `Status`, ob ausgetretene
Mitgliedschaften überhaupt geladen werden (nur bei `StatusfilterAusgetreten`)
— der Aufrufer setzt das nicht mehr separat.

**Blocked by:** 01 (Zahnrad-Grundgerüst)

**Status:** ready-for-agent

- [ ] `Suchfilter.AuchEhemalige` entfernt, ersetzt durch `Status Statusfilter`
- [ ] Neuer Typ `Ruhendfilter` (alle/ruhend/laufend) mit `trifft`-Methode,
      Feld `Ruhend` in `Suchfilter`
- [ ] `Suchfilter{}` (Nullwert) liefert weiterhin exakt dieselbe Menge wie
      `List()` — keine Ausgetretenen, keine sonstige Einschränkung
- [ ] `Status = Ausgetreten` lädt ausgetretene Mitgliedschaften erst dazu und
      liefert ausschließlich diese
- [ ] `Status ∈ {Neu, Aktiv, InKündigungsfrist}` grenzt korrekt ein, ohne
      zusätzlich Ausgetretene zu laden
- [ ] `Ruhend`-Filter grenzt unabhängig vom Statusfilter korrekt ein
      (Kombination z. B. Status=Aktiv + Ruhend=ruhend liefert nur aktive UND
      ruhende Mitglieder)
- [ ] Neue Unit-Tests in `service/member_service_test.go` für Status, Ruhend
      und deren Kombination
- [ ] UI: Status-Zahnrad zeigt beide Auswahlfelder, Zahnrad markiert aktiv,
      wenn einer der beiden Filter vom Default abweicht
- [ ] Alte Ehemalige-Checkbox samt zugehörigem `ehemalige`-Query-Parameter
      ist vollständig entfernt
- [ ] Filter aus Ticket 01 (Rückstand, Frequenz, Termin, Geschlecht) bleiben
      unverändert funktionsfähig
- [ ] `wails dev` und `wails build` laufen weiterhin auf Windows und Linux
