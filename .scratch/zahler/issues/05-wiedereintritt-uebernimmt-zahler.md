# 05: Wiedereintritt übernimmt den Zahler

**What to build:** Ein Wiedereintritt startet die neue Mitgliedschaft mit dem Zahler der Vorgängerin als Vorschlag. Der Verweis gehört der neuen Mitgliedschaft und ist änderbar; die alte behält ihren eigenen Zahler. Anders als Anmeldedatum, Anmeldegebühr und Trainingstermine wird der Zahler also übernommen.

**Blocked by:** 03 (Abweichender Kontoinhaber)

**Status:** done

- [x] Nach einem Wiedereintritt zeigt die neue Mitgliedschaft auf denselben Zahler wie die alte
- [x] Ändert der Admin danach den Kontoinhaber der neuen Mitgliedschaft, bleibt die alte unverändert
- [x] Hatte die Vorgängerin einen abweichenden Kontoinhaber, bleibt er auch bei einem nicht mehr minderjährigen Mitglied vorgeschlagen (siehe Ticket 06)
- [x] Tests nach dem Muster von `service/lebenszyklus_test.go`; `go test ./...` grün
