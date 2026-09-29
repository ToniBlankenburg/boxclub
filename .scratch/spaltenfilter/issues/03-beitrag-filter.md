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

**Status:** done

- [x] `Suchfilter` hat neue Felder `BeitragVonCents`, `BeitragBisCents`
      (`*int64`, `nil` = unbegrenzt)
- [x] `Search` filtert korrekt: nur Von gesetzt, nur Bis gesetzt, beide
      gesetzt, beide `nil`
- [x] Grenzfall 0 € ist als gesetzter Wert (nicht `nil`) korrekt filterbar
      und schließt Mitglieder mit exakt 0 € Beitrag korrekt ein/aus
- [x] Regler-Grenzen werden serverseitig aus den tatsächlich vorkommenden
      Beitragswerten berechnet und ans Template übergeben
- [x] UI: Beitrag-Zahnrad zeigt Doppel-Schieberegler, Zahnrad markiert aktiv,
      wenn ein Wert vom Nullwert abweicht
- [x] Neue Unit-Tests in `service/search_test.go` für alle Grenzfälle (dort
      liegen die übrigen Suchfilter-Tests desselben Bestands, siehe
      `suchbestandAnlegen`)
- [x] `wails dev` und `wails build` laufen weiterhin auf Windows und Linux

## Comments

Die Regler-Grenzen kommen aus einer neuen `MemberService.BeitragBereich()` —
sie werden nicht nur beim ersten Laden der Seite geholt, sondern bei jedem
Aufruf von `listeDatenLesen` (also auch bei jedem Such-/Filter-Fragment):
sonst bräche der Regler nach der ersten anderen Filteränderung auf
Grenzen 0/0 ein, weil das Ergebnis-Fragment (in dem der Regler jetzt steckt,
ADR-0020) sonst keine frischen Grenzen bekäme.

Als "Doppel-Schieberegler" dienen zwei unabhängige `<input type="range">`
(Von/Bis) statt eines echten Zwei-Griff-Reglers — letzterer bräuchte eigenes
JS/CSS, das dieser Vanille-Stack sonst nirgends hat.

`frontend/src/main.js` (`sortierfallbackPruefen`) musste um die beiden neuen
Parameter ergänzt werden, sonst ginge der Beitragsfilter verloren, sobald die
Beitrag-Spalte während aktiver Beitragssortierung ausgeblendet wird.
