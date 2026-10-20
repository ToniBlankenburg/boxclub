# ADR-0021: Zahler als technische Entität

**Status:** Accepted — weicht an einer Stelle von [ADR-0006](0006-rueckstand-statt-bezahlt-bis.md) ab (IBAN-Normalisierung) und baut auf [ADR-0013](0013-moneymoney-csv-export-fuer-lastschrifteinzug.md) auf.
**Datum:** 2026-10-20

## Kontext

Die App ging bisher davon aus, dass das Mitglied selbst zahlt: Name und IBAN
der Person waren zugleich Name und Konto der Lastschrift (MoneyMoney-Export,
ADR-0013) und die Vorbelegung des Rechnungsempfängers. Bei Kindern stimmt das
nicht — die Eltern zahlen. Der Admin konnte das nur abbilden, indem er die
Daten der Eltern in die Felder des Kindes tippte (dann war der Name des Kindes
falsch) oder sie bei jedem Kind erneut erfasste (dann stand dasselbe Konto
mehrfach in der Datenbank, und eine neue IBAN der Eltern erreichte nur ein Kind).

## Entscheidung

Eine Tabelle `zahler(id, name, iban, mitglied_id NULL)` trennt „wer trainiert"
von „wer zahlt". `mitgliedschaft.zahler_id` ist `NOT NULL`: jede
Mitgliedschaft hat genau einen Zahler.

- **Entität statt Felder an der Person.** Geschwister mit demselben Konto
  teilen sich eine Zeile, eine neue IBAN gilt für alle. Name und IBAN am
  Mitglied hätten das nicht geleistet.
- **Selbstzahler.** Ist `mitglied_id` gesetzt, zahlt das Mitglied selbst; der
  Name kommt beim Lesen aus der Person und folgt einer Umbenennung. Für
  Erwachsene ändert sich in der Oberfläche nichts.
- **Zusammenführen über die normalisierte IBAN.** Beim Speichern wird ein
  Zahler mit gleicher IBAN wiederverwendet; der zuletzt gespeicherte Name gilt
  für alle. Eine leere IBAN verbindet nie: jede Mitgliedschaft ohne IBAN
  bekommt eine eigene Zeile.
- **Aufräumen ohne Verwaltung.** Ein Zahler ohne Mitgliedschaft wird
  gelöscht. Es gibt keine Zahler-Ansicht, -Liste oder -Pflege; die Oberfläche
  sagt *Kontoinhaber* und *IBAN*, nie *Zahler*.
- **Kind aus dem Geburtsdatum.** Ein Kontoinhaber ist nur bei einem Kind
  erlaubt. Der Status wird bei jedem Lesen abgeleitet und nie gespeichert; nur
  ohne Geburtsdatum zählt das Handkennzeichen `mitglied.kind_manuell`.
  Wird das Mitglied volljährig, bleibt der Zahler wirksam und das Formular
  zeigt einen Hinweis, bis der Admin das Feld leert.
- **Wiedereintritt.** Die neue Mitgliedschaft übernimmt den Zahler der
  vorherigen als Vorschlag; eine Änderung lässt die alte unberührt, weil sie
  festhält, wer damals gezahlt hat.
- **Abweichung von ADR-0006.** `zahler.iban` wird beim Schreiben immer
  normalisiert (Leerzeichen entfernt, Großbuchstaben). Nur so sind
  `DE89 3704 0044 0532 0130 00` und `de89370400440532013000` dasselbe Konto.
  Das gilt für diese eine Spalte; Validierung und Formatprüfung bleiben
  ausgeschlossen.
- **`mitglied.iban` entfällt.** Die IBAN lebt ausschließlich am Zahler; es gibt
  nur eine IBAN-Zeile im Formular.
- Anschrift, Rückstand und Dokumente bleiben am Mitglied. Der Zahler hat
  bewusst nur Name und IBAN.

## Konsequenzen

**Gut:**

- Der MoneyMoney-Export nennt als „Zahlungspflichtiger" den Kontoinhaber und
  nimmt dessen IBAN; weiter eine Zeile je Mitgliedschaft, die Mandatsreferenz
  bleibt die Mitglieds-ID (ADR-0013). Eine Rechnung füllt den Empfängernamen
  mit dem Kontoinhaber vor, die Anschrift kommt vom Mitglied, beides bleibt
  überschreibbar.
- Der Excel-Importer behält seine Schnittstelle: er legt je Zeile Mitglied,
  Mitgliedschaft und Zahler „selbst" an und braucht keine Kind-Spalte.
- Rückstandsliste, Dashboard, Monatssoll, Suche und Serienmail bleiben
  unverändert.

**Schlecht:**

- **Keine Datenmigration.** Der Schnitt entfernt `mitglied.iban` und führt
  `mitgliedschaft.zahler_id` als Pflichtspalte ein; das ist keine additive
  Spalte mit eindeutigem Default im Sinne von
  [ADR-0019](0019-spaltenergaenzung-fuer-bestehende-datenbanken.md). Die
  Entwicklungsdatenbank wird gelöscht.
- **Ein Name pro Konto.** Wer dieselbe IBAN mit zwei Namen einträgt, überschreibt
  den älteren still. Bewusst: dasselbe Konto hat einen Namen.
- Das Zusammenführen ist unsichtbar. Der Admin sieht nicht, welche Mitglieder
  sich einen Zahler teilen, nur an den Folgen einer Änderung.

## Alternativen

- **Name und IBAN des Kontoinhabers als Felder an der Person.** Weniger Schema,
  aber Geschwister führen das Konto doppelt, und eine neue IBAN muss bei jedem
  Kind nachgezogen werden.
- **Zahler als eigene Ansicht mit Verwaltung.** Mehr Oberfläche für ein Thema,
  das bei 50–200 Mitgliedern nur eine Handvoll Kinder betrifft; Aufräumen und
  Zusammenführen sind automatisch genug.
- **Mehrere Zahler je Mitgliedschaft** (geteilte Beiträge). Nicht gebraucht;
  jede Mitgliedschaft hat genau einen.
- **Kind-Status speichern.** Wäre am Geburtstag veraltet und stillschweigend
  falsch; abgeleitet kann er nicht vergessen werden.
- **Beim 18. Geburtstag automatisch auf Selbstzahler umstellen.** Würde die
  Lastschrift ohne Zutun vom falschen Konto laufen lassen; der Hinweis lässt
  den Admin entscheiden.
