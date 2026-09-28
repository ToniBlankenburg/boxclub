# 03: Beitrag-Filter (Schieberegler) an der Beitrag-Spalte

**What to build:** Das Zahnrad der Beitrag-Spalte bekommt einen
Doppel-Schieberegler (Minimum und Maximum), der die Mitgliederliste auf einen
Beitragsbereich eingrenzt. Die Grenzen des Reglers berechnen sich dynamisch
aus den tatsächlich vorkommenden Beitragswerten der aktuell geladenen Liste,
nicht aus einem festen Rahmen. Ein Beitrag von 0 € ist dabei ein normaler,
wählbarer Wert — kein "kein Filter"-Sonderfall (siehe CONTEXT.md → Beitrag).

`service.Suchfilter` bekommt zwei neue optionale Felder,
`BeitragVonCents *int64` und `BeitragBisCents *int64` — als Pointer, nicht als
Sentinel-Wert, weil 0 ein gültiger Beitrag ist. `nil` grenzt jeweils nicht
ein. Die Euro-Eingabe des Reglers wird über das bestehende
`service.BeitragAusEuro` nach Cents umgerechnet.

**Blocked by:** 01 (Zahnrad-Grundgerüst)

**Status:** ready-for-agent

- [ ] `Suchfilter` hat neue Felder `BeitragVonCents`, `BeitragBisCents`
      (`*int64`, `nil` = unbegrenzt)
- [ ] `Search` filtert korrekt: nur Von gesetzt, nur Bis gesetzt, beide
      gesetzt, beide `nil`
- [ ] Grenzfall 0 € ist als gesetzter Wert (nicht `nil`) korrekt filterbar
      und schließt Mitglieder mit exakt 0 € Beitrag korrekt ein/aus
- [ ] Regler-Grenzen werden serverseitig aus den tatsächlich vorkommenden
      Beitragswerten berechnet und ans Template übergeben
- [ ] UI: Beitrag-Zahnrad zeigt Doppel-Schieberegler, Zahnrad markiert aktiv,
      wenn ein Wert vom Nullwert abweicht
- [ ] Neue Unit-Tests in `service/member_service_test.go` für alle
      Grenzfälle
- [ ] `wails dev` und `wails build` laufen weiterhin auf Windows und Linux
