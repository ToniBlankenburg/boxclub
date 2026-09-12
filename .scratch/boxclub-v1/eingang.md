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

| # | Anforderung | Scope | Weg |
|---|---|---|---|
| 1 | Vertrag als PDF am Mitglied speichern | v1 | `/grill-with-docs` → ADR → Ticket |
| 2 | Trainingstermine als eigene Entität, eigene View, Zuordnung zur Mitgliedschaft | v1 | `/domain-modeling` → `/to-spec` → `/to-tickets` |
| 3 | Mitglieds-ID in der Liste anzeigen | v1 | direkt Ticket |
| 4 | Rechnungen erstellen | **v2** | vertagt, siehe `CLAUDE.md` → Out of scope |
| 5 | Spalten ein- und ausblenden | v1 | Ticket, zusammen mit 6 |
| 6 | Inhalt nach Spalte sortieren | v1 | Ticket, zusammen mit 5 |
| 7 | Finanzüberblick in eigener View | **v2** | vertagt, hängt ohnehin an 4 |

## Getroffene Entscheidungen

**4 und 7 sind v2.** Rechnungen und Finanzüberblick führen das Zahlungsmodell
ein, das [ADR-0006](../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md) bewusst
weggelassen hat. Sie brauchen einen eigenen ADR, der 0006 in Teilen ablöst —
nicht ein Ticket, das man einfach baut. Festgehalten in `CLAUDE.md`.

**2 ist ein Terminkatalog und keine Anwesenheitsverfolgung.** Gepflegt wird eine
Liste echter Trainingstermine, der Mitgliedschaften zugeordnet werden; *wer wann
da war*, wird nicht festgehalten. Damit fällt es nicht unter das in `CLAUDE.md`
ausgeschlossene *attendance tracking* und ist v1-tauglich. Festgehalten in
`CLAUDE.md`.

## Was vor den Tickets noch fehlt

**Zu 2:** Das Material von Angelo („schickt angelo"). Ohne seine Vorstellung von
der eigenen View ist die Anforderung nicht scharf genug für ein Ticket.

**Zu 2, unabhängig von Angelo:** `CONTEXT.md` führt heute **Trainingsslot** (der
wöchentliche Termin als Freitext, an der Mitgliedschaft) und
**Trainingsfrequenz** (die Anzahl, nie gespeichert). Ein Katalog macht aus dem
Freitext einen Verweis — beide Glossareinträge müssen neu geschrieben werden,
bevor Code entsteht, sonst heißen drei Dinge gleichzeitig „Training".

**Zu 2, technisch:** Der gerade gebaute Excel-Import (Ticket 09) liest die
Spalten `Training - 1/2/3` als Freitext direkt in Slots. Werden Slots zu
Verweisen auf einen Katalog, muss der Importer sich entscheiden: unbekannte
Termine anlegen oder in den Fehlerbericht geben. Das Ticket zum Katalog muss den
Importer ausdrücklich mit anfassen.

**Zu 1:** Wo liegen die PDFs? Als Blob in SQLite oder als Datei neben der
Datenbank? Was passiert beim Löschen eines Mitglieds? Wie verträgt sich das mit
[ADR-0003](../../docs/adr/0003-datenbank-im-benutzer-konfigurationsordner.md) und
mit „meine Mitgliederdaten verlassen die Festplatte nicht" (Spec, Story 37)? Das
ist eine ADR-Frage, kein Implementierungsdetail.

**Zu 6:** Sortiert wird heute fest nach Nachname (`nachNamenSortieren`).
Sortierung nach beliebiger Spalte fällt unter
[ADR-0004](../../docs/adr/0004-suche-und-filter-im-speicher.md) — im Speicher,
nicht in SQL.

## Was sofort gehen würde

**3** ist ein Zehn-Zeilen-Ticket: `Listeneintrag.MitgliedID` existiert und wird in
jeder htmx-URL der Liste benutzt, nur nirgends *angezeigt*. Seit Ticket 09 sind
das die echten Vereinsnummern aus der Excel — der Wert steigt dadurch deutlich.
