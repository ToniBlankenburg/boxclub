package service

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// Das Vereinslogo (CONTEXT.md → Vereinslogo, ADR-0012) hängt an den
// Vereinsdaten und nicht an der Dokumentablage: es ist kein PDF und gehört zu
// keinem Mitglied, beides Eigenschaften, auf die sich dokument.go verlässt.
//
// Hochgeladen werden darf PNG, JPEG oder SVG. Gespeichert wird davon nur
// Raster: ein SVG wird hier, beim Upload, in ein PNG umgewandelt. Ab da
// existiert kein Vektor mehr — Web-Anzeige und Rechnungs-Briefkopf greifen auf
// dasselbe Bild zu, ohne Fallunterscheidung nach Ausgangsformat (ADR-0012).

// MaxLogoBytes ist die Obergrenze für ein hochgeladenes Logo: 2 MB. Ein Logo
// ist ein paar hundert Kilobyte groß; die Grenze fängt einen versehentlich
// ausgewählten Foto-Upload ab, lange bevor er in der Datenbank liegt.
const MaxLogoBytes = 2 << 20

// logoZielkante ist die längste Kante, auf die ein hochgeladenes SVG
// gerastert wird (ADR-0012) — genug für die 30-mm-Bounding-Box im
// Rechnungs-Briefkopf (service/rechnung_pdf.go) auch in Druckqualität, aber
// klein genug, um zusammen mit MaxLogoBytes unproblematisch zu bleiben.
const logoZielkante = 1000

// PNGKennung und JPEGKennung sind die Anfangsbytes, an denen logoFormat die
// beiden Rasterformate erkennt — exportiert, damit die Tests dieselben Bytes
// prüfen, die auch die Implementierung verwendet, statt sie ein zweites Mal
// von Hand abzuschreiben.
var (
	PNGKennung  = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	JPEGKennung = []byte{0xFF, 0xD8, 0xFF}
)

// keinLogoformat sagt, warum eine Datei nicht als Logo angenommen wurde, und
// nennt sie beim Namen — hochgeladen wird aus einem Dateidialog heraus, und
// welche der ausgewählten Dateien gemeint ist, soll niemand raten müssen.
func keinLogoformat(name string) Meldung {
	return meldung("validierung.logo.kein_format", name)
}

// logoZuGross ist die eigene Meldung für die Größengrenze des Logos, aus
// demselben Grund wie zuGross bei Dokumenten: beide Zahlen stehen da, sonst
// bleibt unklar, ob die Datei knapp darüber liegt oder ob die falsche
// ausgewählt wurde.
func logoZuGross(name string, groesse int) Meldung {
	return meldung("validierung.logo.zu_gross", name, megabyte(groesse), Logogrenze())
}

// Logogrenze schreibt die Obergrenze so, wie sie dem Verein gezeigt wird —
// derselbe Gedanke wie Dokumentgrenze: der Hinweis unter dem Datei-Dialog und
// die Absage bei zu großem Upload sollen von derselben Zahl sprechen.
func Logogrenze() string {
	return megabyte(MaxLogoBytes)
}

// LogoAusUpload prüft eine hochgeladene Datei und liefert sie in der Form, in
// der sie hinterlegt wird: unverändert bei PNG oder JPEG, als PNG gerastert
// bei SVG (ADR-0012). Beide Prüfungen laufen unabhängig voneinander, aus
// demselben Grund wie bei Dokumenten (siehe dokument.go, pruefen): wer eine
// zu große Datei im falschen Format auswählt, soll beide Gründe auf einmal
// erfahren.
func LogoAusUpload(dateiname string, inhalt []byte) ([]byte, string, error) {
	var meldungen []Meldung

	zuGross := len(inhalt) > MaxLogoBytes
	if zuGross {
		meldungen = append(meldungen, logoZuGross(dateiname, len(inhalt)))
	}

	bild, mime, formatErkannt := logoFormat(inhalt)
	if !formatErkannt {
		meldungen = append(meldungen, keinLogoformat(dateiname))
	}

	if len(meldungen) > 0 {
		return nil, "", &ValidierungsFehler{Meldungen: meldungen}
	}

	return bild, mime, nil
}

// LogoAenderung ist, was das Verein-Formular zum Logo mitbringt: nichts (das
// gespeicherte bleibt stehen), ein angekreuztes Entfernen oder eine neu
// ausgewählte Datei.
type LogoAenderung struct {
	Entfernen bool

	// Hochgeladen unterscheidet "keine Datei ausgewählt" von "eine leere
	// Datei ausgewählt" — Inhalt allein könnte beides sein, ein nil-Slice und
	// ein leerer, aber nicht-nil-Slice sehen für den Aufrufer gleich aus.
	Hochgeladen bool
	Dateiname   string
	Inhalt      []byte
}

// LogoAktualisieren entscheidet, welches Logo eine Speicherung übernehmen
// soll — unverändert, entfernt oder ersetzt, je nachdem, was das Formular
// mitbringt (ADR-0012). Eine neu ausgewählte Datei geht vor einem
// angekreuzten Entfernen: wer beides zugleich tut, wollte ersetzen und hat
// die Checkbox nur nicht wieder ausgeknipst.
//
// aktuellesLogo/aktuellesMime sind das gespeicherte Logo, das unverändert
// bleibt, wenn weder hochgeladen noch entfernt wird — SetVereinsdaten selbst
// kennt kein "unverändert lassen" (siehe dessen Dokumentation), das muss also
// vorher entschieden sein.
func LogoAktualisieren(aktuellesLogo []byte, aktuellesMime string, aenderung LogoAenderung) ([]byte, string, error) {
	if aenderung.Hochgeladen {
		return LogoAusUpload(aenderung.Dateiname, aenderung.Inhalt)
	}
	if aenderung.Entfernen {
		return nil, "", nil
	}

	return aktuellesLogo, aktuellesMime, nil
}

// svgWurzelelement steht in jedem SVG, gleich welche Deklaration oder
// Kommentare davor stehen — anders als bei PNG und JPEG gibt es für ein SVG
// keine feste Anfangssignatur. Ohne diese Vorprüfung würde svgRastern
// akzeptieren: sein XML-Parser bricht bei beliebigem Text, der kein
// SVG-Wurzelelement enthält, nicht mit einem Fehler ab, sondern liefert
// stillschweigend ein leeres Icon — jede Datei wäre dann ein "gültiges", aber
// blankes SVG.
var svgWurzelelement = []byte("<svg")

// logoFormat erkennt PNG und JPEG an ihrer Signatur, denselben Anfangsbytes,
// nach denen dokument.go ein PDF erkennt — geprüft wird der Inhalt und nicht
// die Endung des Namens. Alles mit einem SVG-Wurzelelement wird als SVG
// gerastert; alles andere ist keines der drei Formate.
func logoFormat(inhalt []byte) (bild []byte, mime string, ok bool) {
	switch {
	case bytes.HasPrefix(inhalt, PNGKennung):
		return inhalt, "image/png", true
	case bytes.HasPrefix(inhalt, JPEGKennung):
		return inhalt, "image/jpeg", true
	}

	if !bytes.Contains(inhalt, svgWurzelelement) {
		return nil, "", false
	}

	bild, err := svgRastern(inhalt)
	if err != nil {
		return nil, "", false
	}

	return bild, "image/png", true
}

// svgRastern zeichnet ein SVG in ein PNG mit logoZielkante als längster Kante,
// Seitenverhältnis erhalten (ADR-0012). oksvg und rasterx sind reine
// Go-Bibliotheken ohne CGo, wie auch gopdf und modernc.org/sqlite — die
// Cross-Compilation der App bleibt so unangetastet.
func svgRastern(inhalt []byte) ([]byte, error) {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(inhalt))
	if err != nil {
		return nil, fmt.Errorf("svg lesen: %w", err)
	}

	breite, hoehe := logoMasse(icon.ViewBox.W, icon.ViewBox.H)
	icon.SetTarget(0, 0, float64(breite), float64(hoehe))

	bild := image.NewRGBA(image.Rect(0, 0, breite, hoehe))
	scanner := rasterx.NewScannerGV(breite, hoehe, bild, bild.Bounds())
	raster := rasterx.NewDasher(breite, hoehe, scanner)
	icon.Draw(raster, 1.0)

	var puffer bytes.Buffer
	if err := png.Encode(&puffer, bild); err != nil {
		return nil, fmt.Errorf("svg als png kodieren: %w", err)
	}

	return puffer.Bytes(), nil
}

// logoMasse berechnet die Pixel-Zielgröße für svgRastern: die längere Seite
// der SVG-ViewBox wird logoZielkante, die kürzere im selben Verhältnis
// kleiner. Fehlt die ViewBox (Breite oder Höhe 0 oder negativ), wird
// quadratisch gerastert — ein SVG ohne ViewBox ist selten genug, dass eine
// Annahme hier keine Überraschung ist.
func logoMasse(breite, hoehe float64) (int, int) {
	if breite <= 0 || hoehe <= 0 {
		return logoZielkante, logoZielkante
	}

	b, h := skaliertAufKante(breite, hoehe, logoZielkante)
	return int(b), int(h)
}

// skaliertAufKante ist die Rechnung hinter "längste Kante wird ziel,
// Seitenverhältnis erhalten" — für logoMasse (Pixel, SVG-Rasterung) genauso
// wie für logoMasseMM (Millimeter, PDF-Briefkopf, service/rechnung_pdf.go).
// Voraussetzung ist, dass breite und hoehe schon als positiv geprüft sind;
// was mit einer ungültigen Größe geschehen soll, entscheidet jeder Aufrufer
// für sich — der eine rastert trotzdem quadratisch, der andere bricht ab.
func skaliertAufKante(breite, hoehe, ziel float64) (float64, float64) {
	if breite >= hoehe {
		return ziel, hoehe / breite * ziel
	}
	return breite / hoehe * ziel, ziel
}
