# ADR-0005: Der Beitrag ist individuell vereinbart, es gibt keine Beitragsklassen

**Status:** Accepted
**Datum:** 2026-09-11

## Kontext

Spec und Glossar gingen von **zwei frequenzbasierten Beitragsklassen** aus
(`Erwachsen 1×/Woche` = 60 €, `Erwachsen 2×/Woche` = 80 €). [CONTEXT.md](../../CONTEXT.md)
formulierte das als Regel: *"Die Staffelung ergibt sich bei diesem Verein
ausschließlich aus der Trainingsfrequenz."* Darauf bauen das Schema
(`beitragsklasse`-Tabelle, `mitglied.beitragsklasse_id NOT NULL`), Story 4
(Filter nach Klasse), Story 13 (Klassenpreise pflegen) und Ticket 08
(Beitragsklassen-Ansicht).

Die bestehende Excel-Tabelle des Vereins widerlegt die Regel
(siehe [Bestandsaufnahme](../../.scratch/boxclub-v1/excel-vorlage.md)). Sie führt
**zwei unabhängige Spalten**:

- `Beitrag` mit 18 gepflegten Werten: 0, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75,
  80, 85, 100, 110, 115, 135, 144 (€/Monat)
- `1x 2x Woche` mit 1×, 2×, 3× Woche

Betrag und Frequenz variieren also getrennt: zwei Mitglieder mit derselben
Frequenz können unterschiedlich viel zahlen. Der Beitrag ist das Ergebnis einer
individuellen Absprache (Altbestand, Härtefall, Familie, 0 € für Trainer),
nicht die Zuordnung zu einer Stufe.

## Entscheidung

Die `beitragsklasse`-Tabelle entfällt. Der Beitrag wird zu einem **Feld an der
Mitgliedschaft** (`mitgliedschaft.beitrag_monatlich_cents`, in Cent wie zuvor der
Klassenpreis) — nicht am Mitglied: er ist Teil der Vereinbarung, die mit dem
Eintritt zustande kam, und ein Wiedereintritt bekommt deshalb seinen eigenen,
während der alte an der alten Mitgliedschaft stehen bleibt. Die Trainingsfrequenz
wird eine eigene, davon unabhängige Eigenschaft.

Der entscheidende Grund gegen die naheliegende Kompromisslösung — Tabelle
behalten, mit 18 nach ihrem Betrag benannten Zeilen seeden — ist **Story 13**:
"Preiserhöhungen abbildbar machen, indem man den Preis einer Klasse ändert".
Bei individuell verhandelten Beiträgen wäre das falsches Verhalten. Eine
Preisänderung an der Klasse "60 €" würde stillschweigend den Beitrag *aller*
Mitglieder darin ändern, auch derjenigen, deren 60 € eine Zusage sind. Eine
Klasse, deren Name ihr Preis ist, ist zudem eine Nachschlagetabelle ohne
eigenen Inhalt.

## Konsequenzen

**Gut:**

- Das Modell kann die Realität des Vereins überhaupt abbilden. Mit zwei Klassen
  wären 16 der 18 vorkommenden Beiträge nicht darstellbar, und der Excel-Import
  hätte den Großteil der Zeilen mit "unbekannte Beitragsklasse" abgewiesen.
- Eine Beitragsänderung trifft genau ein Mitglied — es gibt keinen Weg, versehentlich
  fremde Beiträge mitzuändern.
- `mitglied` verliert einen Fremdschlüssel; Anlegen und Import brauchen keine
  vorab existierende Klasse mehr. Eine frische Datenbank startet leer, es wird
  nichts mehr geseedet.

**Schlecht:**

- **Ticket 08 (Beitragsklassen-Ansicht) ist hinfällig**, samt
  `templates/beitragsklassen.html` und `service/beitragsklassen_test.go`.
  Bereits ausgelieferte Arbeit wird zurückgebaut.
- **Story 4** (Filter nach Beitragsklasse) wird zum Filter nach Trainingsfrequenz;
  Such- und Filterlogik aus Ticket 06 samt [ADR-0004](0004-suche-und-filter-im-speicher.md)
  muss entsprechend nachgezogen werden.
- **Story 13** verliert ihren Sinn und entfällt. Eine Preiserhöhung ist künftig
  eine Änderung an vielen Mitgliedern. Sollte das je eine Massenoperation
  erfordern ("alle 60 € auf 65 €"), ist das ein eigenes Feature und keine
  Eigenschaft des Datenmodells.
- Ein Tippfehler im Beitrag fällt nicht mehr durch einen Fremdschlüssel auf.
  Die Excel-Werteliste ist kein Ersatz dafür — sie ist eine Hilfe beim Eintippen,
  keine Einschränkung.

Der Zeitpunkt ist bewusst gewählt: es gibt noch keinen Release, und das Schema
entsteht über `CREATE TABLE IF NOT EXISTS` ohne Migrationsmechanismus. Die
Änderung kostet heute die Dev-Datenbank. Nach dem ersten Mac-Release hätte sie
eine echte Datenmigration gekostet.

## Alternativen

- **Tabelle behalten, 18 Zeilen seeden** — an Story 13 gescheitert, siehe oben.
- **Tabelle behalten, Bedeutung auf die Frequenz umdeuten** (3 Zeilen: 1×/2×/3×)
  — dann ist "Beitragsklasse" ein Name, der nichts mit dem Beitrag zu tun hat.
  Die Frequenz gehört als eigene Eigenschaft ans Mitglied, nicht in eine Tabelle
  mit drei Zeilen und einem irreführenden Namen.
- **Beides: Klasse als Regelpreis, individueller Betrag als Ausnahme** — zwei
  Quellen der Wahrheit für denselben Wert. Bei 200 Mitgliedern und einem Admin
  ist die Frage "gilt hier die Klasse oder die Ausnahme?" teurer als der
  Verzicht auf Klassen.
