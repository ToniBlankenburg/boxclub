# 04: Export und Rechnung für fremde Zahler

**What to build:** Zahlt jemand anderes, folgen MoneyMoney-Export und Rechnung dem Zahler. Der Export nennt den Kontoinhaber als „Zahlungspflichtigen“ und nimmt dessen IBAN; weiter eine Zeile je Mitgliedschaft, die Mandatsreferenz bleibt die Mitglieds-ID, Beitrag und Anmeldegebühr bleiben unverändert (ADR-0013). Die Rechnung belegt den Empfängernamen vom Zahler der aktuellen Mitgliedschaft vor, die Anschrift kommt vom Mitglied; überschreibbar wie bisher; das Dokument bleibt am Mitglied (ADR-0007, ADR-0009).

**Blocked by:** 03 (Abweichender Kontoinhaber)

**Status:** ready-for-agent

- [ ] Export: Zwei Kinder derselben Eltern ergeben zwei Zeilen mit demselben Zahlungspflichtigen und derselben IBAN, aber je eigener Mandatsreferenz und eigenem Betrag
- [ ] Export: Ein Mitglied, das selbst zahlt, erscheint mit seinem aktuellen Namen (auch nach einer Namensänderung)
- [ ] Rechnung: Name vom Zahler, Anschrift vom Mitglied, beides überschreibbar; Dokument hängt am Mitglied
- [ ] Rückstand, Monatssoll und Dashboard bleiben unverändert
- [ ] Tests nach dem Muster von `service/moneymoney_test.go` und `service/rechnung_test.go`; `go test ./...` grün
