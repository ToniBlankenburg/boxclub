# 03: Abweichender Kontoinhaber

**What to build:** Im Mitglieds-Formular steht neben der IBAN-Zeile das Feld „Abweichender Kontoinhaber“. Es ist nur aktiv, solange das Mitglied ein Kind ist (Ticket 02); sonst deaktiviert und leer. Ist es gefüllt, zahlt nicht das Mitglied, sondern ein anderer Zahler (Name + IBAN). Das Wort „Zahler“ erscheint in der UI nicht. Zusammenführen: gleiche normalisierte IBAN heißt gleicher Zahler (Geschwister teilen sich eine Zeile); bei unterschiedlichen Namen gewinnt der zuletzt gespeicherte und gilt für alle; eine leere IBAN bildet nie einen gemeinsamen Zahler. Ein Zahler ohne Mitgliedschaft wird automatisch gelöscht. Leert der Admin den Kontoinhaber, zahlt das Mitglied wieder selbst.

**Blocked by:** 01 (Zahler „selbst“ als Grundlage), 02 (Kind-Status)

**Status:** ready-for-agent

- [ ] Beschriftung „Abweichender Kontoinhaber“ in beiden Sprachkatalogen (ADR-0017); Fehlermeldungen als Schlüssel (ADR-0018)
- [ ] Bei Erwachsenen ist das Feld deaktiviert; der Service lehnt einen abweichenden Kontoinhaber für Nicht-Kinder ab
- [ ] Zwei Kinder mit gleicher IBAN (andere Schreibweise) teilen sich einen Zahler: eine Namensänderung bei dem einen zeigt sich bei dem anderen
- [ ] Gleiche IBAN mit unterschiedlichem Namen: der zuletzt gespeicherte gewinnt
- [ ] Zwei Mitgliedschaften ohne IBAN teilen sich keinen Zahler
- [ ] Zahler ohne Mitgliedschaft verschwindet automatisch (beobachtbar über die Service-API, nicht per SQL)
- [ ] Kontoinhaber leeren lässt das Mitglied wieder selbst zahlen
- [ ] Tests über die Service-API; `go test ./...` grün; `wails build` auf Linux
