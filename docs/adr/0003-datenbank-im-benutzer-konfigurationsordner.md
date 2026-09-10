# ADR-0003: Die SQLite-Datei liegt im Benutzer-Konfigurationsordner

**Status:** Accepted
**Datum:** 2026-09-10

## Kontext

Die Spec legt fest: *"Datenbank: SQLite, Datei `boxclub.db` neben der Anwendung."*
Beim ersten Vertikal-Cut (Ticket 02) musste der Ort konkret werden.

"Neben der Anwendung" verträgt sich nicht mit dem Produktionsziel aus
[ADR-0001](0001-go-wails-fuer-desktop-gui.md), macOS. Dort liefert
`wails build` ein `.app`-Bundle aus; neben dem Binary heißt dort
`Boxclub.app/Contents/MacOS/`. Dieses Verzeichnis ist Teil der Code-Signatur:

- Eine Datei, die zur Laufzeit darin entsteht oder wächst, bricht die Signatur —
  Gatekeeper verweigert die App nach dem ersten Start oder nach dem Verschieben.
- Beim Ersetzen der App durch eine neue Version wandert die Datenbank mit in den
  Papierkorb. Genau die Daten, die die Excel-Tabelle ablösen sollen, hingen dann
  am Programm statt am Benutzer.
- Liegt die App in `/Applications`, ist das Verzeichnis für den Benutzer nicht
  schreibbar.

## Entscheidung

Die Datenbank liegt in einem eigenen Verzeichnis unterhalb von
`os.UserConfigDir()`, auf macOS also
`~/Library/Application Support/Boxclub/boxclub.db`. Das Verzeichnis wird beim
Start mit `0700` angelegt.

Die Umgebungsvariable `BOXCLUB_DB` überschreibt den Pfad. Sie ist kein Feature
für den Vereinsadmin, sondern für Entwicklung und manuelle Rauchtests: ohne sie
schriebe jeder `wails dev`-Lauf in dieselbe Datei wie die installierte App.

## Betrachtete Alternativen

### Wörtlich neben dem Binary

**Pro:** Spec-treu, ein Verzeichnis zum Sichern, portabel per USB-Stick.
**Contra:** die drei macOS-Probleme oben. Für Windows/Linux-Dev-Builds
unproblematisch — aber die sind laut Spec ausdrücklich kein Auslieferungsziel.

### Neben dem Binary, aber `~/Documents/Boxclub` auf macOS

**Pro:** für den Nutzer sichtbar und leicht selbst zu sichern.
**Contra:** plattformabhängige Sonderregel im Code, und der Ordner gehört dem
Nutzer — eine App-Datenbank, die man versehentlich verschieben kann.

## Konsequenzen

- **Positiv:** Die Daten überleben App-Updates und Neuinstallationen. Time
  Machine und FileVault decken `~/Library/Application Support` ab, womit die
  Backup-Zusage aus der Spec ("FileVault + Time Machine ist die externe
  Antwort") trägt. `0700` hält die personenbezogenen Daten aus anderen
  Benutzerkonten heraus.
- **Negativ:** Der Pfad ist im Finder standardmäßig ausgeblendet; wer die Datei
  von Hand kopieren will, muss ihn kennen. Er steht deshalb in der `README.md`.
- **Nebeneffekt beim Bauen:** `wails build` und `wails dev` erzeugen zur
  Bindings-Generierung ein Binary und führen es aus. Weil `main()` die Datenbank
  vor `wails.Run` öffnet, legt schon der Build die Datei an, wenn sie fehlt. Das
  Anlegen ist idempotent (`CREATE TABLE IF NOT EXISTS`, `INSERT OR IGNORE`) und
  verliert keine Daten; wer es beim Entwickeln vermeiden will, setzt
  `BOXCLUB_DB`.
- **Spec-Korrektur:** Die Zeile "Datei `boxclub.db` neben der Anwendung" in
  `.scratch/boxclub-v1/spec.md` ist damit überholt.
