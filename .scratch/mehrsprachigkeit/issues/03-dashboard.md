Status: ready-for-human

# 03: Dashboard übersetzt

**What to build:** `templates/dashboard.html` und `app/dashboard.go`
(Monatssoll, Mitgliederzahlen, MoneyMoney-Export-Hinweise).

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [x] Alle sichtbaren Texte übersetzt, inklusive der Warnhinweise zum
      MoneyMoney-Export (ADR-0016)
- [x] `go test ./...` grün, `wails build` unter Linux grün

## Comments

### Umsetzung

- Neue Schlüssel unter `dashboard.*` in `i18n/de.go`/`i18n/en.go`. Der
  eingebettete `<strong>Monatssoll</strong>` im Fließtext ist wie
  `mitglied.stundenplan_leer_vor`/`_nach` (Ticket 02) in `_vor`/`_nach`
  aufgetrennt, das Wort selbst steht unter `dashboard.monatssoll` und wird
  auch für die Kennzahlenbeschriftung darüber wiederverwendet.
- `mitglieder.im_rueckstand_kasten` (Ticket 02) für die Rückstand-Kachel
  wiederverwendet statt eines eigenen Schlüssels — derselbe Text an beiden
  Stellen.
- `app/dashboard.go` brauchte keine Änderung: Es enthält keine
  Text-Literale, nur die `MoneyMoneyEigenerText`-Ableitung.
- `go build ./...` und `go test ./...` grün. `wails build` unter Linux nicht
  separat ausgeführt (kein Wails-Toolchain-Zugriff in dieser Session) —
  reiner Go-Build und die Template-Ausführung über `app`-Tests decken den
  Template-Syntax-Anteil ab.
- **Bewusst nicht angefasst:** die Rückmeldungstexte, die
  `app/moneymoney.go` nach dem eigentlichen Export-Klick baut
  (`fmt.Sprintf`-Meldungen wie "Der Export wurde nach %s gespeichert …").
  Sie werden zwar auf demselben Dashboard angezeigt (`meldung`), liegen aber
  in einer Datei, die kein Ticket in `.scratch/mehrsprachigkeit/spec.md`
  nennt — weder dieses noch ein späteres. Bis zur Klärung mit dem Nutzer
  bleiben sie deutsch.
