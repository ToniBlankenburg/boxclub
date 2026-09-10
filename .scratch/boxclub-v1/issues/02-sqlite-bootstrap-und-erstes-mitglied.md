Status: ready-for-human

# 02: SQLite-Bootstrap & erstes Mitglied anlegen

**What to build:** Die App bindet SQLite an und legt beim ersten Start eine Datenbankdatei mit Schema und Seed-Daten (die zwei **Beitragsklassen**) an. Der Nutzer kann in einem Formular die Stammdaten eines **Mitglied**s eingeben, speichern und den gespeicherten Datensatz sofort sehen. Erster echter Vertikal-Cut durch alle Schichten.

**Blocked by:** 01 (Wails-Projekt-Skelett)

## Acceptance Criteria

- [x] Beim Start ohne bestehende DB wird `boxclub.db` angelegt und mit Schema für `mitglied`, `mitgliedschaft`, `beitragsklasse` initialisiert
- [x] Die zwei Beitragsklassen `Erwachsen 1×/Woche` (60 €/Monat) und `Erwachsen 2×/Woche` (80 €/Monat) sind geseedet, Preise in Cent gespeichert
- [x] Formular erlaubt: Vorname, Nachname, Geburtsdatum, Adresse, E-Mail, Telefon, Eintrittsdatum, Beitragsklasse (Auswahl)
- [x] `MemberService.Create` legt gleichzeitig einen `mitglied`- und einen `mitgliedschaft`-Datensatz (`austritt` NULL) an
- [x] Nach dem Speichern wird der angelegte Datensatz im UI angezeigt (`MemberService.Get`)
- [x] Beim Neustart der App bleibt das Mitglied erhalten
- [x] Tests am `MemberService`-Seam mit echtem SQLite in `t.TempDir()`: Create → Get liefert alle Werte zurück; Fremdschlüssel auf `beitragsklasse` funktioniert; Doppelanlage mit demselben Vor-/Nachnamen + Geburtsdatum bleibt möglich (kein UNIQUE-Constraint darauf)
- [x] `MemberService.Create` etabliert als Referenz-Testdatei das Setup-Muster für alle späteren Service-Tests

## Notes

Treiber: **`modernc.org/sqlite`** (pure-Go). Kein Repository-Interface (siehe spec.md → "Bewusst nicht modelliert"). Vokabular: [CONTEXT.md](../../../CONTEXT.md).

## Comments

### Implementiert (Linux verifiziert, Windows offen)

Vertikal-Cut steht: `service/` → `app/` → `templates/` → WebView. Sechs Tests am
`MemberService`-Seam gegen echtes SQLite in `t.TempDir()`, alle grün;
`gofmt`, `go vet` und `staticcheck` sauber.

`service/member_service_test.go` ist die Referenz-Testdatei. Das Setup-Muster
steckt in `neuerService(t)`: frische DB pro Test, `t.Cleanup` schließt sie,
Beobachtung ausschließlich über die Service-API. Alle späteren Service-Tests
spiegeln das.

**Auf Linux (Mint 22.1) verifiziert:** `wails build` erzeugt
`build/bin/boxclub`; `wails dev` startet, das Formular rendert im WebView, und
Anlegen, Anzeige und Pflichtfeld-Meldungen wurden am laufenden Dev-Server gegen
`/api/...` durchgespielt.

**Noch offen — `wails dev` / `wails build` unter Windows,** wie schon bei
Ticket 01. Wert für einen zweiten Blick auf dem Windows-Notebook, weil dieses
Ticket zum ersten Mal SQLite (`modernc.org/sqlite`, pure-Go — sollte
unproblematisch sein) und `os.UserConfigDir()` benutzt.

### Drei Entscheidungen, die über das Ticket hinausgingen

1. **htmx spricht HTTP, nicht Wails-Bindings** — [ADR-0002](../../../docs/adr/0002-htmx-ueber-den-wails-assetserver.md).
   Bindings sind JS-Funktionen mit Promise; htmx will HTTP. Beides zu verheiraten
   hätte bedeutet, htmx' Swap-Logik von Hand nachzubauen. Stattdessen hängt ein
   `http.ServeMux` aus `app/` an `assetserver.Options.Handler`. `spec.md` und
   `CLAUDE.md` sind nachgezogen.

2. **Datenbank im Benutzer-Konfigurationsordner statt neben der Anwendung** —
   [ADR-0003](../../../docs/adr/0003-datenbank-im-benutzer-konfigurationsordner.md).
   "Neben der Anwendung" heißt auf macOS `Boxclub.app/Contents/MacOS/` — dort zu
   schreiben bricht die Code-Signatur, und die Daten wandern beim App-Update in
   den Papierkorb. Jetzt: `~/Library/Application Support/Boxclub/boxclub.db`,
   Verzeichnis mit `0700`. `BOXCLUB_DB` überschreibt den Pfad für Dev-Läufe.

3. **`frontend/vite.config.js` ist zurück** — aus zwei funktionalen Gründen, beide
   innerhalb von Wails' eingebautem npm-Schritt, also ohne eigene Pipeline:
   - `appType: 'mpa'`. Vites Dev-Server beantwortet sonst jeden unbekannten
     GET-Pfad mit `index.html`. Der Wails-Assetserver reicht aber nur bei 404 an
     Go weiter — unter `wails dev` lieferten GET-Aufrufe auf `/api/...` deshalb
     still die Startseite zurück statt des Fragments. Unter `wails build` tritt
     das nicht auf, weil dort die eingebettete `embed.FS` bedient wird.
   - Das Tailwind-v4-Plugin. Dazu `@source "../../templates"` in
     `frontend/src/style.css`, weil die Utility-Klassen in den Go-Templates
     außerhalb von `frontend/` stehen und Tailwind sie sonst nicht findet.
     Verifiziert: alle 52 in den Templates verwendeten Klassen sind im gebauten
     CSS definiert.

### Nach Code-Review angepasst

- **Tailwind war weggefallen.** Erste Fassung hatte 127 Zeilen handgeschriebenes
  CSS — Verstoß gegen ADR-0001 und `CLAUDE.md`. Jetzt Tailwind v4 über das
  Vite-Plugin, `style.css` ist auf zwei Zeilen geschrumpft.
- **Validierung lag doppelt** in `service/` und `app/`, mit zwei Formulierungen
  derselben Regel. Die Pflichtfeld-Regeln liegen jetzt nur noch im Service und
  kommen als `*service.ValidierungsFehler` mit allen Meldungen auf einmal zurück;
  `app/` prüft nur noch, was wirklich Adapter-Sache ist (unparsebare Datums- und
  Zahlwerte) und zeigt den Rest an.
- **Klassen-Suche wanderte in den Service.** `app/` lief vorher selbst durch die
  Beitragsklassen-Liste, um den Namen zur ID zu finden — jetzt
  `MemberService.Beitragsklasse(id)`.
- **`Beitragsklassen()` → `AktiveBeitragsklassen()`** mit `WHERE aktiv = 1`. Das
  Formular hätte sonst deaktivierte Klassen zur Auswahl angeboten. Ticket 08
  braucht zusätzlich eine Sicht auf alle Klassen inklusive deaktivierter.
- **`<select>` hatte keine Leer-Option,** deshalb war immer die erste Klasse
  vorausgewählt, `required` griff nie und der Zweig "Bitte eine Beitragsklasse
  wählen." war unerreichbar. Jetzt mit `<option value="" disabled selected>`.
- Tote Route `POST /api/mitglied/formular` entfernt; Template-Funktion `text` zu
  `oderStrich` umbenannt; sieben fast identische `<label>`-Blöcke im Formular auf
  ein Teil-Template `feld` zusammengezogen.
- Wortlaut in ADR-0002 geschärft: maßgeblich ist nicht "genau ein Service-Aufruf
  pro Handler", sondern dass keine Fachregel in `app/` liegt.

### Bekannter Nebeneffekt

`wails build` und `wails dev` erzeugen zur Bindings-Generierung ein Binary und
führen es aus. Weil `main()` die Datenbank vor `wails.Run` öffnet, legt schon
der Build die Datei an, wenn sie fehlt. Das ist idempotent
(`CREATE TABLE IF NOT EXISTS`, `INSERT OR IGNORE`) und verliert keine Daten; wer
es beim Entwickeln vermeiden will, setzt `BOXCLUB_DB`. Es aufzulösen hieße, die
DB erst in `OnStartup` zu öffnen und den Handler auf einen noch leeren Service
zeigen zu lassen — dafür ist der Nebeneffekt zu harmlos.

### Für spätere Tickets vorgemerkt

- **Ticket 08** braucht neben `AktiveBeitragsklassen()` eine Sicht auf alle
  Klassen inklusive deaktivierter, sobald Deaktivieren möglich wird.
- **Ticket 10** muss `BOXCLUB_DB` und den DB-Ort aus ADR-0003 dokumentieren.
