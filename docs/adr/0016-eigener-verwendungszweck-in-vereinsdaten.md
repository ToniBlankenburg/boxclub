# ADR-0016: Eigener Verwendungszweck in den Vereinsdaten statt Freitext je Lauf

**Status:** Accepted — erweitert [ADR-0013](0013-moneymoney-csv-export-fuer-lastschrifteinzug.md), löst es nicht ab
**Datum:** 2026-09-22

## Kontext

ADR-0013 legt den Verwendungszweck des MoneyMoney-Exports fest verdrahtet an:
"Vereinsname Beitrag MM/JJJJ", ergänzt um "Anmeldegebühr + ", wenn die Zeile
eine offene Anmeldegebühr enthält. In der Praxis gibt es Monate, in denen der
Verein einen anderen Text braucht — etwa einen Hinweis auf eine
Nachzahlung oder einen Korrekturlauf. Der reine Automatismus deckt das nicht,
und ohne App-Unterstützung bliebe nur wieder das Abtippen in MoneyMoney
selbst, das ADR-0013 gerade abgeschafft hat.

## Entscheidung

Die Vereinsdaten bekommen ein neues optionales Textfeld für den
Verwendungszweck. Leer bleibt das Verhalten wie in ADR-0013. Gesetzt ersetzt
es nur den "Vereinsname Beitrag MM/JJJJ"-Teil — das "Anmeldegebühr + "-Präfix
bleibt in jedem Fall automatisch pro Zeile, weil es Faktenstand der
jeweiligen Mitgliedschaft ist (ob diese Zeile eine offene Anmeldegebühr
mitführt), kein Stiltext, den der Verein frei wählt.

Das Feld ist an den Vereinsdaten dauerhaft gepflegt, nicht pro Exportlauf
einzeln eingegeben: ein Ort für den Wert, keine zweite Eingabe am
Export-Button und kein Merken eines "letzten Werts" — beides wären zwei
Wahrheiten für denselben Text.

Weil ein gesetzter Text sich nicht selbst nach Monat und Jahr fortschreibt,
bekommt das Vereinsdaten-Formular einen Hinweistext neben dem Feld, und der
Export-Button einen Warnhinweis, sobald das Feld aktuell gesetzt ist —
ersterer erklärt die grundsätzliche Falle, zweiterer fängt den akuten Fall
(vergessenes Zurücksetzen) im Moment ab, in dem er zählt.

## Konsequenzen

**Gut:**

- Deckt den Ausnahmefall ab, ohne den Regelfall zu verkomplizieren.
- Ein einziger Ort für den Wert, keine Widersprüche zwischen
  Vereinsdaten-Feld und Exportlauf.
- Die Anmeldegebühr-Kennzeichnung bleibt zuverlässig, weil sie nicht vom
  Freitext abhängt.

**Schlecht:**

- Das Feld pflegt sich nicht selbst nach; ein vergessenes Zurücksetzen lässt
  einen veralteten Text stehen, bis jemand es bemerkt — abgefedert, nicht
  verhindert, durch die zwei Warnhinweise.
- Für einen abweichenden Text muss der Verein erst in die Vereinsdaten
  wechseln, exportieren und danach wieder zurückwechseln — mehr Klicks als
  ein Eingabefeld direkt am Export.

## Alternativen

- **Eingabefeld direkt am Export-Button**, pro Lauf neu, nichts gespeichert.
  Verworfen: bei jedem Lauf müsste der Vereinsname+Monat-Teil erneut
  eingetippt werden, ohne Vorschlag als Ausgangspunkt.
- **Platzhalter-Vorlage** mit Monat/Jahr-Platzhaltern statt festem Text,
  damit sich der Text selbst nachführt. Verworfen für jetzt: mehr Aufwand
  (Platzhalter-Syntax, Parsing, Tippfehleranfälligkeit) für einen Fall, der
  selten genug ist, dass zwei Warnhinweise reichen.
- **Freitext auch für den Anmeldegebühr-Hinweis.** Verworfen: das ist
  Faktenstand einer einzelnen Zeile, kein Stiltext — ein globaler Text kann
  das nicht abbilden, ohne für Zeilen ohne Anmeldegebühr falsch zu werden.
