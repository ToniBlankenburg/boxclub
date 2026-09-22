package service_test

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// pdf baut ein Minimal-PDF: die Kennung, an der die Ablage eine PDF-Datei
// erkennt, und dahinter etwas Unterscheidbares, damit sich zwei Verträge im Test
// auseinanderhalten lassen.
func pdf(inhalt string) []byte {
	return []byte("%PDF-1.7\n" + inhalt)
}

// vertrag ist ein hochgeladener Vertrag für die Fixtures.
func vertrag(name, inhalt string) service.NeuesDokument {
	return service.NeuesDokument{Name: name, Inhalt: pdf(inhalt)}
}

// laufendeMitgliedschaftID liefert den Zeitraum, an dem gerade gearbeitet wird.
// Die Vertragsablage arbeitet an einer Mitgliedschaft und nicht an einer Person,
// und deren ID hat der Test sonst nirgends her.
func laufendeMitgliedschaftID(t *testing.T, svc *service.MemberService, mitgliedID int64) int64 {
	t.Helper()

	m, err := svc.Get(mitgliedID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	laufend := m.LaufendeMitgliedschaft()
	if laufend == nil {
		t.Fatalf("Mitglied %d hat keine laufende Mitgliedschaft", mitgliedID)
	}

	return laufend.ID
}

func TestVertragAblegen_LegtIhnAmZeitraumAb(t *testing.T) {
	svc := neuerService(t)

	mitgliedID := mitgliedAnlegen(t, svc, "Anna", "Berger")
	mitgliedschaftID := laufendeMitgliedschaftID(t, svc, mitgliedID)

	if err := svc.VertragAblegen(mitgliedschaftID, vertrag("vertrag-berger.pdf", "unterschrieben")); err != nil {
		t.Fatalf("VertragAblegen: %v", err)
	}

	m, err := svc.Get(mitgliedID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	abgelegt := m.Mitgliedschaften[0].Vertrag
	if abgelegt == nil {
		t.Fatal("Vertrag = nil, erwartet den abgelegten")
	}
	if abgelegt.Name != "vertrag-berger.pdf" {
		t.Errorf("Name = %q, erwartet \"vertrag-berger.pdf\"", abgelegt.Name)
	}
	if abgelegt.Art != service.ArtVertrag {
		t.Errorf("Art = %q, erwartet %q", abgelegt.Art, service.ArtVertrag)
	}
	if abgelegt.AbgelegtAm.IsZero() {
		t.Error("AbgelegtAm ist der Nullwert, erwartet den Zeitpunkt der Ablage")
	}

	geholt, err := svc.VertragInhalt(mitgliedschaftID)
	if err != nil {
		t.Fatalf("VertragInhalt: %v", err)
	}
	if !bytes.Equal(geholt.Inhalt, pdf("unterschrieben")) {
		t.Errorf("Inhalt = %q, erwartet die hochgeladene Datei", geholt.Inhalt)
	}
	if geholt.ID != abgelegt.ID {
		t.Errorf("ID = %d, erwartet %d — dasselbe Dokument", geholt.ID, abgelegt.ID)
	}
}

// Der Vertrag hängt an der Mitgliedschaft und nicht an der Person: wer austritt
// und wiederkommt, unterschreibt einen neuen, und jeder bleibt bei seinem
// Zeitraum (CONTEXT.md → Vertrag).
func TestVertragAblegen_JederZeitraumHatSeinenEigenen(t *testing.T) {
	svc := neuerService(t)

	mitgliedID := mitgliedAnlegen(t, svc, "Jonas", "Hartmann")
	ersterZeitraum := laufendeMitgliedschaftID(t, svc, mitgliedID)

	if err := svc.VertragAblegen(ersterZeitraum, vertrag("alt.pdf", "von 2026")); err != nil {
		t.Fatalf("VertragAblegen (erster Zeitraum): %v", err)
	}

	if err := svc.SetKuendigung(mitgliedID, austrittZum(heuteVersetzt(-30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(mitgliedID, heuteVersetzt(-1)); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	zweiterZeitraum := laufendeMitgliedschaftID(t, svc, mitgliedID)

	// Ein Wiedereintritt beginnt ohne Vertrag — so wie er ohne Anmeldegebühr
	// und ohne Trainingstermine beginnt.
	m, err := svc.Get(mitgliedID)
	if err != nil {
		t.Fatalf("Get nach Wiedereintritt: %v", err)
	}
	if v := m.Mitgliedschaften[1].Vertrag; v != nil {
		t.Errorf("Vertrag des neuen Zeitraums = %+v, erwartet nil", *v)
	}

	if err := svc.VertragAblegen(zweiterZeitraum, vertrag("neu.pdf", "von 2029")); err != nil {
		t.Fatalf("VertragAblegen (zweiter Zeitraum): %v", err)
	}

	m, err = svc.Get(mitgliedID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 2 {
		t.Fatalf("%d Mitgliedschaften, erwartet 2", len(m.Mitgliedschaften))
	}

	for i, erwartet := range []string{"alt.pdf", "neu.pdf"} {
		abgelegt := m.Mitgliedschaften[i].Vertrag
		if abgelegt == nil {
			t.Errorf("Vertrag des %d. Zeitraums = nil, erwartet %q", i+1, erwartet)
			continue
		}
		if abgelegt.Name != erwartet {
			t.Errorf("Vertrag des %d. Zeitraums = %q, erwartet %q", i+1, abgelegt.Name, erwartet)
		}
	}
}

func TestVertragAblegen_ErsetztDenVorhandenen(t *testing.T) {
	svc := neuerService(t)

	mitgliedID := mitgliedAnlegen(t, svc, "Sina", "Wolf")
	mitgliedschaftID := laufendeMitgliedschaftID(t, svc, mitgliedID)

	if err := svc.VertragAblegen(mitgliedschaftID, vertrag("schief.pdf", "erste Seite fehlt")); err != nil {
		t.Fatalf("erstes VertragAblegen: %v", err)
	}
	if err := svc.VertragAblegen(mitgliedschaftID, vertrag("richtig.pdf", "vollständig")); err != nil {
		t.Fatalf("zweites VertragAblegen: %v", err)
	}

	geholt, err := svc.VertragInhalt(mitgliedschaftID)
	if err != nil {
		t.Fatalf("VertragInhalt: %v", err)
	}
	if geholt.Name != "richtig.pdf" {
		t.Errorf("Name = %q, erwartet \"richtig.pdf\" — der zweite ersetzt den ersten", geholt.Name)
	}
	if !bytes.Equal(geholt.Inhalt, pdf("vollständig")) {
		t.Errorf("Inhalt = %q, erwartet die zweite Datei", geholt.Inhalt)
	}

	// Danebengelegt wurde nichts: der Zeitraum hat einen Vertrag, nicht zwei.
	m, err := svc.Get(mitgliedID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Mitgliedschaften[0].Vertrag.ID != geholt.ID {
		t.Errorf("am Zeitraum hängt Dokument %d, geholt wurde %d",
			m.Mitgliedschaften[0].Vertrag.ID, geholt.ID)
	}
}

func TestVertragEntfernen_NimmtIhnAusDerAblage(t *testing.T) {
	svc := neuerService(t)

	mitgliedID := mitgliedAnlegen(t, svc, "Katrin", "Lorenz")
	mitgliedschaftID := laufendeMitgliedschaftID(t, svc, mitgliedID)

	if err := svc.VertragAblegen(mitgliedschaftID, vertrag("vertrag.pdf", "unterschrieben")); err != nil {
		t.Fatalf("VertragAblegen: %v", err)
	}
	if err := svc.VertragEntfernen(mitgliedschaftID); err != nil {
		t.Fatalf("VertragEntfernen: %v", err)
	}

	m, err := svc.Get(mitgliedID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v := m.Mitgliedschaften[0].Vertrag; v != nil {
		t.Errorf("Vertrag = %+v, erwartet nil nach dem Entfernen", *v)
	}

	if _, err := svc.VertragInhalt(mitgliedschaftID); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("VertragInhalt = %v, erwartet ErrNichtGefunden", err)
	}

	// Ein zweites Entfernen kommt aus einer veralteten Ansicht und soll sich als
	// solches zu erkennen geben, statt stillschweigend zu gelingen. Gemeldet wird
	// dabei, *was* fehlt: der Vertrag und nicht der Zeitraum — die Ansicht soll
	// nicht behaupten, es gäbe den Zeitraum nicht mehr.
	err = svc.VertragEntfernen(mitgliedschaftID)
	if !errors.Is(err, service.ErrKeinVertrag) {
		t.Errorf("zweites VertragEntfernen = %v, erwartet ErrKeinVertrag", err)
	}
	if !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("zweites VertragEntfernen = %v, erwartet auch ErrNichtGefunden", err)
	}
}

func TestVertragAblegen_WeistAlleAbAusserPDF(t *testing.T) {
	svc := neuerService(t)

	mitgliedID := mitgliedAnlegen(t, svc, "Timo", "Reuter")
	mitgliedschaftID := laufendeMitgliedschaftID(t, svc, mitgliedID)

	// Die Endung lügt: entschieden wird am Anfang der Datei.
	err := svc.VertragAblegen(mitgliedschaftID, service.NeuesDokument{
		Name: "vertrag.pdf", Inhalt: []byte("PK\x03\x04 das ist eine Word-Datei"),
	})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("VertragAblegen = %v, erwartet einen ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) != 1 {
		t.Fatalf("Meldungen = %+v, erwartet genau eine", validierung.Meldungen)
	}
	m := validierung.Meldungen[0]
	if m.Schluessel != "validierung.dokument.kein_pdf" || len(m.Args) == 0 || m.Args[0] != "vertrag.pdf" {
		t.Errorf("Meldung = %+v, erwartet eine, die die Datei beim Namen nennt", m)
	}

	if _, err := svc.VertragInhalt(mitgliedschaftID); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("VertragInhalt = %v, erwartet ErrNichtGefunden — abgelegt wurde nichts", err)
	}
}

func TestVertragAblegen_WeistZuGrosseDateiAb(t *testing.T) {
	svc := neuerService(t)

	mitgliedID := mitgliedAnlegen(t, svc, "Nils", "Bauer")
	mitgliedschaftID := laufendeMitgliedschaftID(t, svc, mitgliedID)

	// Ein Byte über der Grenze: gültiges PDF, nur zu groß.
	zuGross := append(pdf(""), bytes.Repeat([]byte("x"), service.MaxDokumentBytes)...)

	err := svc.VertragAblegen(mitgliedschaftID, service.NeuesDokument{Name: "scan.pdf", Inhalt: zuGross})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("VertragAblegen = %v, erwartet einen ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) != 1 || validierung.Meldungen[0].Schluessel != "validierung.dokument.zu_gross" {
		t.Errorf("Meldungen = %+v, erwartet eine eigene Meldung zur Größe", validierung.Meldungen)
	}

	// Genau an der Grenze geht es durch: die Grenze schließt ihren Wert ein.
	gerade := append(pdf(""), bytes.Repeat([]byte("x"), service.MaxDokumentBytes-len(pdf("")))...)
	if err := svc.VertragAblegen(mitgliedschaftID, service.NeuesDokument{Name: "scan.pdf", Inhalt: gerade}); err != nil {
		t.Fatalf("VertragAblegen an der Grenze: %v", err)
	}
}

func TestVertragAblegen_MeldetBeideGruendeAufEinmal(t *testing.T) {
	svc := neuerService(t)

	mitgliedID := mitgliedAnlegen(t, svc, "Lea", "Schuster")
	mitgliedschaftID := laufendeMitgliedschaftID(t, svc, mitgliedID)

	video := bytes.Repeat([]byte("x"), service.MaxDokumentBytes+1)

	err := svc.VertragAblegen(mitgliedschaftID, service.NeuesDokument{Name: "training.mov", Inhalt: video})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("VertragAblegen = %v, erwartet einen ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) != 2 {
		t.Errorf("Meldungen = %q, erwartet beide Gründe auf einmal", validierung.Meldungen)
	}
}

func TestVertragAblegen_NimmtDenDateinamenOhneVerzeichnis(t *testing.T) {
	svc := neuerService(t)

	mitgliedID := mitgliedAnlegen(t, svc, "Ruth", "Pohl")
	mitgliedschaftID := laufendeMitgliedschaftID(t, svc, mitgliedID)

	faelle := []struct {
		hochgeladen string
		erwartet    string
	}{
		{`C:\Users\verein\Scans\vertrag.pdf`, "vertrag.pdf"},
		{"/home/verein/scans/vertrag.pdf", "vertrag.pdf"},
		{"  vertrag.pdf  ", "vertrag.pdf"},
		// Ohne brauchbaren Namen springt ein Standard ein: ein gültiges PDF
		// deswegen abzuweisen wäre unverhältnismäßig.
		{"   ", "dokument.pdf"},
	}

	for _, fall := range faelle {
		if err := svc.VertragAblegen(mitgliedschaftID, vertrag(fall.hochgeladen, "egal")); err != nil {
			t.Fatalf("VertragAblegen(%q): %v", fall.hochgeladen, err)
		}

		geholt, err := svc.VertragInhalt(mitgliedschaftID)
		if err != nil {
			t.Fatalf("VertragInhalt: %v", err)
		}
		if geholt.Name != fall.erwartet {
			t.Errorf("Name nach %q = %q, erwartet %q", fall.hochgeladen, geholt.Name, fall.erwartet)
		}
	}
}

func TestVertragAblegen_MeldetUnbekanntenZeitraum(t *testing.T) {
	svc := neuerService(t)

	if err := svc.VertragAblegen(4711, vertrag("vertrag.pdf", "egal")); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("VertragAblegen = %v, erwartet ErrNichtGefunden", err)
	}
}

// Der Blob ist die einzige Spalte des Schemas, die teuer werden kann: ein
// SELECT, das ihn über mehrere Zeilen mitzieht, macht die Ansicht unbenutzbar —
// und zwar erst dann, wenn Daten drin sind, also nicht in der Entwicklung
// (ADR-0007). Das Schema schützt davor nicht; dieser Test tut es.
//
// Geprüft wird die Quelle und nicht ein Laufzeitverhalten: die Regel ist eine
// über Abfragen, und wer sie bricht, tut das beim Schreiben einer Abfrage. Die
// Ausnahmen stehen namentlich darunter — eine neue kommt nur dazu, wenn jemand
// sie hier einträgt und dabei liest, warum es sie gibt.
//
// Geprüft wird die *zusammengesetzte* Anweisung und nicht das einzelne Literal:
// eine Spaltenliste steht in diesem Package auch mal als eigene Konstante da
// (mitgliedschaftsspalten), und Literal für Literal gelesen entginge dem Test
// genau das Muster, das er fangen soll.
func TestInhaltStehtInKeinerAbfrageMitMehrerenZeilen(t *testing.T) {
	// schema legt die Spalte an und migrate führt schema aus. VertragAblegen und
	// rechnungAblegen schreiben je genau eine Zeile, VertragInhalt und
	// RechnungInhalt lesen je genau eine.
	erlaubt := []string{"schema", "migrate", "VertragAblegen", "VertragInhalt", "rechnungAblegen", "RechnungInhalt"}

	for _, datei := range quelldateien(t) {
		baum, err := parser.ParseFile(token.NewFileSet(), datei, nil, 0)
		if err != nil {
			t.Fatalf("%s parsen: %v", datei, err)
		}

		konstanten := zeichenkettenkonstanten(baum)

		for _, deklaration := range baum.Decls {
			name := deklarationsname(deklaration)
			anweisungen := sqlText(deklaration, konstanten)

			if !strings.Contains(anweisungen, "SELECT") && !strings.Contains(anweisungen, "INSERT") &&
				!strings.Contains(anweisungen, "UPDATE") && !strings.Contains(anweisungen, "CREATE") {
				continue
			}

			if strings.Contains(strings.ToLower(anweisungen), "inhalt") && !slices.Contains(erlaubt, name) {
				t.Errorf("%s: %s nennt die Spalte inhalt. Sie gehört in keine Abfrage, die "+
					"mehr als eine Zeile liefert (ADR-0007) — liefert diese wirklich nur eine, "+
					"gehört der Name in die Liste erlaubt in diesem Test", datei, name)
			}
		}
	}
}

// Ein SELECT * ist der Weg, auf dem der Blob in eine Abfrage gerät, ohne dass
// jemand ihn hinschreibt — genau der Fall, den ADR-0007 beim Namen nennt. In
// diesem Package steht keiner, und dieser Test hält das so: eine Spaltenliste
// ist etwas, das man beim Lesen der Abfrage sieht, und ein Stern ist es nicht.
func TestKeineAbfrageLiestAlleSpalten(t *testing.T) {
	for _, datei := range quelldateien(t) {
		baum, err := parser.ParseFile(token.NewFileSet(), datei, nil, 0)
		if err != nil {
			t.Fatalf("%s parsen: %v", datei, err)
		}

		konstanten := zeichenkettenkonstanten(baum)

		for _, deklaration := range baum.Decls {
			if strings.Contains(sqlText(deklaration, konstanten), "SELECT *") {
				t.Errorf("%s: %s liest alle Spalten. Welche gebraucht werden, gehört hingeschrieben — "+
					"sonst steht eines Tages der Blob aus dokument mit darin (ADR-0007)",
					datei, deklarationsname(deklaration))
			}
		}
	}
}

// quelldateien sind die Go-Dateien des Packages ohne die Tests. Geprüft wird,
// was in Betrieb geht, und nicht, was hier danebensteht.
func quelldateien(t *testing.T) []string {
	t.Helper()

	alle, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("Quelldateien suchen: %v", err)
	}

	quellen := make([]string, 0, len(alle))
	for _, datei := range alle {
		if !strings.HasSuffix(datei, "_test.go") {
			quellen = append(quellen, datei)
		}
	}
	if len(quellen) == 0 {
		t.Fatal("keine Quelldateien gefunden — der Test prüfte sonst nichts")
	}

	return quellen
}

// zeichenkettenkonstanten sammelt die Zeichenketten-Konstanten der Datei nach
// Namen. Über sie setzt sqlText eine Anweisung wieder zusammen, die im Code aus
// mehreren Stücken besteht.
func zeichenkettenkonstanten(baum *ast.File) map[string]string {
	werte := map[string]string{}

	for _, deklaration := range baum.Decls {
		allgemein, ok := deklaration.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spezifikation := range allgemein.Specs {
			wert, ok := spezifikation.(*ast.ValueSpec)
			if !ok || len(wert.Names) != 1 || len(wert.Values) != 1 {
				continue
			}
			if text, ok := zeichenkette(wert.Values[0]); ok {
				werte[wert.Names[0].Name] = text
			}
		}
	}

	return werte
}

// sqlText setzt zusammen, was eine Deklaration an SQL enthält: alle
// Zeichenkettenliterale darin und die Werte der Konstanten, auf die sie sich
// beruft. Das Ergebnis ist keine gültige Anweisung, sondern der Text, in dem
// Spaltennamen und Schlüsselwörter vorkommen — mehr braucht die Prüfung nicht.
func sqlText(deklaration ast.Decl, konstanten map[string]string) string {
	var text strings.Builder

	ast.Inspect(deklaration, func(knoten ast.Node) bool {
		switch k := knoten.(type) {
		case *ast.BasicLit:
			if wert, ok := zeichenkette(k); ok {
				text.WriteString(wert)
				text.WriteString(" ")
			}
		case *ast.Ident:
			if wert, bekannt := konstanten[k.Name]; bekannt {
				text.WriteString(wert)
				text.WriteString(" ")
			}
		}

		return true
	})

	return text.String()
}

// zeichenkette liefert den Wert eines Zeichenkettenliterals.
func zeichenkette(ausdruck ast.Expr) (string, bool) {
	literal, ok := ausdruck.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}

	text, err := strconv.Unquote(literal.Value)
	if err != nil {
		return "", false
	}

	return text, true
}

// deklarationsname benennt die Deklaration, in der ein Literal steht — eine
// Funktion bei ihrem Namen, eine Konstante bei dem der ersten Angabe.
func deklarationsname(deklaration ast.Decl) string {
	switch d := deklaration.(type) {
	case *ast.FuncDecl:
		return d.Name.Name
	case *ast.GenDecl:
		for _, spezifikation := range d.Specs {
			if wert, ok := spezifikation.(*ast.ValueSpec); ok && len(wert.Names) > 0 {
				return wert.Names[0].Name
			}
		}
	}

	return fmt.Sprintf("Deklaration in Zeile %d", deklaration.Pos())
}
