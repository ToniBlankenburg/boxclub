# ADR-0012: Vereinslogo als eigene Spalte, SVG wird beim Upload rasterisiert

**Status:** Accepted
**Datum:** 2026-09-18

## Kontext

Ticket 23 hatte ein Logo bewusst ausgeklammert: "Kein Logo in v1. Die Fußzeile
ist der Ort, an dem später auch ein Bild landen kann; solange niemand danach
fragt, bleibt es bei Text." Jetzt wird danach gefragt — das Logo soll im
Verein-Bereich der App erscheinen und im Briefkopf jeder erzeugten Rechnung
([ADR-0009](0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md)).

Damit stellen sich zwei Fragen, die beide nicht offensichtlich zu beantworten
sind.

**Wo liegt das Bild?** Die App hat bereits einen Blob-Speicher für Anhänge:
die `dokument`-Tabelle ([ADR-0007](0007-dokumente-als-blob-in-sqlite.md)). Sie
naheliegend wiederzuverwenden hieße aber, ihre beiden Eigenschaften zu
verletzen: sie verlangt per `CHECK` eine PDF-Signatur, und jede Zeile hängt an
genau einem Mitglied oder einer Mitgliedschaft. Ein Logo ist kein PDF und
gehört zu keinem Mitglied — es gehört zu den Vereinsdaten selbst.

**In welchem Format?** Vereinslogos liegen in der Praxis oft als Vektorgrafik
vor (SVG, vom Grafiker geliefert), nicht als Foto. Die Rechnungs-PDF-Erzeugung
nutzt aber `gopdf` (siehe `service/rechnung_pdf.go`), das nur Rasterbilder
einbetten kann — ein SVG lässt sich dort nicht direkt platzieren.

## Entscheidung

Das Logo bekommt eine **eigene Spalte in `vereinsdaten`** (Blob plus
Format-Angabe), nicht einen Eintrag in `dokument`. `vereinsdaten` hat ohnehin
genau eine Zeile; keine Fremdschlüssel-Logik nötig.

Der Upload akzeptiert **PNG, JPEG und SVG** (bis 2 MB). Ein hochgeladenes SVG
wird **sofort beim Upload zu PNG rasterisiert** (längste Kante 1000 px,
Seitenverhältnis erhalten) und nur noch in dieser Form gespeichert. Ab diesem
Punkt existiert kein Vektor mehr — Web-Anzeige und Rechnungs-Briefkopf greifen
beide auf dasselbe Rasterbild zu, ohne Fallunterscheidung nach Ausgangsformat.
Die Rasterisierung braucht eine zusätzliche reine-Go-Bibliothek (z. B.
`oksvg`+`rasterx`), bewusst ohne CGo, damit die Cross-Compilation
(`modernc.org/sqlite`, `gopdf`) nicht durchbrochen wird.

## Konsequenzen

**Gut:**

- Web-Anzeige und PDF-Briefkopf zeigen **garantiert dasselbe Bild** — kein
  Sonderfall "Logo ist SVG, Rechnung bleibt text-only".
- Die PDF-Erzeugung bleibt einfach: sie bettet immer ein Rasterbild in eine
  feste Bounding-Box ein, nie eine Fallunterscheidung nach Format.
- `dokument` bleibt, was es ist: eine Ablage für PDFs an Mitglied oder
  Mitgliedschaft. Kein Sonderfall in ihrem `CHECK` oder ihrer
  Signaturprüfung.

**Schlecht:**

- Der Vektor-Vorteil eines SVGs (verlustfreie Skalierung auf jede Größe) geht
  mit der Rasterisierung verloren. Wer später ein schärferes Logo braucht,
  muss neu hochladen, nicht nur skalieren.
- Eine zusätzliche Abhängigkeit nur für diesen einen Umwandlungsschritt.

## Alternativen

- **SVG unverändert speichern, im Browser nativ anzeigen; die
  Rechnungserzeugung lässt das Bild bei einem SVG-Logo einfach weg.** Verworfen,
  weil der Verein dann in der App anders aussähe als auf seinen Rechnungen,
  ohne dass das irgendwo erklärt wäre.
- **Nur PNG/JPEG erlauben, SVG beim Upload ablehnen.** Verworfen, weil ein
  Vereinslogo häufig nur als Vektor vorliegt — der Nutzer müsste dieselbe
  Konvertierung von Hand vornehmen, die die App ihm sonst abnimmt.
- **Logo in der `dokument`-Tabelle ablegen**, um die bestehende
  Blob-Infrastruktur wiederzuverwenden. Verworfen, weil das ihre
  PDF-Signaturprüfung und ihre Bindung an Mitglied/Mitgliedschaft aufgeweicht
  hätte — beides Eigenschaften, auf die sich `dokument` als Begriff verlässt.
