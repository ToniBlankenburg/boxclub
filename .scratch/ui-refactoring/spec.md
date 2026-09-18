Status: ready-for-agent

# Spec: UI-Refactoring

## Problem Statement

Die Boxclub-Mitgliederverwaltung ist fachlich fertig (alle Tickets in
`.scratch/boxclub-v1/issues/` stehen auf `ready-for-human`, bis auf Ticket 11,
den Mac-Release-Build), aber sie sieht noch nicht danach aus: eine flache
Reihe gleich aussehender Buttons als Navigation, ein zentrierter, schmaler
Hauptbereich, der bei der spaltenreichen Mitgliederliste unnötig eng wird,
unstrukturierte Formulare mit vielen Feldern hintereinander, ein Dashboard,
dessen Kacheln nicht auf einen Blick überfliegbar sind. Der Vereinsadmin
möchte, dass die App vor dem ersten Mac-Release so aufgeräumt, übersichtlich
und modern wirkt wie kommerzielle Vereinssoftware (wiso mein Verein,
farblich/stilistisch angelehnt an WISO Steuer) — damit die tägliche Nutzung
sich nach einem fertigen Werkzeug anfühlt statt nach einem Prototyp.

## Solution

Eine strukturelle (nicht nur kosmetische) Überarbeitung der gesamten
Präsentationsschicht: eine einzige neue Akzentfarbe (gedämpftes Blaugrau)
ersetzt Rot überall außer an den zwei bewusst beibehaltenen Warnstellen
(Rückstand-Kennzeichen, Formular-Fehlerlisten), die Navigation bleibt
horizontal oben — bewusst keine Sidebar, siehe
[ADR-0010](../../docs/adr/0010-navigation-bleibt-horizontale-top-leiste.md) —
wird aber zu kompakten Pillen mit Icon, der Hauptbereich nutzt die volle
Breite statt einer zentrierten Spalte, jedes Eingabeformular (Mitglied,
Trainingstermin, Rechnung, Verein) folgt demselben Muster aus getönten
Gruppen-Boxen, und jede Bericht-/Übersichtsansicht (Dashboard,
Excel-Import-Bericht) folgt demselben Muster einer durchgehenden
Kennzahlenleiste. Die Mitgliederliste bleibt eine Tabelle, die Dokumentablage
bleibt bei ihrer bestehenden Statuskarte — beide wurden geprüft und passend
befunden, nicht übersehen. Die Mitglied-Bearbeiten-Seite (Stammdaten /
Verträge / Rechnung) und das Verein-Formular (seine drei Gruppen-Boxen)
zeigen ihre Bereiche zusätzlich als Registerkarten statt gestapelt — die
eigenständige Rechnung und das Trainingstermin-Formular bewusst nicht
(Begründung in [ADR-0011](../../docs/adr/0011-registerkarten-fuer-mitglied-und-verein.md)).

Jede Entscheidung wurde vorher an einem Wegwerf-Prototyp auf einem eigenen
`prototype/*`-Branch visuell geprüft, bevor sie nach `main` gefoltet wurde.
Diese Branches bleiben als Primärquelle stehen:

- `prototype/mitgliederliste-layout` — Tabelle vs. Kartenraster vs. kompakte
  Liste
- `prototype/stilrichtungen` — Farbe, Navigation, Hauptbereich-Breite,
  Mitglied-Formular, Dashboard (drei Gesamtvarianten A/B/C)
- `prototype/vier-restbereiche` — Trainingstermine, Excel-Import, Rechnung,
  Verein
- `prototype/dokument-ablage` — Vertragsablage (drei Varianten, bestehende
  Statuskarte bestätigt)
- `prototype/registerkarten` — Mitglied- und Verein-Seite (drei Varianten:
  Reiter oben/Unterstrich, Pillen, seitliche Reiter — siehe ADR-0011)

Ein Teil der Umsetzung liegt durch die Prototyp-Folds bereits in `main`
(Commits `8fa5778`, `f3a19fa`, `<Fold-Commit Ticket 08>`); die aus dieser
Spec abgeleiteten Tickets verifizieren, härten und schließen diese Arbeit ab,
statt sie neu zu entwerfen.

## User Stories

1. Als Vereinsadmin möchte ich eine aufgeräumte, kompakte Bereichsnavigation
   mit einem kleinen Icon je Bereich, damit ich auf einen Blick erkenne, in
   welchem Bereich ich bin und wohin ich wechseln kann.
2. Als Vereinsadmin möchte ich, dass die Navigation weiterhin horizontal oben
   steht statt in einer Sidebar, damit der Hauptbereich die volle Breite für
   die spaltenreiche Mitgliederliste behält (ADR-0010).
3. Als Vereinsadmin möchte ich, dass der Hauptbereich die volle
   Fensterbreite nutzt statt in einer schmalen, zentrierten Spalte zu
   stehen, damit ich bei vielen eingeblendeten Spalten nicht horizontal
   scrollen muss.
4. Als Vereinsadmin möchte ich eine einheitliche, gedämpfte Akzentfarbe statt
   des bisherigen Rot in der ganzen App sehen, damit die Anwendung ruhiger
   wirkt und nicht wie eine durchgehende Warnmeldung aussieht.
5. Als Vereinsadmin möchte ich, dass das Rückstand-Kennzeichen weiterhin
   rot/grün bleibt, damit ich Zahlungsrückstände auch nach dem Redesign
   sofort erkenne.
6. Als Vereinsadmin möchte ich, dass Formular-Fehlerlisten weiterhin rot
   bleiben, damit Fehleingaben auch nach dem Redesign auffallen.
7. Als Vereinsadmin möchte ich, dass die Akzentfarbe an einer einzigen
   Stelle (CSS-Variablen) definiert ist, damit ein späterer Farbwechsel eine
   Änderung ist, keine Suche durch alle Templates.
8. Als Vereinsadmin möchte ich das Mitglied-Formular in klar getrennten,
   beschrifteten Gruppen (Person / Kontakt & Anschrift / Mitgliedschaft /
   Sonstiges) sehen, damit ich beim Anlegen oder Bearbeiten nicht durch eine
   unstrukturierte Feldwand scrolle.
9. Als Vereinsadmin möchte ich, dass das Trainingstermin-Formular demselben
   Gruppen-Prinzip folgt wie das Mitglied-Formular, damit sich die App
   überall gleich bedienen lässt.
10. Als Vereinsadmin möchte ich, dass das Rechnung-Formular demselben
    Gruppen-Prinzip folgt, damit ich es ohne Umgewöhnung nutze.
11. Als Vereinsadmin möchte ich, dass das Vereinsdaten-Formular demselben
    Gruppen-Prinzip folgt, damit die Stammdaten-Pflege konsistent aussieht.
12. Als Vereinsadmin möchte ich, dass das Dashboard seine Kennzahlen
    (Monatssoll, Mitgliederzahlen) als durchgehende Leiste statt als
    einzelne Karten zeigt, damit ich sie auf einen Blick überfliege, statt
    den Blick über ein Raster wandern zu lassen.
13. Als Vereinsadmin möchte ich, dass der Excel-Import-Bericht seine
    Kennzahlen (übernommene Sätze, Fehler, Hinweise) ebenfalls als
    Kennzahlenleiste zeigt statt als Fließtext, damit ich das Ergebnis eines
    Imports sofort einschätze.
14. Als Vereinsadmin möchte ich, dass die Mitgliederliste weiterhin eine
    Tabelle mit ein-/ausblendbaren Spalten ist, nicht Kacheln, damit ich bei
    50–200 Mitgliedern weiterhin schnell scanne und vergleiche (bestätigt in
    Prototyp-Runde 1, `prototype/mitgliederliste-layout`).
15. Als Vereinsadmin möchte ich, dass die Dokumentablage weiterhin ihre
    bestehende Statuskarte mit Datei-Aktionen zeigt, damit ich dort nichts
    neu lerne, wo die bisherige Lösung bereits passt (bestätigt in
    Prototyp-Runde 4, `prototype/dokument-ablage`).
16. Als Vereinsadmin möchte ich, dass Formulare weiterhin einzelne,
    aufgeräumte Formulare bleiben statt mehrstufiger Assistenten, damit das
    Anlegen/Bearbeiten bei der überschaubaren Feldzahl nicht unnötig in
    Schritte zerlegt wird.
17. Als Vereinsadmin möchte ich, dass Dark Mode nicht Teil dieses
    Refactorings ist, damit der Umfang überschaubar bleibt.
18. Als Vereinsadmin möchte ich, dass Icons in der Navigation als reine
    Verzierung markiert sind (aria-hidden), damit ein Screenreader keine
    bedeutungslosen Icon-Beschreibungen vorliest.
19. Als Vereinsadmin möchte ich, dass der aktive Navigationseintrag sichtbar
    hervorgehoben ist (Farbe plus `aria-current`), damit ich immer weiß, in
    welchem Bereich ich gerade bin.
20. Als Vereinsadmin möchte ich, dass kleinere UI-Bausteine — die
    Inline-Zeilen-Formulare in der Mitgliederliste (Rückstand/Kündigung/
    Ruhend/Wiedereintritt), die Meldung/Toast-Komponente, ein generelles
    Button-/Badge-System, Lade- und Leerzustände — hier nicht vorentschieden
    werden, damit der jeweilige Ticket-Implementierer sie beim Bauen
    pragmatisch per Augenschein entscheidet, ohne auf eine globale Vorgabe
    zu warten.
21. Als Vereinsadmin möchte ich, dass jede Design-Entscheidung an einem
    einsehbaren Prototyp-Branch mit Verdict-Datei hängt, damit ich später
    nachvollziehen kann, warum eine Variante gewählt und welche verworfen
    wurde.
22. Als Vereinsadmin möchte ich, dass der erste Mac-Release-Build (Ticket 11
    in `boxclub-v1`) erst nach Abschluss dieses Refactorings läuft, damit das
    ausgelieferte v1 das neue statt das alte UI zeigt.
23. Als Vereinsadmin möchte ich, dass die Mitglied-Bearbeiten-Seite und das
    Verein-Formular ihre Bereiche als Registerkarten statt gestapelt zeigen,
    damit ich nicht durch drei aneinandergereihte Formulare bzw. Gruppen-
    Boxen scrolle, um den gesuchten Abschnitt zu finden. Das sind
    Registerkarten und kein mehrstufiger Assistent (Story 16 bleibt in
    Kraft): jeder Reiter ist jederzeit anwählbar, kein Speichern-Schritt
    hängt von einer Reihenfolge ab.

## Implementation Decisions

- **Betroffene Module:** ausschließlich Präsentationsschicht —
  `frontend/index.html`, `frontend/src/style.css`, alle Templates in
  `templates/` außer keine strukturelle Änderung an `mitglieder_liste.html`
  und `dokument.html` (dort nur Farbtoken-Ersetzung, plus mit Ticket 08 der
  Wegfall eines `mt-6`, das aus der alten Stapel-Anordnung stammte — die
  Statuskarte selbst aus Ticket 06 bleibt unangetastet, nur ihr Platz auf der
  Seite ändert sich). Kein Fachcode in `service/` oder `importer/` ist
  betroffen oder darf es werden — das ist eine harte Grenze, keine
  Beobachtung.
- **Akzentfarbe als CSS-Variablen** in `frontend/src/style.css`:
  `--akzent: #2f5f78`, `--akzent-dunkel: #24485d`, `--akzent-hell: #eef4f7`,
  plus zwei Ring-Opazitäten. Templates referenzieren sie per
  Tailwind-Arbitrary-Value (`bg-[var(--akzent)]` usw.) statt Rot-Utility-
  Klassen fest zu kodieren. Ausgenommen von der Umfärbung: Formular-
  Fehlerlisten (`border-red-700`/`bg-red-50`/`text-red-800`) und das
  Rückstand-Kennzeichen samt seiner Inline-Fehlermeldungen — beide bilden
  ein rot/grün-Warnpaar und bleiben rot.
- **Navigation** (`templates/navigation.html`): horizontale Pillen-Leiste,
  ein kleines `aria-hidden`-Icon-SVG je Bereich (Schlüssel aus
  `navigationseintrag.Schluessel` in `app/app.go`), aktiver Eintrag über
  `aria-current="page"` plus Akzentfarbe hervorgehoben. Bleibt eine
  `hx-swap-oob`-Komponente wie bisher (ADR-0002-Kontrakt unverändert).
- **Hauptbereich** (`frontend/index.html`): `<main id="inhalt">` ohne
  `max-w`-Begrenzung, nur noch horizontales Padding.
- **Formular-Gruppen-Boxen-Muster:** ein wiederkehrendes visuelles Muster
  (getönte `rounded-md border bg-neutral-50 p-4`-Boxen mit Überschrift pro
  Themengruppe), angewendet auf `mitglied_formular.html`,
  `trainingstermine.html` (Formularteil), `rechnung.html`, `verein.html`.
  Kein gemeinsames Go- oder Template-Include für dieses Muster — jede
  Formular-Datei setzt es lokal um, weil `html/template` kein generisches
  "Gruppen-Box"-Partial mit variabler Feldliste sinnvoll anbietet und die
  Feldmengen pro Formular zu unterschiedlich sind, um sich zu lohnen zu
  abstrahieren.
- **Kennzahlenleisten-Muster:** eine durchgehende, horizontale Leiste
  einzelner Kennzahlen statt eines Kartenrasters, angewendet auf
  `dashboard.html` und den Ergebnisbericht in `import.html`. Im Zuge dessen
  entfernt: totes Kartenraster-Teil-Template `dashboard-kachel`, Go-Typ
  `kachelDaten` (`app/dashboard.go`) und Template-Funktion `kachel`
  (`app/app.go`) — durch die Kennzahlenleiste nicht mehr gebraucht.
- **Mitgliederliste bleibt Tabelle**, ein-/ausblendbare Spalten
  (Ticket 27) unverändert; nur Farbtoken-Ersetzung.
- **Registerkarten-Muster** (Ticket 08, ADR-0011): seitliche Reiter mit
  Trennlinie und getöntem Hintergrund für die Navigationsspalte, angewendet
  auf `mitglied_formular.html` (Stammdaten/Verträge/Rechnung, nur beim
  Bearbeiten) und `verein.html` (seine drei Gruppen-Boxen). Optik
  ausschließlich über Tailwind-Klassen (`aria-selected:`-Varianten), Umschalten
  über einen einzigen delegierten Klick-Handler
  (`frontend/src/registerkarten.js`, `data-registerkarten`/`data-rk-tab`/
  `data-rk-panel`). Bewusst nicht angewendet auf die eigenständige Rechnung
  (Pflichtfelder über mehrere Boxen verteilt in einem einzigen Formular —
  native Validierung bräche bei einem versteckten Pflichtfeld still ab) und
  das Trainingstermin-Formular (nur eine Box, vier Felder).
- **Dokumentablage bleibt die bestehende Statuskarte** mit Datei-Aktionen;
  nur Farbtoken-Ersetzung.
- **ADR-0010** hält die Entscheidung "Top-Nav statt Sidebar" fest — bei
  Rückfragen zur Navigationsform ist das die Primärquelle, keine erneute
  Debatte in Tickets dieser Spec.
- **Reihenfolge/Blocker:** Ticket 11 (`boxclub-v1`, Mac-Release-Build) ist
  zusätzlich zu seinen bestehenden Blockern auf den Abschluss dieser Spec
  blockiert (bereits im Ticket vermerkt).

## Testing Decisions

- Reine Präsentationsschicht-Arbeit: kein `service/`- oder `importer/`-Code
  betroffen, daher keine neuen Unit-Tests in diesen Paketen nötig.
- Vorbild/Konvention: `CLAUDE.md` → Testing — *"`app/`-Handler und
  htmx-Templates sind nicht unit-getestet; manueller Smoke-Test genügt."*
  Dieselbe Seam gilt für die Tickets dieser Spec.
- **Harte, automatisierte Prüfung pro Ticket:** `go build ./...`,
  `go vet ./...`, `go test ./...` müssen grün bleiben (stellt sicher, dass
  Refactoring kein bestehendes Verhalten in `service`/`importer`/`app`
  bricht).
- **Manueller Smoke-Test pro Ticket:** `wails dev` starten, den betroffenen
  Bereich ansehen — passend zur bereits gelebten Praxis der vier
  Prototyp-Runden (dort wurden Ergebnisse zusätzlich per Wegwerf-`httptest`/
  `a.tpl.ExecuteTemplate`-Aufrufen gegengeprüft und danach wieder entfernt,
  nicht committet; dasselbe Vorgehen ist für die Umsetzungs-Tickets
  angemessen, wo ein Zweifel an der Handler-Verdrahtung besteht).
- Kein neues Test-Tooling (kein visueller Regressionstest, kein
  UI-Test-Framework) — passend zu "UI-Testautomatisierung" unter "Out of
  scope for v1" in `CLAUDE.md`.

## Out of Scope

- Die Kleinteile aus User Story 20 (Inline-Zeilen-Formulare, Meldung/Toast-
  Komponente, Button-/Badge-System, Lade-/Leerzustände) — bewusst nicht
  vorentschieden, siehe dort.
- Dark Mode (User Story 17).
- Alles unter "Out of scope for v1" in `CLAUDE.md` (MoneyMoney-Import,
  Multi-User/Login, Cloud/Web-Deployment, Windows/Linux-Release-Builds,
  UI-Testautomatisierung, Anwesenheitserfassung, Box-spezifische Felder).
- Jede Änderung an Fachlogik, Datenbankschema oder den Domänenbegriffen aus
  `CONTEXT.md` — dieses Refactoring ändert nur, wie Bestehendes aussieht,
  nicht was es bedeutet oder wie es funktioniert.
- Eine gemeinsame, wiederverwendbare Komponentenbibliothek für das
  Gruppen-Boxen- oder Kennzahlenleisten-Muster — beide bleiben lokale,
  pro Template wiederholte Umsetzungen (siehe Implementation Decisions).

## Further Notes

- Ein Teil der Arbeit ist bereits erledigt: die Prototyp-Folds `8fa5778`
  (Farbe/Navigation/Breite/Mitglied-Formular/Dashboard) und `f3a19fa`
  (Trainingstermine/Rechnung/Verein/Import-Bericht) sowie
  [ADR-0010](../../docs/adr/0010-navigation-bleibt-horizontale-top-leiste.md)
  liegen bereits auf `main`. Die aus dieser Spec abgeleiteten Tickets sollten
  das entsprechend berücksichtigen: eher "nachziehen, prüfen, abschließen"
  als "von null bauen". Die einzige noch nicht gefoltete Entscheidung ist
  Runde 4 (Dokumentablage) — dort gab es nichts zu folden, weil die
  bestehende Lösung gewann.
- Alle vier Prototyp-Branches (`prototype/mitgliederliste-layout`,
  `prototype/stilrichtungen`, `prototype/vier-restbereiche`,
  `prototype/dokument-ablage`) bleiben als Primärquellen bestehen und sollten
  nicht gelöscht werden, auch nachdem die Tickets abgeschlossen sind — sie
  dokumentieren die verworfenen Alternativen.
- `main` liegt aktuell 28 Commits vor `origin/main` (nicht gepusht) — für
  diese Spec ohne Bedeutung, nur zur Einordnung.
