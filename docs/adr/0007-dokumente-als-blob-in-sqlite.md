# ADR-0007: Dokumente liegen als Blob in der Datenbank

**Status:** Accepted
**Datum:** 2026-09-12

## Kontext

Der Verein will zwei Papiere in der App haben: den **eingescannten Vertrag** eines
Mitglieds und die **selbst geschriebene Rechnung** über ein Einzeltraining. Beides
sind PDFs, beide gehören zu genau einem Mitglied.

Damit stellt sich zum ersten Mal die Frage, wo die App etwas ablegt, das keine
Tabellenzeile ist. [ADR-0003](0003-datenbank-im-benutzer-konfigurationsordner.md)
hat den Ort der Datenbank geklärt — `~/Library/Application Support/Boxclub/boxclub.db`
—, sagt aber nichts über Dateien daneben.

Die Größenordnung ist klein und bekannt: 50–200 Mitglieder, je ein Vertrag, dazu
einzelne Rechnungen. Ein gescannter Vertrag liegt bei ein paar hundert Kilobyte,
eine erzeugte Rechnung deutlich darunter. Der gesamte Bestand bleibt im
zweistelligen Megabyte-Bereich.

Maßgeblich ist außerdem, **wer die App bedient**: ein Vereinsadmin auf einem
einzelnen Mac, ohne Server, ohne Betreuung. Seine Sicherung ist das Kopieren einer
Datei.

## Entscheidung

PDFs werden als **Blob in der SQLite-Datenbank** gespeichert, zusammen mit
Dateiname, Art und Zeitpunkt. Es gibt kein Dokumentenverzeichnis im Dateisystem.

Zum Ansehen und Weitergeben bietet die App einen **Export**: das PDF wird auf
Wunsch irgendwohin geschrieben, wo der Benutzer es haben will. Die App selbst
bleibt die einzige Stelle, die den Bestand verwaltet.

## Konsequenzen

**Gut:**

- **Eine Datei ist der ganze Bestand.** Sichern heißt `boxclub.db` kopieren. Wer
  einen Ordner daneben vergisst, verliert sonst genau die Unterlagen, die er
  aufgehoben hat, weil sie wichtig waren.
- **Keine toten Verweise.** Ein Datenbankeintrag, der auf eine verschobene,
  umbenannte oder gelöschte Datei zeigt, kann nicht entstehen. Damit entfällt die
  gesamte Fehlerbehandlung dafür.
- Dokument und Mitglied ändern sich **in derselben Transaktion**. Es gibt keinen
  Zustand, in dem das eine geschrieben ist und das andere nicht.
- „Meine Mitgliederdaten verlassen die Festplatte nicht" bleibt trivial wahr —
  es gibt nur einen Ort, an dem etwas liegt.

**Schlecht:**

- **Ohne die App kommt niemand an ein PDF.** Fällt sie aus, braucht es ein
  SQLite-Werkzeug, um an den Vertrag zu kommen. Der Export mildert das nur,
  solange die App läuft.
- Die Datenbank **wächst um Größenordnungen**: aus einer Datei von wenigen hundert
  Kilobyte wird eine von einigen zehn Megabyte. Für SQLite ist das folgenlos, für
  das Gefühl beim Sichern nicht.
- Ein ersetztes oder entferntes Dokument gibt seinen Platz erst nach einem
  `VACUUM` zurück. Bei dieser Menge ist das eine Fußnote, kein Problem.
- Große Blobs beim Lesen der Liste mitzuschleppen wäre teuer. Die Dokumentenspalte
  darf deshalb **nie** in einer Abfrage stehen, die mehrere Mitglieder liefert —
  das ist eine Regel, an die man sich halten muss, und kein Schutz, den das Schema
  gibt.

## Alternativen

- **Dateien in einem Verzeichnis neben der Datenbank**, benannt nach Mitglied und
  Art (`0042-mustermann-vertrag.pdf`). Der Admin fände seine Verträge im Finder,
  auch wenn die App kaputt ist — das ist das ernsthafteste Argument dagegen.
  Dagegen steht, dass zwei Dinge zueinander passen müssen: Verzeichnis verschoben,
  Datei umbenannt, Datenbank aus dem Backup zurückgespielt und Ordner nicht — jede
  dieser Alltäglichkeiten erzeugt einen Verweis ins Leere, den die App behandeln
  muss und den der Benutzer nicht versteht.
- **Nur Pfade speichern, Dateien beim Benutzer lassen.** Billigste Variante, aber
  die App verwaltet dann nichts, sondern merkt sich Fremdes. Beim ersten Aufräumen
  des Download-Ordners ist die Hälfte weg.
