# Eingang: neue Anforderungen (12.09.2026)

Rohnotizen des Vereinsadmins, wortwörtlich übernommen, darunter die Einordnung.
Dies ist **kein Ticket** — es ist der Ort, an dem die Notizen liegen, bis sie
eines geworden sind. Was daraus ein Ticket wird, entscheidet die Spalte „Weg".

## Rohnotiz

```
Vertrag als pdf am mitgliedspeichernam mitglied,
trainingseinheiten als eigenes DTO, das man in extra view erstellen und plegen und dann dem mitglied/ mitgliedschaft zuordnen-> schickt angelo
mitglides nummer (id) -auch anzeigenn
rechnungen erstellen als feature
spalten ein und ausblenden
inhalt nach spalte sortieren
finanz überlick in extra view
```

## Einordnung

Aufgelöst am 12.09.2026 in einer Grilling-Sitzung. Alle sieben Punkte sind
entschieden; jeder hat jetzt entweder ein Ticket oder einen Grund, keins zu haben.

| # | Anforderung | Ergebnis |
|---|---|---|
| 1 | Vertrag als PDF am Mitglied speichern | Ticket 24 — am **Zeitraum**, nicht an der Person |
| 2 | Trainingstermine als eigene Entität mit eigener View | Tickets 20, 21, 22 — [ADR-0008](../../docs/adr/0008-trainingstermine-als-wochenplan.md) |
| 3 | Mitglieds-ID in der Liste anzeigen | **erledigt**, Ticket 19 |
| 4 | Rechnungen erstellen | Ticket 25 (+ 23 für den Briefkopf) — **v1**, siehe unten |
| 5 | Spalten ein- und ausblenden | Ticket 27 |
| 6 | Inhalt nach Spalte sortieren | Ticket 27 |
| 7 | Finanzüberblick in eigener View | Ticket 26 — als *Monatssoll* |

Reihenfolge: **20 → 21 → 22**, dann **23 → 24 → 25**, danach 26 und 27. Der
Terminblock zuerst, weil er als einziger bestehenden Code umbaut; zwischen 20 und
22 ist der Excel-Import kaputt, die drei müssen zusammen fertig werden. 26 und 27
hängen an nichts und sind jederzeit einschiebbar.

## Getroffene Entscheidungen

**4 und 7 sind doch v1 — weil sie anders gemeint waren.** Die frühere Einordnung
als v2 beruhte auf der Annahme, „Rechnungen erstellen" heiße Forderungen
verwalten. Gemeint ist das **Erzeugen eines PDFs** über Leistungen neben dem
Beitrag (Einzeltrainings), ohne Rechnungsstatus und ohne Zahlungseingang; und der
Finanzüberblick ist ein **Dashboard** aus Zahlen, die längst in der Datenbank
stehen. Beides führt kein Zahlungsmodell ein.
[ADR-0006](../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md) wird dadurch
**nicht abgelöst**, sondern präzisiert; das steht in
[ADR-0009](../../docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md).
Der Absatz in `CLAUDE.md` ist entsprechend umgeschrieben.

**2 ist ein Wochenplan und keine Anwesenheitsverfolgung.** Gepflegt wird eine
Liste **wiederkehrender** Wochentermine, für die sich Mitgliedschaften anmelden —
keine datierten Einheiten. Das ist die Grenze zum ausgeschlossenen attendance
tracking: ein Wochenplan hat keine Stelle, an der stehen könnte, wer da war.
Termine werden **archiviert, nie gelöscht**, und der Excel-Import legt selbst
keine an, sondern meldet Unbekanntes im Fehlerbericht
([ADR-0008](../../docs/adr/0008-trainingstermine-als-wochenplan.md)).

**PDFs liegen als Blob in der Datenbank**, nicht als Dateien daneben. Der
Vereinsadmin sichert, indem er eine Datei kopiert
([ADR-0007](../../docs/adr/0007-dokumente-als-blob-in-sqlite.md)).

**Ein Vertrag hängt an der Mitgliedschaft, eine Rechnung am Mitglied.** Wer
austritt und wiederkommt, unterschreibt einen neuen Vertrag; ein Einzeltraining
hat mit keinem Zeitraum zu tun.

**Die Rechnungsnummer wird eingetippt.** Der Verein führt seine Nummernfolge in
der Buchhaltung; eine App, neben der noch anders Rechnungen entstehen, kann
Lückenlosigkeit nicht versprechen.

**Das Dashboard zeigt ein Soll, keinen Verlauf.** Es gibt keine Beitragshistorie,
also wäre jede Kurve eine Hochrechnung im Gewand einer Messung.

## Erledigte Vorbehalte

Die Punkte, die hier als offen standen, sind es nicht mehr:

**Angelos Material** wird nicht abgewartet. Was er liefert, ist das Aussehen der
View und die konkreten Trainingszeiten — Oberfläche und Inhalt, nicht Modell. Die
Frage, was ein Trainingstermin ist und woran er hängt, ist beantwortet.

**Die Glossareinträge** sind neu geschrieben: *Trainingstermin* kommt hinzu,
*Trainingsslot* ist als abgelöst gekennzeichnet, *Trainingsfrequenz* zählt jetzt
Termine. Dazu neu: *Dokument*, *Vertrag*, *Rechnung*, *Vereinsdaten*, *Monatssoll*.

**Der Importer** wird ausdrücklich mit angefasst — dafür gibt es Ticket 22.

**Der Ablageort der PDFs** ist entschieden (ADR-0007). Die Anschlussfrage „was
passiert beim Löschen eines Mitglieds?" hat sich erledigt: die App hat keinen
Löschpfad für Mitglieder, nirgends.

**Die Sortierung** fällt wie vermutet unter ADR-0004 und läuft im Speicher; die
Spaltenwahl dagegen bleibt im Browser, weil sie nur die Ansicht eines einzelnen
Geräts betrifft (Ticket 27).

**Die lose Datei `anforderungen`** im Wurzelverzeichnis ist gelöscht. Die Rohnotiz
oben ist ihre wortwörtliche Kopie.
