# ADR-0017: Mehrsprachige Oberfläche über einen Go-Katalog, Sprache in eigener Einstellungsdatei

**Status:** Accepted
**Datum:** 2026-09-22

## Kontext

Die App war bislang durchgängig deutsch, mit deutschem Text als Literal in
`templates/*.html` und in `app/*.go`. Das Domänenvokabular aus `CONTEXT.md`
(Mitglied, Mitgliedschaft, Rückstand, Anmeldegebühr, …) ist dort **verbindlich
deutsch** — als Bezeichner im Code, in der Datenbank und in der Spec. Jetzt
soll die Oberfläche selbst zwischen Deutsch und Englisch umschaltbar sein,
und zwar vollständig, einschließlich der Übersetzung dieser Fachbegriffe für
die Anzeige.

Das wirft zwei Fragen auf, die sich beide nicht von selbst beantworten.

**Ändert sich das Domänenmodell?** Nein. `CONTEXT.md`, die Go-Bezeichner und
die DB-Spalten bleiben deutsch — sie beschreiben, *was* die App verwaltet,
nicht, *wie* es angezeigt wird. Übersetzt wird ausschließlich das, was der
Nutzer als Text sieht: eine `Beschriftung`, ein `<label>`, eine Meldung. Ein
`Mitglied` bleibt im Code ein `Mitglied`; nur der Text "Mitglied" in der
Oberfläche wird bei englischer Sprache zu "Member".

**Wo lebt die Spracheinstellung?** Die App hat kein Login und keinen
Mehrbenutzerbetrieb (CLAUDE.md → Out of scope) — die Sprache gilt fürs ganze
Programm, nicht pro Nutzer. Naheliegend wäre die `vereinsdaten`-Zeile
(ADR-0003 kennt bereits eine Einstellung fürs ganze Programm dort:
`BOXCLUB_DB`). Dagegen spricht: `vereinsdaten` beschreibt den Verein
(CONTEXT.md → Vereinsdaten), nicht die Bedienung der App, und die
Schema-Politik dieser App kennt keine Migration — jede Schema-Änderung
bedeutet, die Dev-Datenbank zu löschen (CLAUDE.md). Eine Spracheinstellung in
derselben Zeile hieße, sie bei jedem Löschen der Dev-Datenbank mit zu
verlieren, obwohl sie mit den Mitgliederdaten nichts zu tun hat.

## Entscheidung

**Übersetzung über einen Go-Katalog, keine Bibliothek.** Zwei Maps
(`i18n/de.go`, `i18n/en.go`) von Schlüssel auf Text, nachgeschlagen über
`i18n.Text(sprache, schluessel, args...)`. Kein `go-i18n`, kein
gettext/PO-Workflow: Die App hat kein Übersetzungsteam und keinen externen
Werkzeugkette-Bedarf, und eine Go-Map ist für ~200–400 Schlüssel (Schätzung
über alle Tickets) genau die richtige Größe — durchsuchbar, typo-sicher
gegen den Compiler, ohne Laufzeit-Parsing.

**Rückfallkette Sprache → Deutsch → Schlüssel.** Fehlt ein Schlüssel in
einer Sprache, liefert `Text` den deutschen Eintrag; fehlt er auch dort, den
Schlüssel selbst. Eine unvollständige Übersetzung zeigt damit nie eine leere
Stelle, sondern höchstens deutschen Text in einer sonst englischen Ansicht —
wichtig, weil die Übersetzung über mehrere Tickets/Sessions hinweg entsteht
(siehe `.scratch/mehrsprachigkeit/spec.md`) und der Zwischenstand jederzeit
lauffähig bleiben muss.

**Spracheinstellung in einer eigenen JSON-Datei**, nicht in `vereinsdaten`:
`~/Library/Application Support/Boxclub/einstellungen.json` auf macOS,
überschreibbar über `BOXCLUB_EINSTELLUNGEN` — dasselbe Muster wie
`BOXCLUB_DB` (ADR-0003), nur als eigene Datei. Fehlt sie, gilt Deutsch als
Standard: eine bestehende Installation, die noch nie umgeschaltet hat, soll
nicht plötzlich englisch starten. Ein Lesefehler (z. B. eine von Hand
verstümmelte Datei) ist beim Start **nicht fatal** — anders als ein
Datenbankfehler, der die App zu Recht abbricht, ist eine unlesbare
Spracheinstellung kein Grund, die App gar nicht erst zu starten; sie fällt
dann auf Deutsch zurück.

**Übersetzung geschieht an der Stelle, an der der Text feststeht, nicht
doppelt.** Text, den `app/` bereits als Go-Wert zusammenbaut (z. B.
`navigationseintrag.Beschriftung`), wird dort übersetzt, bevor er ins
Template geht. Text, der ausschließlich im Template als Literal steht, wird
dort über die Template-Funktion `t` übersetzt. Beides gleichzeitig für
denselben Text wäre zwei Quellen der Wahrheit.

**`service/` bekennt keine Sprache** (Ticket 07 der Feature-Spec hält das
offen): die Fachlogik bleibt darstellungsunabhängig, wie ADR-0002 es für die
Trennung `service/`↔`app/` ohnehin verlangt. Wie Fehlermeldungen aus dem
Service übersetzt werden, ohne dass `service/` einen `i18n`-Import bekommt,
ist eine eigene Entscheidung für Ticket 07.

## Konsequenzen

- Ein neuer Schlüssel im Katalog ohne Eintrag in einer Sprache bricht nichts,
  fällt aber lautlos auf Deutsch zurück — es gibt keinen automatisierten
  Check auf Vollständigkeit der Kataloge. Das ist für die Größe der App und
  der zwei Sprachen bewusst in Kauf genommen; ein Test, der beide
  Katalog-Schlüsselmengen vergleicht, ist eine mögliche spätere Ergänzung,
  keine Voraussetzung.
- Die Rechnungs-PDF (ADR-0009) wird von dieser Entscheidung zunächst nicht
  erfasst — sie bleibt deutsch, bis Ticket 06 klärt, ob eine erzeugte
  Rechnung die UI-Sprache übernehmen soll oder unabhängig davon ist.
- `frontend/index.html` ist eine statische, von Vite gebaute Datei außerhalb
  des Go-Template-Systems (ADR-0002) und bleibt vorerst unübersetzt (siehe
  Ticket 08) — der sichtbare Bruch ist ein Lade-Blitz von unter einer
  Sekunde, kein dauerhafter Zustand.
