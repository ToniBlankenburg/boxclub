Status: ready-for-agent

# Spec: Zahler als technische Entität

## Problem Statement

Heute zahlt in der App immer das Mitglied selbst: Name und IBAN der Person
sind zugleich Name und Konto der Lastschrift (MoneyMoney-Export, ADR-0013)
und der Vorbelegung des Rechnungsempfängers. Bei Kindern stimmt das nicht —
die Eltern zahlen, das Kind trainiert. Der Admin kann das heute nur
abbilden, indem er die Daten der Eltern in die Felder des Kindes tippt;
dann ist der Name des Kindes falsch, oder die Daten der Eltern stehen
mehrfach in der Datenbank, einmal je Kind.

## Solution

Eine technische Entität **Zahler** trennt „wer trainiert“ (Mitglied) von
„wer zahlt“. Jede Mitgliedschaft hat genau einen Zahler. Für Erwachsene ist
das unsichtbar: der Zahler ist die Person selbst. Für Kinder (Mitglieder
unter 18) schaltet das Formular das Feld „Abweichender Kontoinhaber“ frei.
Der Zahler taucht in der Oberfläche **nicht als Begriff** auf — es gibt
keine Zahler-Ansicht, keine Zahler-Liste, kein Zahler-Menü; die UI sagt nur
„Kontoinhaber“ und „IBAN“.

Geschwister mit demselben Konto teilen sich automatisch einen Zahler, weil
Zahler über die normalisierte IBAN zusammengeführt werden.

## User Stories

1. Als Vereinsadmin möchte ich bei einem Kind einen abweichenden
   Kontoinhaber eintragen, damit der Lastschrifteinzug auf das Konto der
   Eltern läuft und nicht auf das des Kindes.
2. Als Vereinsadmin möchte ich bei einem Erwachsenen nichts Neues sehen und
   nichts zusätzlich pflegen müssen, damit sich der Alltag nicht ändert.
3. Als Vereinsadmin möchte ich eine einzige IBAN-Zeile im Formular haben,
   damit ich nicht überlegen muss, welche von zwei IBANs gilt.
4. Als Vereinsadmin möchte ich, dass das Feld „Abweichender Kontoinhaber“
   nur bei Kindern aktiv ist, damit ich es bei Erwachsenen nicht versehentlich
   fülle.
5. Als Vereinsadmin möchte ich, dass sich der Kind-Status aus dem
   Geburtsdatum ergibt, damit ich ihn nicht pflegen und nicht vergessen kann.
6. Als Vereinsadmin möchte ich bei einem Mitglied ohne Geburtsdatum den
   Kind-Status von Hand setzen können, damit auch Mitglieder ohne bekanntes
   Alter Eltern als Zahler haben können.
7. Als Vereinsadmin möchte ich einen Hinweis sehen, wenn ein Mitglied mit
   abweichendem Kontoinhaber volljährig geworden ist, damit ich prüfe, ob
   die Eltern weiter zahlen sollen.
8. Als Vereinsadmin möchte ich, dass sich beim 18. Geburtstag nichts still
   ändert, damit die Lastschrift nicht plötzlich vom falschen Konto läuft.
9. Als Vereinsadmin möchte ich bei Geschwistern mit demselben Konto die
   IBAN nur einmal pflegen müssen, damit eine neue IBAN der Eltern beide
   Kinder erreicht.
10. Als Vereinsadmin möchte ich, dass `DE89 3704 0044 0532 0130 00` und
    `de89370400440532013000` dasselbe Konto sind, damit Tippweise und
    Leerzeichen keinen zweiten Zahler erzeugen.
11. Als Vereinsadmin möchte ich, dass ein Mitglied ohne IBAN nie mit einem
    anderen Mitglied ohne IBAN als derselbe Zahler zusammengelegt wird,
    damit leere Felder keine Verbindung herstellen.
12. Als Vereinsadmin möchte ich, dass beim Umbenennen eines Zahlers
    (Eltern heiraten) der neue Name für alle Kinder gilt, damit der Export
    nicht zwei Schreibweisen für dasselbe Konto enthält.
13. Als Vereinsadmin möchte ich, dass ein Zahler verschwindet, sobald keine
    Mitgliedschaft mehr auf ihn zeigt, damit ich nichts aufräumen muss.
14. Als Vereinsadmin möchte ich bei einem Wiedereintritt den bisherigen
    Zahler vorgeschlagen bekommen, damit ich ihn nicht neu eintippen muss.
15. Als Vereinsadmin möchte ich, dass der MoneyMoney-Export als
    „Zahlungspflichtiger“ den Kontoinhaber nennt und dessen IBAN verwendet,
    damit die Lastschrift auf das richtige Konto geht.
16. Als Vereinsadmin möchte ich, dass die Mandatsreferenz im Export die
    Mitglieds-ID bleibt, damit sich an den bestehenden Mandaten nichts ändert.
17. Als Vereinsadmin möchte ich für zwei Kinder derselben Eltern zwei
    Exportzeilen (je Mitgliedschaft eine) sehen, damit jede Abbuchung ihren
    Beitrag und ihre Anmeldegebühr trägt.
18. Als Vereinsadmin möchte ich, dass der Rechnungsempfänger bei einem Kind
    mit dem Namen des Kontoinhabers vorbelegt wird, damit ich eine Rechnung
    für ein Einzeltraining an die Eltern adressiere.
19. Als Vereinsadmin möchte ich, dass die Anschrift auf der Rechnung vom
    Mitglied kommt, damit ich nicht für den Zahler eine eigene Anschrift
    pflegen muss.
20. Als Vereinsadmin möchte ich den vorbelegten Rechnungsempfänger weiterhin
    überschreiben können, damit Sonderfälle möglich bleiben.
21. Als Vereinsadmin möchte ich, dass die Rechnung als Dokument am Mitglied
    bleibt, damit ich sie dort wiederfinde.
22. Als Vereinsadmin möchte ich, dass der Rückstand am Mitglied bleibt,
    damit sich die Rückstandsliste nicht verändert.
23. Als Vereinsadmin möchte ich, dass der Excel-Import Mitglied,
    Mitgliedschaft und Zahler anlegt, damit ich nach dem Import keine
    Zahler nacharbeiten muss.
24. Als Vereinsadmin möchte ich, dass der Import den Kind-Status aus dem
    Geburtsdatum ableitet, damit der Import keine neue Spalte braucht.
25. Als Vereinsadmin möchte ich, dass ein Mitglied, das selbst zahlt, nach
    einer Namensänderung des Mitglieds im Export unter dem neuen Namen
    erscheint, damit der Zahler nicht veraltet.
26. Als Vereinsadmin möchte ich, dass sich die Oberfläche in Deutsch und
    Englisch gleich verhält, damit der neue Hinweis und die Beschriftungen
    in beiden Sprachen vorhanden sind.
27. Als Entwickler möchte ich, dass `CONTEXT.md` den Zahler, das Kind und
    den Kontoinhaber definiert, damit künftige Arbeit dieselben Wörter
    benutzt.
28. Als Entwickler möchte ich eine ADR, die die Normalisierung der IBAN als
    Abweichung von ADR-0006 festhält, damit sie nicht als Fehler gelesen wird.

## Implementation Decisions

**Begriffe**

- **Zahler** ist rein technisch: eine Entität, die Mitglied und Bezahlenden
  trennt. In der UI heißt das Feld **Kontoinhaber**; das Wort „Zahler“
  erscheint dort nicht. In Code, Tests und `CONTEXT.md` heißt es *Zahler*.
- **Kind** ist ein Mitglied, das am heutigen Tag jünger als 18 ist. Die
  Eigenschaft wird bei jedem Lesen **aus dem Geburtsdatum abgeleitet**, nie
  gespeichert (wie die Trainingsfrequenz). Hat das Mitglied kein
  Geburtsdatum, zählt ein von Hand gesetztes Kennzeichen an der Person.

**Schema** (keine Migration: die Entwicklungsdatenbank wird gelöscht, siehe
CLAUDE.md zu Umdeutung)

- Neue Tabelle `zahler(id, name, iban, mitglied_id NULL)`.
  - `iban` wird beim Schreiben **immer normalisiert** (Leerzeichen entfernt,
    Großbuchstaben). Das ist eine bewusste Abweichung von „IBAN als reiner
    Text“ aus ADR-0006 und gilt nur für diese Spalte; Validierung und
    Formatprüfung bleiben ausgeschlossen.
  - `mitglied_id` ist gesetzt, wenn das Mitglied selbst zahlt; der Name kommt
    dann beim Lesen aus der Person, nicht aus `zahler.name`.
- `mitgliedschaft.zahler_id` ist `NOT NULL`: jede Mitgliedschaft zeigt auf
  genau einen Zahler. Eine Mitgliedschaft hat nie mehrere Zahler.
- `mitglied.iban` entfällt. Die IBAN lebt am Zahler.
- An der Person kommt ein Kennzeichen für den von Hand gesetzten Kind-Status
  hinzu; es zählt nur, solange kein Geburtsdatum da ist.
- Anschrift, Rückstand und Rückstandsnotiz bleiben an der Person. Der Zahler
  hat bewusst **nur Name und IBAN**.

**Zusammenführen und Aufräumen**

- Beim Speichern einer Mitgliedschaft wird der Zahler über die normalisierte
  IBAN gesucht. Gibt es einen mit gleicher IBAN, wird er wiederverwendet,
  sonst entsteht einer. Tragen zwei Mitgliedschaften dieselbe IBAN mit
  unterschiedlichen Namen ein, gewinnt der zuletzt gespeicherte Name und
  gilt für beide.
- Eine leere IBAN bildet **keinen** gemeinsamen Zahler: jede Mitgliedschaft
  ohne IBAN bekommt eine eigene Zeile.
- Ein Zahler, auf den keine Mitgliedschaft mehr zeigt, wird automatisch
  gelöscht. Es gibt keine Verwaltung von Hand.

**Verhalten im Lebenszyklus**

- Ein Wiedereintritt übernimmt den Zahler der Vorgängerin als Vorschlag;
  die Mitgliedschaft hat danach trotzdem einen eigenen, änderbaren Verweis.
- Wird ein Mitglied 18 und hat einen abweichenden Kontoinhaber, ändert sich
  **nichts** an den Abbuchungen: der Zahler bleibt wirksam, das Feld ist
  gesperrt, aber gefüllt, und das Formular zeigt einen **Hinweis**, dass das
  Mitglied jetzt volljährig ist und der Kontoinhaber geprüft werden soll.
  Erst wenn der Admin den Kontoinhaber leert, zahlt das Mitglied selbst.

**Formular** (htmx-Fragment, keine Zahler-Ansicht)

- Es gibt eine IBAN-Zeile. Das Feld „Abweichender Kontoinhaber“ steht
  daneben und ist nur aktiv, solange das Mitglied ein Kind ist. Bei
  Erwachsenen ist es deaktiviert (und leer). Hinweis und Beschriftungen
  kommen in beiden Sprachen in die Kataloge (ADR-0017); Fehlermeldungen
  bleiben Schlüssel (ADR-0018).

**Andere Abnehmer**

- **MoneyMoney-Export:** „Zahlungspflichtiger“ und IBAN kommen vom Zahler;
  eine Zeile je Mitgliedschaft; Mandatsreferenz bleibt die Mitglieds-ID;
  Beitrag und Anmeldegebühr unverändert (ADR-0013).
- **Rechnung:** Empfängername aus dem Zahler der aktuellen Mitgliedschaft,
  Anschrift aus dem Mitglied; weiter überschreibbar; Dokument bleibt am
  Mitglied (ADR-0007, ADR-0009).
- **Excel-Importer:** legt je Zeile Mitglied, Mitgliedschaft und Zahler
  „selbst“ mit der IBAN an. Er führt kein Kind-Kennzeichen ein; der
  Status ergibt sich aus dem Geburtsdatum. Seine Schnittstelle (Zeilen plus
  Fehlerbericht) ändert sich nicht, nur der Weg, auf dem die IBAN in die
  Datenbank kommt.
- **Rückstand, Dashboard, Monatssoll, Suche, Spaltenmenü, Serienmail:**
  unverändert, es gibt keine neue Spalte.

**Dokumentation**

- `CONTEXT.md` bekommt die Begriffe *Zahler*, *Kind* und *Kontoinhaber*
  (UI-Wort), mit „nicht zu verwechseln mit“ gegenüber dem Mitglied.
- Eine neue ADR hält fest: Zahler als technische Entität; IBAN-Normalisierung
  weicht an einer Stelle von ADR-0006 ab; `mitglied.iban` entfällt;
  Verweis auf ADR-0013 und ADR-0019. `CLAUDE.md` wird im Schemaabschnitt
  nachgezogen.

## Testing Decisions

- **Ein Seam:** die öffentliche API von `MemberService` (CLAUDE.md). Der
  Zahler hat keine eigene API, er ist vollständig über Anlegen und Ändern der
  Mitgliedschaft, den Export und die Rechnungsvorbelegung beobachtbar. Der
  Importer wird wie bisher über sein Ergebnis plus `MemberService`
  geprüft.
- Ein guter Test prüft **von außen sichtbares Verhalten**: wer im Export als
  Zahlungspflichtiger steht, welche IBAN verwendet wird, ob zwei
  Mitgliedschaften dieselbe Zeile teilen (beobachtbar daran, dass eine
  Namensänderung bei beiden ankommt), ob ein Zahler nach dem Entfernen
  verschwindet (beobachtbar an einer erneut angelegten Mitgliedschaft ohne
  Altlast). Nie direkt in SQL schauen.
- Zu testen: Normalisierung der IBAN; Zusammenführen und „zuletzt
  gespeicherter Name gewinnt“; leere IBAN bleibt getrennt; Kind-Status aus
  dem Geburtsdatum (Stichtage um den 18. Geburtstag) und Handkennzeichen ohne
  Geburtsdatum; Hinweis bei Volljährigkeit ohne Änderung der Abbuchung;
  Wiedereintritt schlägt den Zahler vor; Export nimmt Name und IBAN vom
  Zahler; Rechnungsvorbelegung nimmt Name vom Zahler und Anschrift vom
  Mitglied; Importer legt Zahler „selbst“ an.
- Prior art: `service/member_service_test.go` (Aufbau mit `t.TempDir()` und
  echter SQLite), `service/moneymoney_test.go` (Export),
  `service/rechnung_test.go`, `service/import_test.go`,
  `service/lebenszyklus_test.go` (Wiedereintritt). `app/`-Handler und
  Templates werden wie üblich nur von Hand geprüft.

## Out of Scope

- Mehrere Zahler je Mitgliedschaft (geteilte Beiträge).
- Eine Zahler-Ansicht, -Liste oder -Verwaltung in der Oberfläche.
- Eine Anschrift, E-Mail oder Telefonnummer am Zahler; Rückstand am Zahler.
- Sammellastschrift (eine Exportzeile je Zahler) und Verwendungszweck mit
  mehreren Kindern.
- IBAN-Validierung, BIC, SEPA-Mandate.
- Datenmigration bestehender Datenbanken: die Entwicklungsdatenbank wird
  gelöscht.
- Eine Spalte oder ein Filter „Kind“ in der Mitgliederliste.

## Further Notes

- Beim Zusammenführen über die IBAN gewinnt der zuletzt gespeicherte Name.
  Das ist eine bewusste Vereinfachung: dasselbe Konto hat einen Namen.
- Der Hinweis bei Volljährigkeit ist eine reine Anzeige im Formular, keine
  Aufgabenliste und kein Benachrichtigungssystem.
