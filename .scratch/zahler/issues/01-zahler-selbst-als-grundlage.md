# 01: Zahler „selbst“ als Grundlage

**What to build:** Die IBAN wandert von der Person an eine neue technische Entität *Zahler*. Jede Mitgliedschaft zeigt auf genau einen Zahler (`zahler_id NOT NULL`); `mitglied.iban` entfällt. Beim Anlegen und Ändern eines Mitglieds entsteht bzw. gilt ein Zahler „selbst“ (mit `mitglied_id`, Name kommt beim Lesen aus der Person). Die IBAN wird beim Schreiben immer normalisiert (Leerzeichen weg, Großbuchstaben) — bewusste Abweichung von „reiner Text“ aus ADR-0006, nur für diese Spalte. Formular, Excel-Import, MoneyMoney-Export und Rechnungsvorbelegung lesen und schreiben darüber. Für den Admin ändert sich nichts Sichtbares: eine IBAN-Zeile wie bisher. Die Entwicklungsdatenbank wird gelöscht, es gibt keine Migration.

**Blocked by:** None (can start immediately)

**Status:** done

- [ ] Tabelle `zahler(id, name, iban, mitglied_id NULL)` und `mitgliedschaft.zahler_id NOT NULL` im Schema; `mitglied.iban` entfällt
- [ ] Anlegen und Ändern eines Mitglieds über `MemberService` legt bzw. aktualisiert den Zahler „selbst“; die IBAN wird normalisiert gespeichert
- [ ] Ein Mitglied ohne IBAN bekommt trotzdem einen Zahler; zwei Mitglieder ohne IBAN teilen sich keinen
- [ ] Der Excel-Importer legt je Zeile Mitglied, Mitgliedschaft und Zahler „selbst“ an; seine Schnittstelle (Zeilen + Fehlerbericht) bleibt gleich
- [ ] Der MoneyMoney-Export liefert Zahlungspflichtigen (Name der Person) und IBAN über den Zahler; Mandatsreferenz bleibt die Mitglieds-ID
- [ ] Die Rechnungsvorbelegung liefert dasselbe wie bisher
- [ ] Tests nach dem Muster von `service/member_service_test.go` (echte SQLite, nur über die Service-API); `go test ./...` grün
- [ ] `wails dev` und `wails build` laufen auf Linux
