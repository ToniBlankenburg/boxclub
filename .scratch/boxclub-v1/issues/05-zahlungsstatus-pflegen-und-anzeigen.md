Status: ready-for-agent

# 05: Zahlungsstatus pflegen & anzeigen

**What to build:** Jede Zeile der Mitgliederliste zeigt einen sichtbaren **Zahlungsstatus** — grün wenn bezahlt, rot wenn nicht — abgeleitet aus `bezahlt_bis` und `heute`. Eine Zeilen-Aktion lässt den Nutzer `bezahlt_bis` nach einem Blick in MoneyMoney manuell setzen.

**Blocked by:** 03 (Mitgliederliste anzeigen)

## Acceptance Criteria

- [ ] Jede Mitgliederzeile zeigt einen visuell klaren Status: **grün** wenn `bezahlt_bis >= heute`, **rot** sonst
- [ ] Wenn `bezahlt_bis` nie gesetzt wurde, wird ein neutraler "nicht gesetzt"-Zustand angezeigt (nicht fälschlich grün)
- [ ] Eine Aktion pro Zeile öffnet einen Datums-Picker zum Setzen von `bezahlt_bis`
- [ ] Nach dem Speichern aktualisiert sich der Zeilenstatus sofort per htmx-Fragment
- [ ] `MemberService.SetBezahltBis(id, datum)` persistiert den Wert; wiederholter Aufruf mit demselben Datum ist idempotent
- [ ] Tests am Seam für die Status-Ableitung: `bezahlt_bis = gestern` → nicht bezahlt, `= heute` → bezahlt, `= morgen` → bezahlt, `NULL` → nicht gesetzt
- [ ] Test am Seam: Setzen und Auslesen von `bezahlt_bis` verändert keine anderen Felder

## Notes

Zweistufig, keine gelbe Vorwarnstufe (siehe spec.md und Grill-Runde 3, Q4).
