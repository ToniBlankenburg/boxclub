# Bestandsaufnahme: bestehende Excel-Tabelle

Quelle: Muster-Datei des Vereinsadmins (nicht im Repo — enthält echte
personenbezogene Daten inkl. IBAN). **Die Spaltennamen und die Wertelisten
sind echt und produktionsnah**; die Datei enthält nur eine Beispielzeile.
Die Beispielwerte unten sind anonymisiert.

## Blatt 1: „Verwaltung" — die Mitgliederzeilen

21 Spalten, Kopfzeile in Zeile 1, Daten ab Zeile 2.

| # | Spaltenname | Beispielwert (anonymisiert) | Anmerkung |
|---|---|---|---|
| A | `Training - 1` | `Samstag 10:30 Uhr` | Trainingszeit, Dropdown aus Blatt „Quelle" |
| B | `Mandatsreferenz` | `1` | **Kein SEPA-Mandat** trotz des Namens, sondern die laufende Mitglieds-Nummer — die fachliche Identität |
| C | `Vorname` | `Erika` | |
| D | `Nachname` | `Musterfrau` | |
| E | `Beitrag` | `60` | Euro/Monat als Zahl, kein Klassenname |
| F | `Status` | `Neu` | Dropdown, 7 Werte |
| G | `Mitgliedschaft` | `46296` → 2026-10-01 | **Excel-Seriendatum**; Bedeutung unklar (Beginn?) |
| H | `Gekündigt` | *(leer)* | Datum oder leer |
| I | `IBAN` | `DE.. .... .... .... .... ..` | personenbezogen |
| J | `Eintrit` *(sic)* | `46270` → 2026-09-05 | Excel-Seriendatum; Tippfehler im Header |
| K | `1x 2x Woche` | `1x Woche` | Dropdown: 1x/2x/3x Woche |
| L | `Telefonnummer` | `01511 0000000` | Freitext mit Leerzeichen |
| M | `E-Mail` | `erika@example.org` | |
| N | `Adresse` | `Musterstrasse 15/1` | nur Straße + Nr., **ohne** PLZ/Ort |
| O | `Postleitzahl` | `70174` | eigene Spalte |
| P | `Ort` | `Stuttgart` | eigene Spalte |
| Q | `Geburtstag` | `36601` → 2000-03-16 | Excel-Seriendatum |
| R | `Geschlecht` | `Mann` | Dropdown: Frau/Mann |
| S | `Anmeldegebühr` | `60` | Euro, Einmalbetrag |
| T | `Bewertung` | `❌` | Emoji-Freitext |
| U | `Digital` | `Digital` | |

**Datumsformat:** alle Datumsspalten sind Excel-Seriennummern (Epoche
1899-12-30), keine Strings. Der Importer muss beides lesen können, weil in
der echten Tabelle mit 50–200 Zeilen auch handgetippte Datumsstrings stehen
können.

## Blatt 2: „Quelle" — die Dropdown-Wertelisten

- **Beitrag:** 0, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 100, 110,
  115, 135, 144 — danach in derselben Spalte noch `Still`, `Offen`,
  `Bezahlt`, `Gekündigt` (vermutlich eine zweite, angehängte Liste; zu klären)
- **Status:** Mitglied, Neu, Stillgelegt, Gekündigt, Kündigungsfrist,
  Inakiv *(sic)*, Aktiv
- **Gruppen/Trainingszeiten:** ~18 Einträge, Muster
  `<Wochentag> <Uhrzeit> Uhr`, teils Kombinationen
  (`Di - 19:30 - Sa - 10:30 Uhr`), teils Tippfehler (`Mitwoch`), plus `Kein`
- **1x 2x Woche:** 1x Woche, 2x Woche, 3x Woche
- **Geschlecht:** Frau, Mann
- **Bank:** Eingezogen, Bezahlt, Postbank, Fyrst, Offen, Stillgelegt

## Offene Konflikte mit dem aktuellen Modell

Vier Punkte, die vor neuen Tickets entschieden werden müssen — siehe
`CONTEXT.md` und `docs/adr/`:

1. **`beitragsklasse` (zwei Klassen, 60/80 €) hält nicht.** Realität: 18
   freie Beträge von 0 bis 144 €. Betragsfeld am Mitglied oder 18 Klassen?
2. **`Status` (7 Werte) überlagert drei Konzepte:** Lebenszyklus
   (Neu/Aktiv/Mitglied/Stillgelegt/Inaktiv), Kündigung (Gekündigt/
   Kündigungsfrist) und — über die Bank-Liste — Zahlung. Kollidiert mit
   unserem abgeleiteten Zahlungsstatus und mit `mitgliedschaft.austritt IS NULL`.
3. **„Mitgliedschaft" ist in der Excel ein Datum, bei uns eine Entität.**
   Überladener Begriff, gehört ins Glossar.
4. **Felder ohne Platz im Schema:** Trainingszeit/Gruppe, 1x/2x Woche, IBAN +
   Mandatsreferenz, Bankstatus, Anmeldegebühr, Geschlecht, PLZ/Ort getrennt,
   Bewertung, Digital, Gekündigt-Datum.

## Bekanntes Risiko

Die Muster-Datei zeigt das **Schema** verlässlich, aber nicht den **Schmutz**
der echten Tabelle: fehlende Zellen, abweichende Datumsformate, Tippfehler in
Dropdown-Werten, Dubletten. Der Import-Fehlerbericht (Ticket 09) ist deshalb
kein Nice-to-have, sondern das Werkzeug, mit dem der Schmutz überhaupt
sichtbar wird.
