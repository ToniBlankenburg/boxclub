package service

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strconv"
	"strings"

	"github.com/signintech/gopdf"
)

// Das PDF wird mit gopdf gesetzt (github.com/signintech/gopdf) — einer reinen
// Go-Bibliothek ohne CGo, wie Ticket 25 verlangt. Die Wahl fiel nicht auf
// go-pdf/fpdf: das Repository ist als "Archived" markiert, während gopdf zum
// Zeitpunkt der Umsetzung (September 2026) noch Releases bekam.
//
// Für Umlaute braucht es eine eingebettete Schrift — auf die in jedem PDF-Reader
// eingebauten Basisschriften ist bei "ü", "ö", "ä" und "ß" kein Verlass.
// Eingebettet ist DejaVu Sans (Regular und Bold, service/fonts): sie deckt
// Umlaute und das Eurozeichen ab und steht unter einer Lizenz, die das
// Einbetten ausdrücklich erlaubt (service/fonts/LICENSE.txt).

//go:embed fonts/DejaVuSans.ttf
var schriftDatenRegulaer []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var schriftDatenFett []byte

// Namen, unter denen die beiden Schriftschnitte bei gopdf registriert werden.
// Es sind zwei eigenständige Schriftfamilien und keine Stile derselben: gopdf
// verlangt für "B" eine mitgeladene Fettschrift unter demselben Familiennamen,
// und zwei getrennte Namen sind hier der einfachere Weg dahin.
const (
	schriftRegulaer = "rechnung-regulaer"
	schriftFett     = "rechnung-fett"
)

// Layout in Millimetern — die Einheit, die Config.Unit unten für alle Koordinaten
// festlegt. Die Seite selbst bleibt A4, unabhängig von dieser Einheit: PageSizeA4
// trägt seine eigene, in Punkt (siehe gopdf/page_sizes.go).
const (
	seitenbreite  = 210.0
	rand          = 20.0
	inhaltsbreite = seitenbreite - 2*rand
)

// rechnungPDF setzt eine Rechnung aus den Vereinsdaten (Briefkopf,
// Bankverbindung, Fußzeile) und der Eingabe (Empfänger, Positionen, Beträge) in
// ein PDF um. Das Layout ist bewusst schlicht — Ticket 25 verzichtet ausdrücklich
// auf einen Unit-Test dafür und verlangt stattdessen einen Blick von Hand auf ein
// erzeugtes Exemplar.
func rechnungPDF(verein Vereinsdaten, eingabe RechnungEingabe) ([]byte, error) {
	betraege := eingabe.Betraege()

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{Unit: gopdf.UnitMM, PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	if err := pdf.AddTTFFontData(schriftRegulaer, schriftDatenRegulaer); err != nil {
		return nil, fmt.Errorf("schrift laden: %w", err)
	}
	if err := pdf.AddTTFFontData(schriftFett, schriftDatenFett); err != nil {
		return nil, fmt.Errorf("schrift laden: %w", err)
	}

	z := &zeichner{pdf: pdf}

	z.briefkopf(verein)
	z.empfaengerUndKopf(eingabe)
	z.titel(eingabe)
	tabellenEnde := z.positionstabelle(eingabe.Positionen)
	z.summen(betraege, eingabe.SteuersatzProzent, tabellenEnde)
	z.fusszeile(verein)

	if z.err != nil {
		return nil, z.err
	}

	return pdf.GetBytesPdfReturnErr()
}

// zeichner hält den laufenden Fehler einer PDF-Seite fest, statt ihn nach jedem
// einzelnen Aufruf durchzureichen. Sobald einer auftritt, werden alle weiteren
// Aufrufe zu No-Ops — dasselbe Muster wie bei einem bufio.Writer, das sich hier
// anbietet, weil eine Rechnung aus einer langen Kette kleiner Schreibaufrufe
// besteht und jeder einzelne dieselbe, uninteressante Fehlerbehandlung bräuchte.
type zeichner struct {
	pdf *gopdf.GoPdf
	err error
}

// schreiben schreibt eine einzeilige Angabe linksbündig ab (x, y).
func (z *zeichner) schreiben(schriftart string, groesse, x, y float64, text string) {
	if z.err != nil || text == "" {
		return
	}

	if err := z.pdf.SetFont(schriftart, "", groesse); err != nil {
		z.err = fmt.Errorf("schriftart setzen: %w", err)
		return
	}

	z.pdf.SetXY(x, y)
	if err := z.pdf.Cell(nil, text); err != nil {
		z.err = fmt.Errorf("text schreiben: %w", err)
	}
}

// schreibenRechts schreibt eine einzeilige Angabe rechtsbündig, endend an der
// x-Koordinate xRechts.
func (z *zeichner) schreibenRechts(schriftart string, groesse, xRechts, y, breite float64, text string) {
	if z.err != nil || text == "" {
		return
	}

	if err := z.pdf.SetFont(schriftart, "", groesse); err != nil {
		z.err = fmt.Errorf("schriftart setzen: %w", err)
		return
	}

	z.pdf.SetXY(xRechts-breite, y)
	if err := z.pdf.CellWithOption(&gopdf.Rect{W: breite, H: groesse},
		text, gopdf.CellOption{Align: gopdf.Right | gopdf.Top}); err != nil {
		z.err = fmt.Errorf("text schreiben: %w", err)
	}
}

// logoMaxMM ist die längste Kante der Bounding-Box, in die das Vereinslogo im
// Briefkopf gesetzt wird — Seitenverhältnis erhalten (ADR-0012).
const logoMaxMM = 30.0

// logoAbstand ist der Weißraum zwischen einem gezeichneten Logo und dem Text
// daneben.
const logoAbstand = 6.0

// briefkopf setzt das Vereinslogo, sofern eines hinterlegt ist, sowie Name,
// Anschrift und Kontakt des Vereins oben links — die Angaben aus Ticket 23,
// die genau dafür gepflegt werden, plus das Logo aus ADR-0012. Steht ein Logo
// da, rückt der Text um seine Breite nach rechts; ohne Logo steht er wie
// bisher am Rand.
func (z *zeichner) briefkopf(verein Vereinsdaten) {
	x := rand
	if breite := z.logo(verein.Logo); breite > 0 {
		x += breite + logoAbstand
	}

	z.schreiben(schriftFett, 14, x, 18, verein.Name)

	y := 25.0
	for _, zeile := range vereinKontaktzeilen(verein) {
		z.schreiben(schriftRegulaer, 9, x, y, zeile)
		y += 4.5
	}
}

// logo zeichnet das Vereinslogo oben links in eine Bounding-Box von
// logoMaxMM und liefert seine tatsächliche Breite, an der sich briefkopf für
// den Text daneben richtet. Ohne Logo (inhalt leer) liefert es 0 und tut
// sonst nichts.
func (z *zeichner) logo(inhalt []byte) float64 {
	if z.err != nil || len(inhalt) == 0 {
		return 0
	}

	breite, hoehe, err := logoMasseMM(inhalt)
	if err != nil {
		z.err = fmt.Errorf("logo vermessen: %w", err)
		return 0
	}

	holder, err := gopdf.ImageHolderByBytes(inhalt)
	if err != nil {
		z.err = fmt.Errorf("logo laden: %w", err)
		return 0
	}

	if err := z.pdf.ImageByHolder(holder, rand, 14, &gopdf.Rect{W: breite, H: hoehe}); err != nil {
		z.err = fmt.Errorf("logo zeichnen: %w", err)
		return 0
	}

	return breite
}

// logoMasseMM liest die Pixelmaße des gespeicherten Logos (PNG oder JPEG,
// nie SVG — das wurde beim Upload schon gerastert, ADR-0012) und skaliert sie
// auf logoMaxMM als längste Kante, Seitenverhältnis erhalten.
func logoMasseMM(inhalt []byte) (breite, hoehe float64, err error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(inhalt))
	if err != nil {
		return 0, 0, fmt.Errorf("bildgröße lesen: %w", err)
	}

	b, h := float64(cfg.Width), float64(cfg.Height)
	if b <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("ungültige bildgröße %dx%d", cfg.Width, cfg.Height)
	}

	breite, hoehe = skaliertAufKante(b, h, logoMaxMM)
	return breite, hoehe, nil
}

// vereinKontaktzeilen sind Anschrift, Telefon und E-Mail des Vereins als
// einzelne Zeilen unter dem Vereinsnamen — jede Angabe ist freiwillig (Ticket
// 23) und fehlt einfach, wenn sie nicht gepflegt ist.
func vereinKontaktzeilen(verein Vereinsdaten) []string {
	var zeilen []string

	if !verein.Anschrift.Leer() {
		if verein.Anschrift.Adresse != "" {
			zeilen = append(zeilen, verein.Anschrift.Adresse)
		}
		if ort := verein.Anschrift.OrtZeile(); ort != "" {
			zeilen = append(zeilen, ort)
		}
	}
	if verein.Telefon != "" {
		zeilen = append(zeilen, "Telefon: "+verein.Telefon)
	}
	if verein.Email != "" {
		zeilen = append(zeilen, "E-Mail: "+verein.Email)
	}

	return zeilen
}

// empfaengerUndKopf setzt den Empfänger links und Rechnungsnummer, -datum und
// Zahlungsziel rechtsbündig daneben.
func (z *zeichner) empfaengerUndKopf(eingabe RechnungEingabe) {
	const startY = 50.0

	y := startY
	z.schreiben(schriftRegulaer, 11, rand, y, eingabe.Empfaenger.Name)
	y += 5
	if !eingabe.Empfaenger.Anschrift.Leer() {
		if eingabe.Empfaenger.Anschrift.Adresse != "" {
			z.schreiben(schriftRegulaer, 11, rand, y, eingabe.Empfaenger.Anschrift.Adresse)
			y += 5
		}
		if ort := eingabe.Empfaenger.Anschrift.OrtZeile(); ort != "" {
			z.schreiben(schriftRegulaer, 11, rand, y, ort)
		}
	}

	const breiteKopf = 80.0
	xRechts := rand + inhaltsbreite

	kopfzeilen := []string{
		"Rechnungsnummer: " + eingabe.Nummer,
		"Rechnungsdatum: " + eingabe.Rechnungsdatum.Format(deutschesDatum),
		"Zahlungsziel: " + eingabe.Zahlungsziel.Format(deutschesDatum),
	}
	y = startY
	for _, zeile := range kopfzeilen {
		z.schreibenRechts(schriftRegulaer, 10, xRechts, y, breiteKopf, zeile)
		y += 5
	}
}

// titel setzt die Überschrift über der Positionstabelle.
func (z *zeichner) titel(eingabe RechnungEingabe) {
	z.schreiben(schriftFett, 13, rand, 78, "Rechnung "+eingabe.Nummer)
}

// positionsTabelleStartY und -zeilenhoehe legen fest, wo die Tabelle beginnt
// und wie hoch jede ihrer Zeilen ist — beide werden auch für die Berechnung des
// Endpunkts gebraucht, an dem die Summenzeilen weitergehen.
const (
	positionsTabelleStartY = 85.0
	positionsZeilenhoehe   = 8.0
)

// positionstabelle zeichnet die Positionen als Tabelle mit Kopfzeile und
// liefert die y-Koordinate, an der sie endet.
func (z *zeichner) positionstabelle(positionen []Rechnungsposition) float64 {
	ende := positionsTabelleStartY + float64(len(positionen)+1)*positionsZeilenhoehe
	if z.err != nil {
		return ende
	}

	rahmen := gopdf.BorderStyle{
		Top: true, Left: true, Right: true, Bottom: true,
		Width: 0.2, RGBColor: gopdf.RGBColor{R: 120, G: 120, B: 120},
	}

	tabelle := z.pdf.NewTableLayout(rand, positionsTabelleStartY, positionsZeilenhoehe, len(positionen))
	tabelle.AddColumn("Bezeichnung", inhaltsbreite*0.46, "left")
	tabelle.AddColumn("Menge", inhaltsbreite*0.14, "right")
	tabelle.AddColumn("Einzelpreis (netto)", inhaltsbreite*0.2, "right")
	tabelle.AddColumn("Summe (netto)", inhaltsbreite*0.2, "right")

	tabelle.SetHeaderStyle(gopdf.CellStyle{
		BorderStyle: rahmen, FillColor: gopdf.RGBColor{R: 235, G: 235, B: 235},
		Font: schriftFett, FontSize: 10,
	})
	tabelle.SetCellStyle(gopdf.CellStyle{BorderStyle: rahmen, Font: schriftRegulaer, FontSize: 10})

	for _, p := range positionen {
		tabelle.AddRow([]string{
			p.Bezeichnung,
			strconv.FormatInt(p.Menge, 10),
			euroAnzeige(p.EinzelpreisCents),
			euroAnzeige(p.SummeCents()),
		})
	}

	if err := tabelle.DrawTable(); err != nil {
		z.err = fmt.Errorf("positionstabelle zeichnen: %w", err)
	}

	return ende
}

// summen setzt Netto, Steuer und Brutto unter die Tabelle, rechtsbündig. Bei
// 0 % entfällt die Steuerzeile — es gäbe nichts anzuzeigen, was Netto und Brutto
// nicht schon zeigen (siehe Ticket 25, AC).
func (z *zeichner) summen(betraege Rechnungsbetraege, steuersatzProzent, y float64) {
	const breite = 80.0
	xRechts := rand + inhaltsbreite

	y += 4
	z.schreibenRechts(schriftRegulaer, 10, xRechts, y, breite, "Netto: "+euroAnzeige(betraege.NettoCents))
	y += 5

	if steuersatzProzent > 0 {
		beschriftung := fmt.Sprintf("zzgl. %s %% USt: %s", SteuersatzAlsText(steuersatzProzent), euroAnzeige(betraege.SteuerCents))
		z.schreibenRechts(schriftRegulaer, 10, xRechts, y, breite, beschriftung)
		y += 5
	}

	z.schreibenRechts(schriftFett, 11, xRechts, y, breite, "Gesamtbetrag: "+euroAnzeige(betraege.BruttoCents))
}

// fusszeile setzt Bankverbindung und die frei getippte Fußzeile unten auf die
// Seite. Beide sind freiwillig (Ticket 23) und fehlen einfach, wenn sie nicht
// gepflegt sind.
func (z *zeichner) fusszeile(verein Vereinsdaten) {
	const startY = 255.0

	y := startY
	for _, zeile := range bankverbindungKontaktzeilen(verein) {
		z.schreiben(schriftRegulaer, 8, rand, y, zeile)
		y += 4
	}

	if verein.Fusszeile == "" {
		return
	}

	y += 2
	for _, zeile := range strings.Split(verein.Fusszeile, "\n") {
		z.schreiben(schriftRegulaer, 8, rand, y, zeile)
		y += 4
	}
}

// bankverbindungKontaktzeilen sind IBAN, BIC und Kreditinstitut als einzelne
// Zeilen — dieselben drei Angaben, die Ticket 23 am Verein pflegt.
func bankverbindungKontaktzeilen(verein Vereinsdaten) []string {
	var zeilen []string

	if verein.IBAN != "" {
		zeilen = append(zeilen, "IBAN: "+verein.IBAN)
	}
	if verein.BIC != "" {
		zeilen = append(zeilen, "BIC: "+verein.BIC)
	}
	if verein.Kreditinstitut != "" {
		zeilen = append(zeilen, verein.Kreditinstitut)
	}

	return zeilen
}

// euroAnzeige schreibt einen Cent-Betrag, wie er auf dem PDF steht — mit
// Währungszeichen, anders als BeitragAlsEuro, das für ein Eingabefeld ohne
// Zeichen schreibt.
func euroAnzeige(cents int64) string {
	return BeitragAlsEuro(cents) + " €"
}
