Status: ready-for-agent

# 17: Kündigung erfassen und Mitglied ruhend schalten

**What to build:** Der Vereinsadmin erfasst eine **Kündigung** mit zwei Daten: wann gekündigt wurde und wann der Austritt wirksam wird. Dazwischen liegt die **Kündigungsfrist**, in der das Mitglied weiter trainiert und weiter zahlt. Unabhängig davon kann er ein Mitglied **ruhend** schalten, wenn es vorübergehend nicht trainiert (Verletzung, Ausland) — es bleibt Mitglied, aber für diese Zeit wird kein Beitrag eingezogen.

**Blocked by:** 13 (Beitrag individuell) — beide Felder hängen an der Mitgliedschaft, die dort umstrukturiert wird

## Acceptance Criteria

- [ ] An der Mitgliedschaft: Kündigungsdatum und ein Ruhend-Kennzeichen
- [ ] Eine Kündigung lässt sich mit Kündigungsdatum **und** Austrittsdatum erfassen
- [ ] Das **Austrittsdatum wird nicht aus dem Kündigungsdatum berechnet**. Fristen haben Sonderfälle (Kulanz, Aufhebungsvertrag, Quartalsende); ein errechnetes Datum, das man überschreiben muss, ist lästiger als ein leeres Feld
- [ ] Ein Austrittsdatum **darf in der Zukunft liegen** — das ist der Normalfall bei laufender Kündigungsfrist
- [ ] Ein Kündigungsdatum ohne Austrittsdatum ist erlaubt (Kündigung liegt vor, Termin noch offen)
- [ ] Ein Mitglied lässt sich ruhend schalten und wieder aktiv setzen
- [ ] Ruhend und Rückstand sind **unabhängig**: für ein ruhendes Mitglied wird keine Lastschrift losgeschickt, es kann also kein *neuer* Rückstand entstehen — ein bestehender bleibt aber stehen
- [ ] Tests am Service-Seam: Kündigung mit und ohne Austrittstermin, Austritt in der Zukunft, ruhend setzen und zurücknehmen, ruhend mit bestehendem Rückstand
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

Dieses Ticket legt nur die **Felder und ihre Pflege** an. Dass daraus der Status *In Kündigungsfrist* abgelesen wird und sich die Bedeutung von "aktiv" ändert, ist Ticket 18 — sonst wäre die Abhängigkeit zirkulär.

Entwicklungs-Datenbank vor dem ersten Start löschen.
