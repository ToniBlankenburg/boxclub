Status: ready-for-agent

# Spec: Boxclub Mitgliederverwaltung v1

> **Stand nach [ADR-0005](../../docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md)
> und [ADR-0006](../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md).** Die erste
> Fassung dieser Spec ging von zwei frequenzbasierten Beitragsklassen und von
> überweisenden Mitgliedern mit einem gepflegten `bezahlt_bis`-Datum aus. Beides
> hat die Bestandsaufnahme der echten Excel-Tabelle
> ([excel-vorlage.md](excel-vorlage.md)) widerlegt. Die betroffenen Stories und
> Abschnitte sind **umgeschrieben**, nicht nachträglich kommentiert; die Historie
> des Bruchs tragen die beiden ADRs.

## Problem Statement

Der Boxclub verwaltet seine ~50–200 Mitglieder heute in einer Excel-Tabelle mit
21 Spalten. Der Vereinsadmin ist alleiniger Nutzer.

Die Beiträge zieht der Verein per **Lastschrift** ein — Mitglieder überweisen
nicht. Damit ist die Frage "wer hat bezahlt?" im Normalfall beantwortet, und die
Tabelle enthält konsequenterweise auch keine Spalte dafür. Was wirklich weh tut:

- **Die Tabelle ist fragil.** Eine versehentliche Sortierung, eine überschriebene
  Zelle, ein verrutschter Bereich — und der Schaden ist unsichtbar und nicht
  rückholbar. Bei 21 Spalten und 200 Zeilen passiert das.
- **Suchen und Filtern ist unpraktisch.** "Wer trainiert samstags und ist in
  Kündigungsfrist" ist in Excel Handarbeit mit Autofiltern.
- **Für Rücklastschriften gibt es keinen Ort.** Kommt eine Lastschrift zurück,
  merkt sich der Admin das im Kopf oder per Notizzettel. In der Tabelle steht
  nichts, und beim nächsten Einzug ist unklar, was noch offen ist.
- **Wiedereintritte zerstören Historie.** Tritt jemand aus und Jahre später wieder
  ein, wird die Zeile überschrieben. Was vorher galt — Beitrag, Trainingszeit,
  Eintrittsdatum — ist weg.
- **Die Daten sind handgepflegt und entsprechend schmutzig:** Tippfehler in
  Dropdown-Werten (`Mitwoch`, `Inakiv`), Datumswerte teils als Excel-Seriennummer,
  teils vermutlich getippt, Spaltenüberschriften mit Tippfehlern (`Eintrit`).

## Solution

Eine lokale Desktop-App auf Basis von **Go + Wails**, die die Excel-Tabelle
vollständig ablöst — vollständig im wörtlichen Sinn: **jede der 21 Spalten
bekommt einen Ort in der App.** Was die App nicht halten kann, hält die Excel
weiter, und dann laufen zwei Systeme parallel statt einem.

Die App verwaltet die Stammdaten aller **Mitglieder**, ihre **Mitgliedschaften**
als zeitlich begrenzte Zeiträume, den individuell vereinbarten **Beitrag** und
die **Trainingsslots**. Sie zeigt eine Liste mit Suche und Filter, liest den
**Status** eines Mitglieds aus den Datumsfeldern ab, statt ihn pflegen zu lassen,
und bietet mit dem **Rückstands**-Kennzeichen erstmals einen Platz für
Rücklastschriften samt Notiz zum Vorgang. Die bestehende Excel-Tabelle wird
importiert, wiederholbar und mit Fehlerbericht.

Datenhaltung ist eine lokale SQLite-Datei auf dem Mac des Nutzers, geschützt
durch FileVault. Kein Cloud-Sync, kein Multi-User, kein Login.

Das Vokabular dieser Spec ist verbindlich definiert in [CONTEXT.md](../../CONTEXT.md).

## User Stories

### Überblick und Liste

1. Als Vereinsadmin möchte ich eine Liste aller **Mitglieder** sehen, damit ich einen Überblick über den Bestand habe.
2. Als Vereinsadmin möchte ich standardmäßig nur **aktive** Mitglieder sehen, damit die Liste nicht mit Ausgetretenen zugewuchert ist.
3. Als Vereinsadmin möchte ich ausgetretene Mitglieder auf Wunsch einblenden können, damit ich in der Historie nachsehen kann.
4. Als Vereinsadmin möchte ich nach Vorname, Nachname, E-Mail und Telefonnummer suchen, damit ich ein Mitglied schnell finde — auch anhand einer Kontaktangabe aus einer eingegangenen Nachricht.
5. Als Vereinsadmin möchte ich auch nach der **Mitglieds-ID** suchen, damit ich die Nummer, unter der ich jemanden kenne, direkt eingeben kann.
6. Als Vereinsadmin möchte ich die Liste nach **Rückstand** filtern, damit ich die Mitglieder sehe, bei denen eine Lastschrift geplatzt ist und Geld offen ist.
7. Als Vereinsadmin möchte ich die Liste nach **Trainingsfrequenz** filtern (1×, 2×, 3× pro Woche), damit ich z. B. alle 2×-Trainierenden auf einmal sehe.
8. Als Vereinsadmin möchte ich pro Zeile am **Rückstands-Kennzeichen** (grün/rot) sofort erkennen, ob bei diesem Mitglied etwas offen ist, ohne ein Datum interpretieren zu müssen.
9. Als Vereinsadmin möchte ich pro Zeile den **Status** des Mitglieds ablesen (Neu, Aktiv, In Kündigungsfrist, Ausgetreten) und erkennen, ob es **ruhend** ist, damit ich den Zustand ohne Rechnen erfasse.
10. Als Vereinsadmin möchte ich die Liste alphabetisch nach Nachnamen sortiert sehen, mit korrekter Behandlung deutscher Umlaute, damit ich nicht scrollen muss, um jemanden zu finden.

### Stammdaten pflegen

11. Als Vereinsadmin möchte ich ein neues **Mitglied** anlegen können mit Vorname, Nachname, Geburtsdatum, Adresse, Postleitzahl, Ort, E-Mail, Telefon, IBAN und Geschlecht, damit ich Neuzugänge sofort vollständig erfasse.
12. Als Vereinsadmin möchte ich beim Anlegen Anmeldedatum, Eintrittsdatum, Anmeldegebühr, Beitrag und Trainingsslots angeben können, damit die erste **Mitgliedschaft** direkt vollständig ist.
13. Als Vereinsadmin möchte ich die Stammdaten eines bestehenden Mitglieds bearbeiten können, damit Adress-, Kontakt- oder Bankänderungen gepflegt werden.
14. Als Vereinsadmin möchte ich den **Beitrag** eines Mitglieds individuell festlegen und ändern können, damit ich abweichende Vereinbarungen (Altbestand, Härtefall, Trainer mit 0 €) abbilden kann.
15. Als Vereinsadmin möchte ich einem Mitglied **null bis drei Trainingsslots** zuordnen können, damit festgehalten ist, an welchen wöchentlichen Terminen es trainiert.
16. Als Vereinsadmin möchte ich, dass sich die **Trainingsfrequenz** aus der Anzahl der Slots ergibt, damit Frequenz und Slots nicht auseinanderlaufen können.
17. Als Vereinsadmin möchte ich die **IBAN** eines Mitglieds erfassen und ändern können, damit die Lastschrift-Daten in der App liegen und nicht in einer zweiten Tabelle.
18. Als Vereinsadmin möchte ich festhalten können, ob ein Mitglied den Verein bei **Google bewertet** hat, damit ich weiß, wen ich noch fragen kann.

### Lebenszyklus

19. Als Vereinsadmin möchte ich eine **Kündigung** erfassen können — Kündigungsdatum und das Datum, zu dem der Austritt wirksam wird — damit die Kündigungsfrist abgebildet ist.
20. Als Vereinsadmin möchte ich, dass ein Mitglied **in der Kündigungsfrist als aktiv gilt** und in der Standardansicht bleibt, weil es weiter trainiert und weiter zahlt.
21. Als Vereinsadmin möchte ich ein Mitglied als **ausgetreten** sehen, sobald sein Austrittsdatum erreicht ist, damit die aktive Liste sich selbst bereinigt, ohne dass ich etwas tun muss.
22. Als Vereinsadmin möchte ich ein ausgetretenes Mitglied **wieder eintreten** lassen können, sodass eine **neue Mitgliedschaft** entsteht und derselbe Mitglieds-Datensatz mit derselben Mitglieds-ID weiterläuft.
23. Als Vereinsadmin möchte ich, dass bei einem Wiedereintritt der **alte Beitrag und die alten Trainingsslots erhalten bleiben** und die neue Mitgliedschaft eigene bekommt, damit die Historie nicht überschrieben wird.
24. Als Vereinsadmin möchte ich ein Mitglied als **ruhend** markieren können, wenn es vorübergehend nicht trainiert (Verletzung, Ausland), damit klar ist, dass für diese Zeit kein Beitrag eingezogen wird und es trotzdem Mitglied bleibt.
25. Als Vereinsadmin möchte ich ein ruhendes Mitglied wieder aktiv setzen können, damit der Einzug wieder läuft.

### Rücklastschriften

26. Als Vereinsadmin möchte ich ein Mitglied als **im Rückstand** markieren können, wenn seine Lastschrift zurückgekommen ist, damit ich die offene Forderung nicht im Kopf behalten muss.
27. Als Vereinsadmin möchte ich zum Rückstand eine **freie Notiz** hinterlegen können ("Rücklastschrift Oktober, angeschrieben am 05.10."), damit der Stand des Vorgangs dokumentiert ist.
28. Als Vereinsadmin möchte ich den Rückstand wieder aufheben können, sobald das Geld da ist, damit die rote Liste wieder kurz ist.
29. Als Vereinsadmin möchte ich, dass der Rückstand am **Mitglied** hängt und einen Austritt überlebt, weil eine offene Forderung nicht verschwindet, wenn jemand geht.

### Excel-Import

30. Als Vereinsadmin möchte ich meine bestehende `.xlsx`-Tabelle importieren können, damit ich nicht 200 Mitglieder händisch anlege.
31. Als Vereinsadmin möchte ich beim Import einen **Bericht** sehen: wie viele Zeilen erfolgreich waren, wie viele nicht, und jede gescheiterte Zeile einzeln mit Grund, damit ich die Fehler in Excel korrigieren und erneut importieren kann.
32. Als Vereinsadmin möchte ich den Import **wiederholen** können, sodass ein zweiter Lauf bestehende Mitglieder aktualisiert statt sie zu duplizieren.
33. Als Vereinsadmin möchte ich, dass die **Mitglieds-IDs aus der Excel übernommen** werden, damit die Nummern, unter denen ich meine Mitglieder kenne, dieselben bleiben und ich beide Systeme während der Umstellung vergleichen kann.
34. Als Vereinsadmin möchte ich, dass der Import Datumswerte sowohl als **Excel-Seriennummer** als auch als getippten Text liest, weil in meiner Tabelle beides vorkommt.
35. Als Vereinsadmin möchte ich, dass der Import **keine stillen Annahmen** trifft: was er nicht zuordnen kann, erscheint im Bericht, statt mit einem Standardwert überschrieben zu werden.

### Betrieb

36. Als Vereinsadmin möchte ich, dass die App beim ersten Start ohne bestehende Datenbank die SQLite-Datei selbst anlegt, damit ich sofort loslegen kann.
37. Als Vereinsadmin möchte ich, dass meine Mitgliederdaten die Festplatte nicht verlassen, damit die DSGVO-Situation überschaubar bleibt.
38. Als Vereinsadmin möchte ich, dass die App produktiv auf macOS läuft — auf demselben Rechner wie MoneyMoney.
39. Als Entwickler möchte ich die App auf meinen Windows- und Linux-Notebooks per `wails dev` starten und per `wails build` bauen können, damit ich ohne täglichen Mac-Zugriff iterieren kann. Diese Binaries sind Entwicklungs-Artefakte, keine Auslieferungsziele.

## Implementation Decisions

### Stack & Deployment

Unverändert gegenüber der ersten Fassung:

- Sprache **Go**, UI **Wails v2**, Frontend **Vanilla HTML + htmx + Tailwind** im nativen WebView. Kein SPA-Framework, kein eigener npm-Build. Begründung: [ADR-0001](../../docs/adr/0001-go-wails-fuer-desktop-gui.md).
- Transport: htmx spricht per HTTP gegen `/api/...`-Routen am Wails-Assetserver. **Wails-Bindings werden nicht verwendet** — [ADR-0002](../../docs/adr/0002-htmx-ueber-den-wails-assetserver.md).
- **SQLite** über `modernc.org/sqlite` (pure Go, kein CGo, Cross-Compilation bleibt trivial). Datei im Benutzer-Konfigurationsordner — [ADR-0003](../../docs/adr/0003-datenbank-im-benutzer-konfigurationsordner.md). `BOXCLUB_DB` überschreibt den Pfad für die Entwicklung.
- Produktionsziel **ausschließlich macOS**. Entwicklung auf **Windows + Linux**, beide müssen `wails dev` und `wails build` nativ durchlaufen.
- Der Mac-Release wird auf einem echten Mac gebaut, nicht cross-kompiliert. Die Wahl zwischen gelegentlich zugänglichem Mac und `macos-latest`-Runner fällt im Release-Ticket.

### Module und Seams

Zwei primäre Packages, jeweils an einem Seam — **es kommt kein dritter dazu**:

- **`service`** — der `MemberService`. Bündelt die gesamte Business-Logik: CRUD auf Mitgliedern und Mitgliedschaften, Trainingsslots, Beitrag, Rückstand, Lebenszyklus, Suche/Filter und alle abgeleiteten Werte. Spricht `database/sql` direkt.
- **`importer`** — der `ExcelImporter`. Nimmt einen `io.Reader` einer `.xlsx`-Datei, liefert geparste Zeilen plus Fehlerbericht. Kennt das **Format** (Spaltennamen, Excel-Seriendaten), nicht die **Regeln**. Ruft den `MemberService` nicht selbst auf; ein dünner Orchestrator im Wails-Layer setzt die Zeilen über den Service ein.

Ausdrücklich **verworfen**: ein eigener `MitgliedschaftService` für Mitgliedschaften und Slots. Mitglied, Mitgliedschaft und Trainingsslot sind ein Aggregat mit einer Transaktionsgrenze; ein Seam mitten hindurch wäre Zeremonie ohne Gegenwert.

Adapter (nicht Teil der Seams): **`app`** — dünne HTTP-Handler, die HTML-Fragmente liefern; **`templates`** — Go-`html/template`-Dateien für Zeilen, Formulare und Fehlerlisten.

### Schema

Vier Tabellen. Die Aufteilung folgt der Frage "wer ist die Person" gegen "was gilt in diesem Zeitraum":

**`mitglied`** — die dauerhaften Eigenschaften der Person:
`id` (die **Mitglieds-ID**, eindeutig, automatisch hochzählend, beim Import aus der Excel übernommen), `vorname`, `nachname`, `geburtsdatum`, `adresse`, `postleitzahl`, `ort`, `email`, `telefon`, `iban`, `geschlecht`, `google_bewertung`, `digital`, `rueckstand`, `rueckstand_notiz`

**`mitgliedschaft`** — was in einem Zeitraum vereinbart ist, mehrere Zeilen pro Mitglied möglich:
`id`, `mitglied_id`, `anmeldedatum`, `eintritt`, `austritt` (NULL = läuft), `kuendigungsdatum`, `anmeldegebuehr_cents`, `beitrag_monatlich_cents`, `ruhend`

**`trainingsslot`** — null bis drei Zeilen pro Mitgliedschaft:
`id`, `mitgliedschaft_id`, `bezeichnung` (Wochentag + Uhrzeit als Text)

Geldbeträge durchgängig in **Cent**, um Fließkomma-Ungenauigkeit zu vermeiden.

**Entfallen** gegenüber der ersten Fassung:

- die Tabelle **`beitragsklasse`** und `mitglied.beitragsklasse_id` — [ADR-0005](../../docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md). Damit entfällt auch das Seeden beim ersten Start: eine frische Datenbank ist einfach leer.
- **`mitglied.bezahlt_bis`** — [ADR-0006](../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md).

Der Rückstand hängt am **Mitglied**, weil eine offene Forderung einen Austritt überlebt. `ruhend` hängt an der **Mitgliedschaft**, weil man nur eine laufende Mitgliedschaft pausieren kann. Beide Felder sind unabhängig: für ein ruhendes Mitglied wird keine Lastschrift losgeschickt, es kann also kein neuer Rückstand entstehen — ein alter bleibt aber stehen.

**Es gibt keinen Migrationsmechanismus.** Das Schema entsteht über `CREATE TABLE IF NOT EXISTS`; eine Schemaänderung bedeutet, die Entwicklungs-Datenbank zu löschen. Bis zum ersten Mac-Release ist das der bewusst gewählte Preis.

### Abgeleitete Werte

Nichts davon wird gespeichert. Alle drei werden bei jeder Abfrage neu bestimmt — dasselbe Muster wie bisher beim Zahlungsstatus:

- **Status** aus den Datumsfeldern der laufenden Mitgliedschaft: *Neu* (Eintritt liegt in der Zukunft), *Aktiv* (Eintritt erreicht, kein Austritt), *In Kündigungsfrist* (Kündigungsdatum gesetzt, Austritt liegt in der Zukunft), *Ausgetreten* (Austritt erreicht). Die Excel-Werte `Mitglied` und `Aktiv` sind derselbe Zustand; `Stillgelegt` und `Inakiv` sind beide *ruhend* und damit das einzige gespeicherte Merkmal in dieser Reihe.
- **Trainingsfrequenz** aus der Anzahl der Trainingsslots der laufenden Mitgliedschaft.
- **Rückstands-Kennzeichen** ist zweiwertig — *in Ordnung* oder *im Rückstand*. Der dritte, neutrale Zustand *nicht gesetzt* aus der ersten Fassung **entfällt**: er existierte nur, weil ein leeres Datum nicht als "nicht bezahlt" gelten durfte. Ein Kennzeichen mit zwei Werten hat das Problem nicht.

**"Aktiv" ändert seine Definition.** Bisher: `austritt IS NULL`. Künftig gilt ein Mitglied als aktiv, solange der Austritt nicht erreicht ist — ein Mitglied in der Kündigungsfrist hat ein gesetztes Austrittsdatum und ist trotzdem aktiv. Das zieht den Aktivitätsfilter nach.

### Suche und Filter

`MemberService.Search(query, filter)` sucht gegen Vorname, Nachname, E-Mail, Telefon und **Mitglieds-ID**. Die vier Textfelder treffen als Teilzeichenkette; die **Mitglieds-ID trifft genau**. Als Teiltreffer wäre sie bei 200 Mitgliedern wertlos — "7" brächte die 7, die 17, die 27 und die 70er zurück, und damit wäre die Nummer als Sprungmarke zu einer bekannten Zeile gerade nicht mehr zu gebrauchen (Story 5). Der Filter kombiniert:

- Rückstand: alle / nur im Rückstand / nur in Ordnung
- Trainingsfrequenz: alle / 1× / 2× / 3×
- Aktivität: nur aktive (Standard) / alle inkl. ausgetretene

Suchbegriff und die wertbasierten Filter werden **in Go** ausgewertet, nicht in SQL — [ADR-0004](../../docs/adr/0004-suche-und-filter-im-speicher.md) (SQLite faltet Groß-/Kleinschreibung nur im ASCII-Bereich, "öztürk" fände "Öztürk" nicht). In SQL bleibt allein die Aktivität, weil sie entscheidet, welche Mitgliedschaft die Zeile überhaupt bekommt. Der Klassenfilter aus der ersten Fassung wird durch den Frequenzfilter ersetzt; da die Frequenz aus der Anzahl der Slots abgeleitet wird, gehört er ebenfalls in den Go-Teil.

Sortierung weiterhin in Go über `collate.German`, damit Umlaute richtig einsortiert werden.

### Excel-Import

- Eingabe `.xlsx` (nicht `.xls`), gelesen über `github.com/xuri/excelize/v2`.
- Quelle der Spaltennamen und Wertelisten: [excel-vorlage.md](excel-vorlage.md). Die Spaltenliste ist vom Nutzer als **echt und vollständig** bestätigt.
- Mapping der 21 Spalten (Kopfzeile in Zeile 1, Daten ab Zeile 2):

| Excel-Spalte | Ziel |
|---|---|
| `Mandatsreferenz` | `mitglied.id` — **trotz des Namens keine SEPA-Referenz**, sondern die Mitglieds-ID |
| `Vorname`, `Nachname`, `Geburtstag` | Mitglied, gleichnamig |
| `Adresse`, `Postleitzahl`, `Ort` | Mitglied, drei getrennte Felder |
| `Telefonnummer`, `E-Mail`, `IBAN`, `Geschlecht` | Mitglied, gleichnamig |
| `Bewertung` | `mitglied.google_bewertung` (ja/nein) |
| `Digital` | `mitglied.digital` — **wortwörtlich als Freitext**, Bedeutung offen |
| `Beitrag` | `mitgliedschaft.beitrag_monatlich_cents` (Euro-Zahl × 100) |
| `Anmeldegebühr` | `mitgliedschaft.anmeldegebuehr_cents` |
| `Eintrit` *(sic)* | `mitgliedschaft.anmeldedatum` |
| `Mitgliedschaft` | `mitgliedschaft.eintritt` — der fachliche Beginn, üblicherweise ein Monatserster |
| `Status` | **Kein Feld, aber steuernd** — er entscheidet, wohin das Datum aus `Gekündigt` gehört: bei *Gekündigt* wird es der Austritt, bei *Kündigungsfrist* das Kündigungsdatum. *Stillgelegt*/*Inakiv* setzen `mitgliedschaft.ruhend`. *Neu*/*Aktiv*/*Mitglied* haben keine Wirkung — diese Zustände werden ohnehin abgelesen. Ein unbekannter Wert ist ein Eintrag im Fehlerbericht |
| `Gekündigt` | `mitgliedschaft.austritt` **oder** `mitgliedschaft.kuendigungsdatum`, je nach `Status` — siehe oben |
| `Training - 1/2/3` | je eine Zeile in `trainingsslot`, leere Zellen ergeben keine Zeile |
| `1x 2x Woche` | **keine Spalte** — die Frequenz ergibt sich aus der Anzahl der Slots. Weicht der Excel-Wert von der Slot-Anzahl ab, ist das ein Eintrag im Fehlerbericht, keine stille Korrektur |

- **Mitglieds-IDs werden übernommen.** Die automatische Vergabe zählt danach oberhalb der höchsten importierten Nummer weiter. Eine doppelte oder leere Nummer ist ein Fehlerfall im Bericht.
- **Datumswerte** kommen sowohl als Excel-Seriennummer (Epoche 1899-12-30) als auch als getippter Text vor. Beides muss gelesen werden; was keins von beidem ist, ist ein Fehlerfall.
- **Wiederholbarer Import:** Match über die Mitglieds-ID. Existiert sie, wird das Mitglied aktualisiert; existiert sie nicht, wird es angelegt. Rückfall auf Vorname + Nachname + Geburtsdatum für Zeilen ohne Nummer.
- **Der Rückstand wird nicht importiert** — es gibt keine Quellspalte. Jedes importierte Mitglied startet als *in Ordnung*, was bei Lastschrift der zutreffende Normalfall ist.
- **Kein stilles Ausweichen auf Standardwerte.** Fehlender Nachname, unlesbares Datum, nicht interpretierbarer Beitrag, doppelte ID: alles erscheint zeilenweise mit Grund im Bericht.

### Operationen des HTTP-Layers

Pro Story eine Route, die ein HTML-Fragment liefert (htmx-Muster) — dünne 1:1-Wrapper über `MemberService` bzw. `ExcelImporter`. Unter anderem: Mitgliederliste mit Filter, Bearbeitungsformular, Speichern einer Zeile, Trainingsslots zuordnen, Rückstand setzen und aufheben, Kündigung erfassen, Wiedereintritt, ruhend schalten, Excel-Import mit Bericht.

## Testing Decisions

### Was macht einen guten Test aus

Getestet wird **externes Verhalten am Seam**, nicht interne Struktur:

- `MemberService`-Tests rufen Service-Methoden auf und beobachten Rückgabewerte und Datenbankzustand über weitere Service-Methoden. Sie greifen **nicht** direkt auf Tabellen zu.
- `ExcelImporter`-Tests reichen eine Fixture-Datei rein und prüfen die zurückgegebenen Zeilen und den Fehlerbericht.
- **Kein Mocking der Datenbank.** Ein echtes SQLite in einer Temp-Datei (`t.TempDir()`) pro Test.
- Kein Test bindet an konkrete SQL-Statements oder an Struct-Felder jenseits der Service-API.

### Prior art

`service/member_service_test.go` ist die Referenz-Testdatei; alle weiteren Tests spiegeln deren Setup-Muster. Für die Themen dieser Fassung existieren schon Vorbilder im Repo: `service/lebenszyklus_test.go` (Aus- und Wiedereintritt) und `service/search_test.go` (Suche und Filterkombinationen).

### Besonders zu testende Verhaltensweisen

Zusätzlich zu den bestehenden Testthemen:

- **Abgeleiteter Status** an den Rändern: Eintritt heute, Austritt heute, Kündigungsdatum ohne Austritt, Austritt in der Zukunft (Kündigungsfrist → gilt als aktiv).
- **Trainingsfrequenz** aus 0, 1, 2 und 3 Slots.
- **Wiedereintritt** lässt Beitrag und Slots der alten Mitgliedschaft unberührt.
- **Rückstand** überlebt einen Austritt.
- **Importer:** Excel-Seriendatum und Textdatum ergeben dasselbe Datum; fehlender Nachname, doppelte Mitglieds-ID und widersprüchliche Frequenz landen im Bericht; ein zweiter Import derselben Datei erzeugt keine Duplikate und aktualisiert geänderte Werte.
- **Geldbeträge** in Cent über die gesamte Strecke, inklusive des Import-Sonderfalls `Beitrag` = 0.

### Adapter, die nicht dediziert getestet werden

Die HTTP-Handler und die htmx-Templates. Sie sind dünn genug, dass die Service- und Importer-Tests sie abdecken; gerendertes HTML wird nicht assertiert. Ein manueller Rauchtest beim Start reicht in v1.

### Multi-Plattform-Rauchtest

Nach jeder größeren Änderung müssen `wails dev` und `wails build` unter **Windows und Linux** fehlerfrei durchlaufen (manuell auf beiden Dev-Notebooks). Der Mac-Build wird nur zum Release geprüft. Ein Fehlschlag auf einer Dev-Plattform blockiert das Ticket wie ein roter Test.

## Out of Scope

- **Automatischer MoneyMoney-Abgleich** (CSV-Import, Umsatz-Matching, automatisches Setzen des Rückstands). Bewusst v2; das Datenmodell ist so gebaut, dass v2 additiv andockt.
- **Zahlungshistorie.** v1 hält den Rückstand als Kennzeichen samt Notiz, nicht als Folge einzelner Vorfälle. Wer im März platzt und im April zahlt, hinterlässt außer der Notiz keine Spur.
- **SEPA-Erzeugung.** Die App speichert IBANs, erzeugt aber keine Lastschrift-Dateien und validiert keine IBAN. Der Einzug passiert weiterhin außerhalb.
- **Eine gepflegte Gruppen-Tabelle** für Trainingszeiten und ein Filter "zeig mir alle im Samstag-Training". Trainingsslots sind Text; die Werteliste der Excel enthält Tippfehler und kombinierte Einträge und ist damit eine Eintipphilfe, kein sauberer Datensatz. Kommt, wenn der Bedarf konkret wird.
- **Automatische Berechnung des Austrittsdatums** aus dem Kündigungsdatum. Fristen haben Sonderfälle (Kulanz, Aufhebungsvertrag, Quartalsende); ein errechnetes Datum, das man überschreiben muss, ist lästiger als ein leeres Feld.
- **Massenoperationen** wie "alle 60-€-Beiträge auf 65 € erhöhen". Folgt aus ADR-0005: Beiträge sind individuell. Wenn der Bedarf kommt, ist es ein eigenes Feature.
- **Anwesenheitsverfolgung**, wer wann im Training war.
- **Boxspezifisches**: Wettkampflizenz, Startbuch, Gewichtsklasse, medizinische Freigabe.
- **Kommunikation aus der App heraus**: Serienbrief, E-Mail-Versand.
- **Mehrbenutzer, Login, Rollen.** Genau ein Admin.
- **Web-, Server- oder Cloud-Deployment.** Rein lokal.
- **Windows- und Linux-Auslieferungsbuilds.** Dev-Plattformen, kein Produktionsziel: keine Installer, keine Signierung, kein Support.
- **UI-Testautomatisierung.**
- **Backups.** FileVault und Time Machine sind die externe Antwort; die App legt keine an.

## Further Notes

### Bewusst nicht modelliert

- **Kein Repository-Interface:** `MemberService` spricht direkt zu `database/sql`. Bei einer Backend-Wahl und einem Solo-Entwickler wäre die Abstraktion Zeremonie.
- **Kein separates Domain-vs-Persistence-Modell:** dieselben Structs für DB und Service-API.
- **Kein Event-Log, kein CQRS.** Die einzige Historie in v1 ist die Folge der Mitgliedschaften.
- **Keine Wertelisten-Prüfung** für Beitrag, Geschlecht oder Trainingsslot. Die Excel-Dropdowns sind Eintipphilfen, keine Einschränkungen — und sie enthalten selbst Tippfehler.

### Offene Punkte

- **Bedeutung der Spalte `Digital`** ist unbekannt. Sie wird als Freitext übernommen, damit keine Daten verloren gehen, bekommt aber keinen Eintrag im Glossar und keine Semantik im Modell. Sobald die Bedeutung klar ist, ist das eine eigene, kleine Änderung.

### Was diese Fassung kostet

Drei bereits ausgelieferte Tickets werden zurückgebaut oder umgehängt:

- **Ticket 05** (Zahlungsstatus) — ersetzt, siehe ADR-0006.
- **Ticket 06** (Suche und Filter) — Zahlungsfilter wird Rückstandsfilter, Klassenfilter wird Frequenzfilter.
- **Ticket 08** (Beitragsklassen-Ansicht) — entfällt ersatzlos, siehe ADR-0005.

Das ist der Preis dafür, dass die Bestandsaufnahme der echten Tabelle erst nach dem achten Ticket kam. Er ist heute bezahlbar, weil es noch keinen Release und keinen Migrationsmechanismus gibt.

### Verweise

- Domain-Vokabular: [CONTEXT.md](../../CONTEXT.md)
- Bestandsaufnahme der Excel-Tabelle: [excel-vorlage.md](excel-vorlage.md)
- Beitrag statt Beitragsklasse: [ADR-0005](../../docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md)
- Rückstand statt `bezahlt_bis`: [ADR-0006](../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md)
- Issue-Tracker-Konvention: [docs/agents/issue-tracker.md](../../docs/agents/issue-tracker.md)
