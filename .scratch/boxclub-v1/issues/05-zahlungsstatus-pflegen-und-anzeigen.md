Status: ready-for-human

# 05: Zahlungsstatus pflegen & anzeigen

**What to build:** Jede Zeile der Mitgliederliste zeigt einen sichtbaren **Zahlungsstatus** — grün wenn bezahlt, rot wenn nicht — abgeleitet aus `bezahlt_bis` und `heute`. Eine Zeilen-Aktion lässt den Nutzer `bezahlt_bis` nach einem Blick in MoneyMoney manuell setzen.

**Blocked by:** 03 (Mitgliederliste anzeigen)

## Acceptance Criteria

- [x] Jede Mitgliederzeile zeigt einen visuell klaren Status: **grün** wenn `bezahlt_bis >= heute`, **rot** sonst
- [x] Wenn `bezahlt_bis` nie gesetzt wurde, wird ein neutraler "nicht gesetzt"-Zustand angezeigt (nicht fälschlich grün)
- [x] Eine Aktion pro Zeile öffnet einen Datums-Picker zum Setzen von `bezahlt_bis`
- [x] Nach dem Speichern aktualisiert sich der Zeilenstatus sofort per htmx-Fragment
- [x] `MemberService.SetBezahltBis(id, datum)` persistiert den Wert; wiederholter Aufruf mit demselben Datum ist idempotent
- [x] Tests am Seam für die Status-Ableitung: `bezahlt_bis = gestern` → nicht bezahlt, `= heute` → bezahlt, `= morgen` → bezahlt, `NULL` → nicht gesetzt
- [x] Test am Seam: Setzen und Auslesen von `bezahlt_bis` verändert keine anderen Felder

## Notes

Zweistufig, keine gelbe Vorwarnstufe (siehe spec.md und Grill-Runde 3, Q4).

## Comments

### Umsetzung (Code-Review eingearbeitet)

- Seam: `Zahlungsstatus` mit drei Werten und `MemberService.SetBezahltBis(id, *time.Time)`. "nicht gesetzt" ist keine dritte Stufe, sondern das Fehlen einer Angabe — `CONTEXT.md`, `CLAUDE.md` und `spec.md` sind entsprechend präzisiert, sie sagten bis dahin nur "zweistufig".
- Der Status wird über die **ISO-Textform** verglichen, nicht über die Zeitpunkte: aus der Datenbank gelesene Daten liegen in UTC, "heute" kommt aus der lokalen Uhr. Ein Zeitpunktvergleich läge bei negativem Zonenversatz (z. B. `UTC-10`) einen Tag daneben — genau am Grenzfall `bezahlt_bis = heute`.
- `MemberService.Eintrag(id)` kam neu dazu: die Zeile, die nach dem Speichern allein zurückkommt, ist dieselbe, die `List` liefert. `List` und `Eintrag` teilen sich dafür eine Abfrage.
- `SetBezahltBis(id, nil)` setzt die Angabe zurück; im Formular tut das ein leeres Feld. Über das Ticket hinaus, aber die einzige ehrliche Antwort auf einen geleerten Datums-Picker — und die einzige Korrektur für ein versehentlich gesetztes Datum. Die Alternative wäre eine Fehlermeldung gewesen.
- Der Zeilen-Austausch braucht zwei Nebenwege: `GET /api/mitglied/{id}/zeile` für "Abbrechen" und `HX-Retarget` auf `#inhalt`, wenn es die Zeile nicht mehr gibt — sonst landete die ganze Liste in einem `<tr>`.
- Die Schaltfläche steht in der Spalte "Bezahlt bis", also bei dem Wert, den sie ändert, und stoppt ihr Click-/Keyup-Ereignis; die restliche Zeile öffnet weiter das Bearbeitungsformular.
- Bekannte Kante für Ticket 06/07: `Eintrag` findet nur Mitglieder mit laufender Mitgliedschaft. Sobald die Liste auch Ausgetretene zeigt, muss die geteilte Abfrage auf die *letzte* statt die *laufende* Mitgliedschaft joinen — heute ist der Fall nicht erreichbar, weil ausgetretene Mitglieder gar nicht in der Liste stehen.
- Verifiziert: `go test ./...` grün, `wails build` unter Linux grün, Handler-Pfade per Wegwerf-Smoke-Test gegen `httptest` durchgespielt (Liste, Formular, Speichern, Zurücksetzen, ungültiges Datum, unbekannte ID). `wails dev` und Windows wurden **nicht** gegengeprüft.
