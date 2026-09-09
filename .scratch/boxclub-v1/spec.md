Status: ready-for-agent

# Spec: Boxclub Mitgliederverwaltung v1

## Problem Statement

Der Boxclub verwaltet seine ~50–200 Mitglieder heute in einer Excel-Tabelle. Als alleiniger Admin muss der Nutzer regelmäßig prüfen, welche Mitglieder ihren Beitrag bezahlt haben und welche nicht — der Zahlungsstatus wird separat über die macOS-App **MoneyMoney** eingesehen und heute manuell in Excel gepflegt. Die Excel-Lösung ist fehleranfällig (versehentliche Sortierung, überschriebene Zellen, keine Historie), unpraktisch zum Suchen/Filtern und lässt sich später schlecht um automatisches Matching mit Kontoumsätzen erweitern.

## Solution

Eine lokale Desktop-App auf Basis von **Go + Wails**, die die Excel-Tabelle ablöst. Die App verwaltet Stammdaten aller **Mitglieder**, ihre Zuordnung zu einer **Beitragsklasse** und ein Datum `bezahlt_bis`, bis zu dem der Beitrag als gezahlt gilt. In v1 wird `bezahlt_bis` **manuell** gepflegt — MoneyMoney bleibt komplett separat. Die App bietet eine übersichtliche Liste mit Suche und Filter, macht den Zahlungsstatus (bezahlt/nicht bezahlt) auf einen Blick sichtbar und importiert einmalig die bestehenden Excel-Daten.

Datenhaltung ist eine lokale SQLite-Datei auf dem Mac des Nutzers, geschützt durch FileVault. Kein Cloud-Sync, kein Multi-User, kein Login.

## User Stories

1. Als Vereinsadmin möchte ich eine Liste aller **Mitglieder** sehen, damit ich einen Überblick über den aktuellen Mitgliederbestand habe.
2. Als Vereinsadmin möchte ich in der Mitgliederliste nach Namen suchen, damit ich ein bestimmtes Mitglied schnell finde.
3. Als Vereinsadmin möchte ich die Mitgliederliste nach Zahlungsstatus filtern (bezahlt / nicht bezahlt), damit ich sehe, wen ich erinnern muss.
4. Als Vereinsadmin möchte ich die Mitgliederliste nach **Beitragsklasse** filtern, damit ich z. B. alle "Erwachsen 2×/Woche" auf einmal sehe.
5. Als Vereinsadmin möchte ich pro Mitglied den Zahlungsstatus an einer klaren visuellen Markierung (grün/rot) erkennen, damit ich nicht jeden Datumswert selbst interpretieren muss.
6. Als Vereinsadmin möchte ich ein neues **Mitglied** anlegen können mit Vor-/Nachname, Adresse, Geburtsdatum, Kontakt (E-Mail und/oder Telefon), Eintrittsdatum und Beitragsklasse, damit ich Neuzugänge sofort erfassen kann.
7. Als Vereinsadmin möchte ich die Stammdaten eines bestehenden Mitglieds bearbeiten können, damit Adressänderungen oder Wechsel der Beitragsklasse gepflegt werden können.
8. Als Vereinsadmin möchte ich für ein Mitglied das Datum `bezahlt_bis` manuell setzen können, damit ich den Zahlungsstand nach einem Blick in MoneyMoney synchronisieren kann.
9. Als Vereinsadmin möchte ich ein Mitglied als ausgetreten markieren können (mit Austrittsdatum), damit ehemalige Mitglieder aus der aktiven Liste verschwinden, aber der historische Datensatz erhalten bleibt.
10. Als Vereinsadmin möchte ich ein ausgetretenes Mitglied wieder eintreten lassen können, damit bei Wiedereintritt derselbe **Mitglied**-Datensatz weiterverwendet wird (neue **Mitgliedschaft**).
11. Als Vereinsadmin möchte ich einmalig meine bestehende Excel-Tabelle importieren können, damit ich nicht alle Mitglieder händisch anlegen muss.
12. Als Vereinsadmin möchte ich beim Excel-Import einen Bericht sehen, welche Zeilen erfolgreich importiert wurden und welche warum nicht (fehlender Name, unbekannte Beitragsklasse, ungültiges Datum), damit ich die Fehler in Excel korrigieren und erneut importieren kann.
13. Als Vereinsadmin möchte ich die verfügbaren Beitragsklassen und ihre Preise sehen und (später) pflegen können, damit Preiserhöhungen abbildbar sind.
14. Als Vereinsadmin möchte ich, dass die App beim Start ohne bestehende Datenbank die SQLite-Datei automatisch anlegt und mit den Standard-Beitragsklassen (`Erwachsen 1×/Woche` = 60 €, `Erwachsen 2×/Woche` = 80 €) initialisiert wird, damit ich sofort loslegen kann.
15. Als Vereinsadmin möchte ich, dass meine Mitgliederdaten die Festplatte nicht verlassen, damit die DSGVO-Situation überschaubar bleibt.
16. Als Vereinsadmin möchte ich, dass die App produktiv auf macOS läuft — auf demselben Rechner, auf dem auch MoneyMoney installiert ist — damit ich Zahlungsstatus und Mitgliederdaten am selben Gerät pflegen kann.
17. Als Vereinsadmin möchte ich die App per Suchbegriff finden, der auch in E-Mail-Adresse oder Telefonnummer vorkommt (nicht nur im Namen), damit ich anhand einer eingegangenen Zahlung schnell das passende Mitglied identifizieren kann.
18. Als Vereinsadmin möchte ich eine "aktive Mitglieder"-Standardansicht haben (nur nicht-ausgetretene), damit die Liste nicht mit Ex-Mitgliedern zugewuchert ist, und trotzdem die Möglichkeit haben, ehemalige einzublenden.
19. Als Entwickler möchte ich die App auf meinen Windows- und Linux-Notebooks per `wails dev` starten und per `wails build` als natives Binary bauen können, damit ich ohne täglichen Mac-Zugriff iterieren und Änderungen testen kann. Die dabei erzeugten Windows-/Linux-Binaries sind Entwicklungs-Artefakte und keine Auslieferungsziele.

## Implementation Decisions

### Stack & Deployment

- Sprache: **Go**. UI-Framework: **Wails v2**. Frontend im nativen WebView: **Vanilla HTML + htmx + Tailwind CSS**. Kein SPA-Framework, kein npm-Build außer dem von Wails eingebautem Schritt. Begründung: siehe [ADR-0001](../../docs/adr/0001-go-wails-fuer-desktop-gui.md).
- Datenbank: **SQLite**, Datei `boxclub.db` neben der Anwendung. Treiber: **`modernc.org/sqlite`** (pure-Go, kein CGo, damit Cross-Compilation trivial bleibt).
- **Ziel-Plattform in Produktion**: ausschließlich **macOS** — der Rechner, auf dem MoneyMoney läuft und der Vereinsadmin die App benutzt.
- **Entwicklungs-Plattformen**: **Windows + Linux**. `wails dev` und `wails build` müssen auf beiden nativ funktionieren, damit tägliche Iteration ohne Mac-Zugriff möglich ist. Voraussetzungen: unter Windows die WebView2-Runtime (auf Win10/11 Standard), unter Linux `libwebkit2gtk-4.0-dev` (bzw. distro-Äquivalent).
- **Mac-Release-Build**: auf einem realen Mac gebaut, **nicht** cross-kompiliert von Windows/Linux aus — Wails' native WebView-Bindung unterscheidet sich pro Plattform, und Cross-Compilation nach Darwin ist mit Wails praktisch fragil. In v1 pragmatisch entweder auf einem gelegentlich zugänglichen Mac oder über einen `macos-latest`-Runner auf GitHub Actions; die konkrete Wahl fällt im Release-Ticket.

### Module

Zwei primäre Go-Packages, jeweils an einem Seam:

- **`service`** — enthält den `MemberService`. Bündelt **die gesamte Business-Logik**: CRUD auf Mitgliedern, Zuordnung zu Beitragsklassen, Suche/Filter, `bezahlt_bis`-Pflege, Aus-/Wiedereintritt. Nutzt `database/sql` direkt gegen SQLite — kein Repository-Interface (siehe "Was **nicht** modelliert wird" weiter unten).
- **`importer`** — enthält den `ExcelImporter`. Nimmt einen `io.Reader` einer `.xlsx`-Datei entgegen, liefert eine geparste Zeilenliste plus Fehlerbericht. **Ruft den `MemberService` nicht selbst auf**; ein dünner Orchestrator im Wails-Layer setzt die geparsten Rows über `MemberService.Create` ein. Der Split entkoppelt Format-Parsing von Business-Regeln und legt das Muster für den v2-MoneyMoney-CSV-Importer vor.

Adapter-Schichten (nicht Teil der Seams):

- **`app`** / Wails-Bindings — dünne Methoden, die Frontend-Aufrufe an `MemberService` bzw. `ExcelImporter` weiterleiten und Ergebnisse als JSON zurückgeben (bzw. bei htmx als HTML-Fragmente).
- **`templates`** — Go-`html/template`-Dateien, die htmx-Fragmente rendern (Mitgliederzeilen, Formulare, Fehlerlisten).

### Schema

Drei Tabellen in SQLite:

- `beitragsklasse (id, name, preis_monatlich_cents, aktiv)` — bei Erstinbetriebnahme mit den beiden aktuellen Klassen (60 € / 80 €) geseedet. `preis_monatlich_cents` in Cent, um Fließkomma-Ungenauigkeit zu vermeiden.
- `mitglied (id, vorname, nachname, geburtsdatum, adresse, email, telefon, beitragsklasse_id, bezahlt_bis)` — Personen-Stammdaten. Ein Mitglied bleibt derselbe Datensatz auch bei Wiedereintritt.
- `mitgliedschaft (id, mitglied_id, eintritt, austritt NULL)` — Zeitraum-Zuordnung. Ein Mitglied kann mehrere Zeilen haben (Aus- und Wiedereintritt). "Aktives Mitglied" heißt: es existiert eine `mitgliedschaft` mit `austritt IS NULL`.

Trennung `mitglied` ↔ `mitgliedschaft` ist bewusst; sie folgt direkt aus dem Vokabular in [CONTEXT.md](../../CONTEXT.md) und erlaubt sauberen Wiedereintritt ohne Datenduplikat.

### Zahlungsstatus-Ableitung

Der Zahlungsstatus wird **nicht gespeichert**, sondern jedes Mal aus `bezahlt_bis` und `heute` abgeleitet:

- `bezahlt` wenn `bezahlt_bis >= heute`
- `nicht bezahlt` sonst

Bewusst zweistufig, keine gelbe Vorwarnstufe (Entscheidung aus Runde 3).

### Suche und Filter

`MemberService.Search(query, filter)` implementiert Suche über SQL `LIKE` gegen Vorname, Nachname, E-Mail, Telefon (Story 17). Der `filter`-Parameter kombiniert:

- Zahlungsstatus: alle / nur bezahlt / nur nicht bezahlt
- Beitragsklasse: alle / bestimmte
- Aktivitätsstatus: nur aktive Mitglieder (Standard) / alle inkl. ausgetretene

### Excel-Import

- Eingabe: `.xlsx` (nicht `.xls`), gelesen via `github.com/xuri/excelize/v2`.
- Konkretes **Spalten-Mapping ist bewusst offen**: der Nutzer muss die Spaltennamen seiner heutigen Tabelle noch liefern. Bis dahin bleibt `ExcelImporter` ein Skelett mit klarem Konfigurationspunkt (ein `ColumnMapping`-Struct), das nach Nachlieferung mit konkreten Spaltennamen befüllt wird.
- Import ist **wiederholbar**, nicht einmalig: ein zweiter Import derselben Zeile aktualisiert das bestehende Mitglied (Match über E-Mail oder Vor-/Nachname + Geburtsdatum, exakte Regel wird beim Mapping-Task festgezurrt).

### Wails-API-Kontrakt (Frontend ↔ Backend)

Wails-Bindings exponieren pro User-Story eine Methode und liefern HTML-Fragmente zurück (htmx-Muster), z. B.:

- `ListMembers(filter) → HTML-Fragment einer Mitgliederliste`
- `MemberForm(id?) → HTML-Fragment eines Bearbeitungsformulars`
- `SaveMember(payload) → HTML-Fragment der aktualisierten Zeile`
- `SetBezahltBis(id, datum) → HTML-Fragment der aktualisierten Zelle`
- `ImportExcel(reader) → HTML-Fragment eines Import-Berichts`

Diese Bindings sind **1:1 dünne Wrapper** über `MemberService`- bzw. `ExcelImporter`-Aufrufe.

## Testing Decisions

### Was macht einen guten Test aus

Getestet wird **externes Verhalten am Seam**, nicht interne Struktur:

- `MemberService`-Tests rufen Service-Methoden auf und beobachten Rückgabewerte und Datenbankzustand über weitere Service-Methoden. Sie greifen **nicht** direkt auf Tabellen zu.
- `ExcelImporter`-Tests reichen eine Fixture-Datei rein und prüfen die zurückgegebenen Rows und den Fehlerbericht.
- Kein Mocking der Datenbank. Ein echtes SQLite in einer Temp-Datei (`t.TempDir()`) wird pro Test frisch angelegt.
- Kein Test bindet an konkrete SQL-Statements oder Struct-Feldnamen jenseits der Service-API.

### Getestete Module

- **`service`**: alle User Stories 1–10, 13, 14, 17, 18 sind direkt Tests am `MemberService`-Seam.
- **`importer`**: User Stories 11, 12 sind Tests am `ExcelImporter`-Seam.

### Adapter, die nicht dediziert getestet werden

- Wails-Bindings: sind so dünn, dass die Service-/Importer-Tests sie vollständig abdecken.
- htmx-Templates: gerendertes HTML wird nicht assertiert; ein einzelner manueller Rauchtest beim Start reicht in v1.

### Prior art

Keine — der Code ist greenfield. Alle Testschemata werden vom ersten Ticket etabliert (`service/member_service_test.go` als Referenz-Testdatei; alle späteren Tests spiegeln deren Setup-Muster).

### Multi-Plattform-Rauchtest

Nach jeder größeren Änderung muss `wails dev` und `wails build` sowohl unter Windows als auch unter Linux ohne Fehler durchlaufen (manueller Test auf beiden Dev-Notebooks; nicht automatisierbar in v1). Der Mac-Build wird nur zum Release-Zeitpunkt geprüft, nicht pro Change. Ein Fehlschlag auf einer Dev-Plattform ist so kritisch wie ein roter Unit-Test — er blockiert das Merge des betroffenen Tickets.

## Out of Scope

Alles, was nicht ausdrücklich in "User Stories" steht, insbesondere:

- **Automatischer MoneyMoney-Abgleich** (CSV-Import, Umsatz-Matching, automatische `bezahlt_bis`-Fortschreibung). Bewusst v2. Das Datenmodell ist so gebaut, dass v2 additiv andocken kann.
- **Anwesenheitsverfolgung** (wer war wann im Training).
- **Boxspezifisches**: Wettkampflizenz, Startbuch, Gewichtsklasse, medizinische Freigabe.
- **Kommunikation aus der App heraus** (Serienbrief, E-Mail-Versand, SEPA-Export).
- **Mehrbenutzer, Login, Rollen**. Genau ein Admin.
- **Web-, Server- oder Cloud-Deployment**. Rein lokal.
- **Alter- oder Ermäßigungs-Klassen**. Der Verein hat bewusst nur die zwei frequenzbasierten Klassen (Runde 4 confirmed).
- **Windows- und Linux-Auslieferungsbuilds**. Auf beiden Plattformen wird während der Entwicklung gebaut und gestartet, aber sie sind kein Produktions-Ziel: keine Installer, keine Signierung, kein Support-Versprechen, keine Aufnahme in ein Release.
- **UI-Testautomatisierung** (Playwright o. ä.).
- **Backups**. FileVault + Time Machine ist die externe Antwort; die App selbst legt keine Backups an.

## Further Notes

### Bewusst nicht modelliert

Die Grenze zwischen "sauber genug" und "overengineered" für ein Solo-Projekt ist eng gezogen:

- **Kein Repository-Interface**: `MemberService` spricht direkt zu `database/sql`. Ein Repository-Abstraktions-Layer wäre bei genau einer Backend-Wahl und einem Solo-Entwickler unnötige Zeremonie.
- **Kein separates Domain-vs-Persistence-Modell**: dieselben Structs werden für DB und Service-API verwendet.
- **Kein Event-Log / kein CQRS**: für 200 Mitglieder unnötig. Die einzige Historie, die v1 vorhält, ist die `mitgliedschaft`-Tabelle.

### Offene Fakten (nicht blockierend für Ticket-Erstellung)

- **Excel-Spaltennamen**: Nutzer reicht nach. Blockiert das Excel-Import-Ticket, nicht die anderen.

### Verweise

- Domain-Vokabular: [CONTEXT.md](../../CONTEXT.md)
- Stack-Entscheidung: [docs/adr/0001-go-wails-fuer-desktop-gui.md](../../docs/adr/0001-go-wails-fuer-desktop-gui.md)
- Issue-Tracker-Konvention: [docs/agents/issue-tracker.md](../../docs/agents/issue-tracker.md)
