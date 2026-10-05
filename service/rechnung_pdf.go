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

// RechnungBeschriftungen sind alle Textbausteine, die rechnungPDF auf dem PDF
// selbst setzt — Präfixe wie "Rechnungsnummer: " und die Spaltenköpfe der
// Positionstabelle. Eine erzeugte Rechnung trägt die zum Erstellzeitpunkt
// aktive Anzeigesprache (Ticket 06): service/ übersetzt dafür nicht selbst
// (ADR-0002/ADR-0017 — kein i18n-Import hier), sondern bekommt den fertigen
// Text von app/rechnung.go hereingereicht.
type RechnungBeschriftungen struct {
	TelefonPraefix         string
	EMailPraefix           string
	RechnungsnummerPraefix string
	RechnungsdatumPraefix  string
	ZahlungszielPraefix    string
	LeistungsdatumPraefix  string
	// LeistungsdatumGleich steht statt eines Datums, wenn keines angegeben ist:
	// der Zeitpunkt der Leistung entspricht dem Rechnungsdatum (§ 14 Abs. 4
	// Nr. 6 UStG erlaubt diesen Hinweis).
	LeistungsdatumGleich string
	SteuernummerPraefix  string
	// RechnungAn steht als kleine Zeile über dem Empfänger.
	RechnungAn string
	// ZahlungshinweisVorlage hat einen %s für das Zahlungsziel. Leer heißt:
	// kein Hinweis unter den Summen.
	ZahlungshinweisVorlage string
	TitelPraefix           string
	SpalteBezeichnung      string
	SpalteMenge            string
	SpalteEinzelpreis      string
	SpalteSumme            string
	NettoPraefix           string
	// SteuerVorlage ist ein fmt.Sprintf-Format mit einem Platzhalter (Steuersatz
	// als Text) — der Betrag steht rechts daneben, siehe zeichner.summen.
	SteuerVorlage       string
	GesamtbetragPraefix string
	IBANPraefix         string
	BICPraefix          string
}

// RechnungBeschriftungenDeutsch ist der deutsche Vorgabetext: Standard für
// jeden Aufrufer, der keine eigene Übersetzung mitgibt — allen voran die
// Tests dieses Pakets, die service/ bewusst ohne i18n-Import halten
// (ADR-0002).
var RechnungBeschriftungenDeutsch = RechnungBeschriftungen{
	TelefonPraefix:         "Telefon: ",
	EMailPraefix:           "E-Mail: ",
	RechnungsnummerPraefix: "Rechnungsnummer: ",
	RechnungsdatumPraefix:  "Rechnungsdatum: ",
	ZahlungszielPraefix:    "Zahlungsziel: ",
	LeistungsdatumPraefix:  "Leistungsdatum: ",
	LeistungsdatumGleich:   "wie Rechnungsdatum",
	SteuernummerPraefix:    "Steuernummer: ",
	RechnungAn:             "Rechnung an",
	ZahlungshinweisVorlage: "Bitte überweisen Sie den Gesamtbetrag bis zum %s unter Angabe der Rechnungsnummer auf das unten genannte Konto.",
	TitelPraefix:           "Rechnung ",
	SpalteBezeichnung:      "Bezeichnung",
	SpalteMenge:            "Menge",
	SpalteEinzelpreis:      "Einzelpreis (netto)",
	SpalteSumme:            "Summe (netto)",
	NettoPraefix:           "Netto: ",
	SteuerVorlage:          "zzgl. %s %% USt",
	GesamtbetragPraefix:    "Gesamtbetrag: ",
	IBANPraefix:            "IBAN: ",
	BICPraefix:             "BIC: ",
}

// rechnungPDF setzt eine Rechnung aus den Vereinsdaten (Briefkopf,
// Bankverbindung, Fußzeile) und der Eingabe (Empfänger, Positionen, Beträge) in
// ein PDF um. Das Layout ist bewusst schlicht — Ticket 25 verzichtet ausdrücklich
// auf einen Unit-Test dafür und verlangt stattdessen einen Blick von Hand auf ein
// erzeugtes Exemplar.
func rechnungPDF(verein Vereinsdaten, eingabe RechnungEingabe, beschriftungen RechnungBeschriftungen) ([]byte, error) {
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

	z := &zeichner{pdf: pdf, beschriftungen: beschriftungen}

	z.briefkopf(verein)
	z.empfaengerUndKopf(verein, eingabe)
	z.titel(eingabe)
	tabellenEnde := z.positionstabelle(eingabe.Positionen)
	summenEnde := z.summen(betraege, eingabe.SteuersatzProzent, tabellenEnde)
	z.zahlungshinweis(eingabe, summenEnde)
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
	pdf            *gopdf.GoPdf
	err            error
	beschriftungen RechnungBeschriftungen
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

// Farben der Rechnung: dunkles Grau statt Schwarz für den Haupttext wirkt
// ruhiger; Nebensächliches (Kontakt, Rechnungsdaten, Fußzeile) steht gedämpft.
var (
	farbeText        = gopdf.RGBColor{R: 30, G: 30, B: 30}
	farbeGedaempft   = gopdf.RGBColor{R: 115, G: 115, B: 115}
	farbeLinie       = gopdf.RGBColor{R: 215, G: 215, B: 215}
	farbeLinieDunkel = gopdf.RGBColor{R: 60, G: 60, B: 60}
)

// textfarbe setzt die Farbe für alle folgenden Schreibaufrufe.
func (z *zeichner) textfarbe(f gopdf.RGBColor) {
	z.pdf.SetTextColor(f.R, f.G, f.B)
}

// linie zieht eine waagerechte Linie von x1 bis x2 auf Höhe y.
func (z *zeichner) linie(x1, x2, y, breite float64, f gopdf.RGBColor) {
	if z.err != nil {
		return
	}

	z.pdf.SetStrokeColor(f.R, f.G, f.B)
	z.pdf.SetLineWidth(breite)
	z.pdf.Line(x1, y, x2, y)
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

	z.textfarbe(farbeText)
	z.schreiben(schriftFett, 15, x, 18, verein.Name)

	z.textfarbe(farbeGedaempft)
	y := 25.0
	for _, zeile := range vereinKontaktzeilen(verein, z.beschriftungen) {
		z.schreiben(schriftRegulaer, 9, x, y, zeile)
		y += 4.5
	}
	z.textfarbe(farbeText)
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
func vereinKontaktzeilen(verein Vereinsdaten, b RechnungBeschriftungen) []string {
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
		zeilen = append(zeilen, b.TelefonPraefix+verein.Telefon)
	}
	if verein.Email != "" {
		zeilen = append(zeilen, b.EMailPraefix+verein.Email)
	}

	return zeilen
}

// empfaengerUndKopf setzt links den Empfänger unter der kleinen Zeile „Rechnung
// an“ und rechts die Rechnungsdaten als Block aus Bezeichnung und Wert. Beide
// Spalten beginnen auf derselben Höhe wie das Anschriftfeld nach DIN 5008.
//
// Zur Rechnung gehören nach § 14 Abs. 4 UStG außer Nummer und Datum der
// Zeitpunkt der Leistung und die Steuernummer des Vereins — beide stehen hier.
func (z *zeichner) empfaengerUndKopf(verein Vereinsdaten, eingabe RechnungEingabe) {
	const (
		startY = 58.0
		zeile  = 5.5
	)

	z.textfarbe(farbeGedaempft)
	z.schreiben(schriftRegulaer, 8, rand, startY, z.beschriftungen.RechnungAn)

	z.textfarbe(farbeText)
	y := startY + 6
	z.schreiben(schriftFett, 11, rand, y, eingabe.Empfaenger.Name)
	y += zeile
	if !eingabe.Empfaenger.Anschrift.Leer() {
		if eingabe.Empfaenger.Anschrift.Adresse != "" {
			z.schreiben(schriftRegulaer, 11, rand, y, eingabe.Empfaenger.Anschrift.Adresse)
			y += zeile
		}
		if ort := eingabe.Empfaenger.Anschrift.OrtZeile(); ort != "" {
			z.schreiben(schriftRegulaer, 11, rand, y, ort)
		}
	}

	leistung := z.beschriftungen.LeistungsdatumGleich
	if !eingabe.Leistungsdatum.IsZero() {
		leistung = eingabe.Leistungsdatum.Format(deutschesDatum)
	}

	angaben := [][2]string{
		{z.beschriftungen.RechnungsnummerPraefix, eingabe.Nummer},
		{z.beschriftungen.RechnungsdatumPraefix, eingabe.Rechnungsdatum.Format(deutschesDatum)},
		{z.beschriftungen.LeistungsdatumPraefix, leistung},
		{z.beschriftungen.ZahlungszielPraefix, eingabe.Zahlungsziel.Format(deutschesDatum)},
	}
	if verein.Steuernummer != "" {
		angaben = append(angaben, [2]string{z.beschriftungen.SteuernummerPraefix, verein.Steuernummer})
	}

	const blockLinks = 120.0
	xRechts := rand + inhaltsbreite

	y = startY + 6
	for _, angabe := range angaben {
		z.textfarbe(farbeGedaempft)
		z.schreiben(schriftRegulaer, 9, blockLinks, y, etikett(angabe[0]))
		z.textfarbe(farbeText)
		z.schreibenRechts(schriftRegulaer, 9, xRechts, y, xRechts-blockLinks-30, angabe[1])
		y += zeile
	}
}

// etikett macht aus einem Präfix wie „Rechnungsnummer: “ die Bezeichnung ohne
// Doppelpunkt — im Block steht der Wert in einer eigenen Spalte daneben.
func etikett(praefix string) string {
	return strings.TrimSuffix(strings.TrimSpace(praefix), ":")
}

// titel setzt die Überschrift über der Positionstabelle.
func (z *zeichner) titel(eingabe RechnungEingabe) {
	z.schreiben(schriftFett, 18, rand, 100, strings.TrimSpace(z.beschriftungen.TitelPraefix)+" "+eingabe.Nummer)
}

// positionsTabelleStartY und -zeilenhoehe legen fest, wo die Tabelle beginnt
// und wie hoch jede ihrer Zeilen ist — beide werden auch für die Berechnung des
// Endpunkts gebraucht, an dem die Summenzeilen weitergehen.
const (
	positionsTabelleStartY = 116.0
	positionsZeilenhoehe   = 10.0
)

// positionstabelle zeichnet die Positionen als Tabelle mit Kopfzeile und
// liefert die y-Koordinate, an der sie endet.
func (z *zeichner) positionstabelle(positionen []Rechnungsposition) float64 {
	ende := positionsTabelleStartY + float64(len(positionen)+1)*positionsZeilenhoehe
	if z.err != nil {
		return ende
	}

	// Kein Gitter: unter jeder Zeile eine feine Linie, unter der Kopfzeile eine
	// kräftigere. Das hält die Tabelle ruhig und lässt die Zahlen die Arbeit tun.
	trennlinie := gopdf.BorderStyle{Bottom: true, Width: 0.2, RGBColor: farbeLinie}
	kopflinie := gopdf.BorderStyle{Bottom: true, Width: 0.4, RGBColor: farbeLinieDunkel}

	tabelle := z.pdf.NewTableLayout(rand, positionsTabelleStartY, positionsZeilenhoehe, len(positionen))
	// Die Köpfe setzt gopdf immer zentriert; deshalb bleiben sie in der Tabelle
	// leer (nur die Linie darunter kommt von dort) und stehen unten ausgerichtet
	// wie ihre Spalten.
	breiten := [4]float64{inhaltsbreite * 0.40, inhaltsbreite * 0.12, inhaltsbreite * 0.26, inhaltsbreite * 0.22}
	ausrichtung := [4]string{"left", "right", "right", "right"}
	for i := range breiten {
		tabelle.AddColumn("", breiten[i], ausrichtung[i])
	}
	tabelle.SetTableStyle(gopdf.CellStyle{})

	tabelle.SetHeaderStyle(gopdf.CellStyle{
		BorderStyle: kopflinie, TextColor: farbeText,
		Font: schriftFett, FontSize: 9,
	})
	tabelle.SetCellStyle(gopdf.CellStyle{
		BorderStyle: trennlinie, TextColor: farbeText,
		Font: schriftRegulaer, FontSize: 10,
	})

	z.tabellenkopf([4]string{
		z.beschriftungen.SpalteBezeichnung, z.beschriftungen.SpalteMenge,
		z.beschriftungen.SpalteEinzelpreis, z.beschriftungen.SpalteSumme,
	}, breiten)

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

// tabellenkopf setzt die Spaltenköpfe in die Kopfzeile der Positionstabelle:
// die erste Spalte linksbündig, die Zahlenspalten rechtsbündig. 2 mm Abstand
// zum Zellrand entsprechen dem Innenabstand von gopdf.
func (z *zeichner) tabellenkopf(koepfe [4]string, breiten [4]float64) {
	const innen = 2.0
	y := positionsTabelleStartY + (positionsZeilenhoehe-3.2)/2 - 0.3

	z.textfarbe(farbeText)
	x := rand
	for i, kopf := range koepfe {
		if i == 0 {
			z.schreiben(schriftFett, 9, x+innen, y, kopf)
		} else {
			z.schreibenRechts(schriftFett, 9, x+breiten[i]-innen, y, breiten[i]-2*innen, kopf)
		}
		x += breiten[i]
	}
}

// summen setzt Netto, Steuer und Brutto unter die Tabelle: rechts ein Block
// aus Bezeichnung links und Betrag rechts. Bei 0 % entfällt die Steuerzeile —
// es gäbe nichts anzuzeigen, was Netto und Brutto nicht schon zeigen (siehe
// Ticket 25, AC); der Grund dafür steht dann in der Fußzeile (§ 19 UStG). Die
// Rückgabe ist die y-Koordinate unter dem Block.
func (z *zeichner) summen(betraege Rechnungsbetraege, steuersatzProzent, y float64) float64 {
	const breite = 80.0
	xRechts := rand + inhaltsbreite
	xLinks := xRechts - breite

	zeile := func(schriftart string, groesse float64, farbe gopdf.RGBColor, bezeichnung, betrag string) {
		z.textfarbe(farbe)
		z.schreiben(schriftart, groesse, xLinks, y, bezeichnung)
		z.schreibenRechts(schriftart, groesse, xRechts, y, 35, betrag)
	}

	y += 8
	zeile(schriftRegulaer, 10, farbeGedaempft, etikett(z.beschriftungen.NettoPraefix), euroAnzeige(betraege.NettoCents))
	y += 6

	if steuersatzProzent > 0 {
		zeile(schriftRegulaer, 10, farbeGedaempft,
			fmt.Sprintf(z.beschriftungen.SteuerVorlage, SteuersatzAlsText(steuersatzProzent)),
			euroAnzeige(betraege.SteuerCents))
		y += 6
	}

	// Die Linie trennt die Summe von den Zwischenbeträgen; der Gesamtbetrag ist
	// das Einzige, was fett und groß steht.
	y += 1
	z.linie(xLinks, xRechts, y, 0.4, farbeLinieDunkel)
	y += 4
	zeile(schriftFett, 12, farbeText, etikett(z.beschriftungen.GesamtbetragPraefix), euroAnzeige(betraege.BruttoCents))

	return y + 8
}

// zahlungshinweis setzt den Satz mit dem Zahlungsziel unter die Summen, über die
// ganze Breite. Er wiederholt, was oben im Block steht, aber in einem Satz, den
// man beim Überweisen liest.
func (z *zeichner) zahlungshinweis(eingabe RechnungEingabe, y float64) {
	if z.err != nil || z.beschriftungen.ZahlungshinweisVorlage == "" {
		return
	}

	text := fmt.Sprintf(z.beschriftungen.ZahlungshinweisVorlage, eingabe.Zahlungsziel.Format(deutschesDatum))
	z.absatz(schriftRegulaer, 9, farbeGedaempft, rand, y+6, inhaltsbreite, text)
}

// absatz setzt Text mit Zeilenumbruch in einer Spalte der gegebenen Breite.
func (z *zeichner) absatz(schriftart string, groesse float64, farbe gopdf.RGBColor, x, y, breite float64, text string) {
	if z.err != nil || text == "" {
		return
	}

	if err := z.pdf.SetFont(schriftart, "", groesse); err != nil {
		z.err = fmt.Errorf("schriftart setzen: %w", err)
		return
	}

	z.textfarbe(farbe)
	z.pdf.SetXY(x, y)
	if err := z.pdf.MultiCellWithOption(&gopdf.Rect{W: breite, H: groesse * 0.6},
		text, gopdf.CellOption{
			Align: gopdf.Left | gopdf.Top,
			// An Leerzeichen umbrechen und nicht mitten im Wort.
			BreakOption: &gopdf.BreakOption{Mode: gopdf.BreakModeIndicatorSensitive, BreakIndicator: ' '},
		}); err != nil {
		z.err = fmt.Errorf("absatz schreiben: %w", err)
	}
}

// fusszeile setzt unter eine feine Linie links die Bankverbindung und rechts die
// frei getippte Fußzeile. Beide sind freiwillig (Ticket 23) und fehlen einfach,
// wenn sie nicht gepflegt sind. Die Fußzeile bricht in ihrer Spalte um, statt
// über den Rand zu laufen.
func (z *zeichner) fusszeile(verein Vereinsdaten) {
	const (
		linieY = 256.0
		startY = 260.0
		spalte = 105.0
	)

	z.linie(rand, rand+inhaltsbreite, linieY, 0.2, farbeLinie)
	z.textfarbe(farbeGedaempft)

	y := startY
	for _, zeile := range bankverbindungKontaktzeilen(verein, z.beschriftungen) {
		z.schreiben(schriftRegulaer, 8, rand, y, zeile)
		y += 4
	}

	if verein.Fusszeile == "" {
		return
	}

	z.absatz(schriftRegulaer, 8, farbeGedaempft, spalte, startY, rand+inhaltsbreite-spalte, verein.Fusszeile)
}

// bankverbindungKontaktzeilen sind IBAN, BIC und Kreditinstitut als einzelne
// Zeilen — dieselben drei Angaben, die Ticket 23 am Verein pflegt.
func bankverbindungKontaktzeilen(verein Vereinsdaten, b RechnungBeschriftungen) []string {
	var zeilen []string

	if verein.IBAN != "" {
		zeilen = append(zeilen, b.IBANPraefix+verein.IBAN)
	}
	if verein.BIC != "" {
		zeilen = append(zeilen, b.BICPraefix+verein.BIC)
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
