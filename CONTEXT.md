# Boxclub Mitgliederverwaltung – Kontext

Glossar der Begriffe, wie sie in diesem Projekt verwendet werden. Kein Implementierungsdetail.

## Mitglied

Eine natürliche Person, die dem Verein bekannt ist. Ein Mitglied bleibt derselbe Datensatz, auch wenn es zwischenzeitlich austritt und wieder eintritt.

**Nicht zu verwechseln mit:** Mitgliedschaft (das ist die zeitliche Zuordnung).

## Mitgliedschaft

Der zeitliche Zeitraum, in dem ein Mitglied aktiv im Verein ist. Hat ein Eintritts- und optional ein Austrittsdatum. Ein Mitglied kann mehrere Mitgliedschaften über die Zeit haben (Aus- und Wiedereintritt).

## Beitragsklasse

Eine Preisstufe. Die Staffelung ergibt sich bei diesem Verein **ausschließlich aus der Trainingsfrequenz**. Alter, Familienstand und Ermäßigungen spielen bewusst keine Rolle. Aktuell existieren genau zwei Klassen:

- **Erwachsen 1× / Woche** — 60 €/Monat
- **Erwachsen 2× / Woche** — 80 €/Monat

Jedes Mitglied ist genau einer Beitragsklasse zugeordnet. Klassen und ihre Preise können sich über die Zeit ändern.

## Zahlung

Ein einzelnes Ereignis auf dem Vereinskonto (kommt später aus MoneyMoney), das einem Mitglied zugeordnet werden kann. **In v1 nicht modelliert** — v1 speichert nur den abgeleiteten Status `bezahlt_bis`.

## bezahlt_bis

Datum, bis zu dem die Beiträge des Mitglieds als bezahlt gelten. In v1 pro Mitglied genau ein Datum, manuell gepflegt. In v2 wird dieses Datum aus der Zahlungshistorie (MoneyMoney-CSV-Import) berechnet.

## Statusanzeige

Zweistufig: **bezahlt** (`bezahlt_bis >= heute`) oder **nicht bezahlt** (`bezahlt_bis < heute`). Keine gelbe Vorwarnstufe.
