package app

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ToniBlankenburg/boxclub/importer"
	"github.com/ToniBlankenburg/boxclub/service"
)

// Der Excel-Import hat zwei Hälften, die absichtlich nichts voneinander wissen:
// der ExcelImporter liest die Datei und weiß nichts von der Datenbank, der
// MemberService schreibt Sätze und weiß nichts von Excel. Was hier steht, ist
// nur die Naht dazwischen — Datei entgegennehmen, Sätze durchreichen, mitzählen
// (siehe .scratch/boxclub-v1/spec.md → Seams).

// maxUpload begrenzt, was der Datei-Dialog annimmt. Die Tabelle des Vereins hat
// 50 bis 200 Zeilen und wiegt ein paar hundert Kilobyte; alles jenseits davon
// ist keine Mitgliederliste mehr.
//
// Durchgesetzt wird die Grenze mit http.MaxBytesReader und nicht mit dem
// Argument von ParseMultipartForm: das ist nur der Speicherpuffer, und was
// darüber hinausgeht, schreibt Go klaglos in eine Datei im Temp-Verzeichnis.
// Eine Grenze, die nichts begrenzt, ist schlechter als gar keine.
const maxUpload = 16 << 20

// dateiendung ist die einzige Endung, die der Import annimmt. Das alte
// .xls-Format liest excelize nicht, und eine Fehlermeldung aus dem Innern der
// Bibliothek wäre für den Verein keine Auskunft.
const dateiendung = ".xlsx"

// importbericht ist, was der Verein nach einem Lauf zu sehen bekommt: vier
// Zahlen und darunter jede gescheiterte Zeile einzeln mit Grund. Die Zahlen
// stehen nebeneinander, weil sie zusammen die Frage beantworten, ob der Lauf
// der erste war oder eine Korrekturrunde.
type importbericht struct {
	Dateiname    string
	Uebernommen  int
	Neu          int
	Aktualisiert int
	Gescheitert  int

	// Fehler sind die Gründe, zeilenweise. Sie sind das eigentliche Ergebnis:
	// der Verein korrigiert damit seine Tabelle und importiert erneut.
	Fehler []importer.Zeilenfehler

	// Meldung sagt, warum die Datei als Ganzes nicht zu gebrauchen war — kein
	// .xlsx, kein Blatt „Verwaltung", eine fehlende Spalte. Sie steht neben den
	// Zeilengründen und nicht bei ihnen, weil dann gar nichts gelesen wurde und
	// die vier Zahlen nichts aussagen.
	Meldung string

	Navigation []navigationseintrag
}

// Erfolgreich sagt, ob nichts zu korrigieren blieb.
func (b importbericht) Erfolgreich() bool {
	return b.Gescheitert == 0
}

// uebernehmen setzt die gelesenen Zeilen ein und zählt mit.
//
// Eine Zeile, die der Service abweist, landet mit ihrer Excel-Zeilennummer im
// Bericht statt den Lauf abzubrechen: die Datei ist ein Altbestand, und eine
// einzelne kaputte Zeile darf die übrigen nicht aufhalten. Ein echter
// Datenbankfehler ist etwas anderes — der beendet den Lauf.
func uebernehmen(svc *service.MemberService, ergebnis importer.Ergebnis) (importbericht, error) {
	bericht := importbericht{
		Fehler:      ergebnis.Fehler,
		Gescheitert: len(ergebnis.Fehler),
	}

	for _, zeilensatz := range ergebnis.Saetze {
		wirkung, err := svc.Uebernehmen(zeilensatz.Satz)

		var vf *service.ValidierungsFehler
		if errors.As(err, &vf) {
			bericht.Fehler = append(bericht.Fehler,
				importer.Zeilenfehler{Zeile: zeilensatz.Zeile, Gruende: vf.Meldungen})
			bericht.Gescheitert++

			continue
		}
		if err != nil {
			return importbericht{}, err
		}

		bericht.Uebernommen++
		if wirkung == service.ImportNeu {
			bericht.Neu++
		} else {
			bericht.Aktualisiert++
		}
	}

	// Sortiert nach Zeilennummer, damit der Bericht dieselbe Reihenfolge hat wie
	// die Tabelle: die Fehler aus dem Einsetzen kommen sonst gesammelt hinter
	// denen aus dem Lesen.
	zeilenfehlerSortieren(bericht.Fehler)

	return bericht, nil
}

// importFormular zeigt den Datei-Dialog.
func (a *App) importFormular(w http.ResponseWriter, r *http.Request) {
	a.rendern(w, "import-formular", importbericht{Navigation: navigation(bereichImport)})
}

// importAusfuehren nimmt die hochgeladene Mappe entgegen, liest sie und setzt
// sie ein.
func (a *App) importAusfuehren(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	if err := r.ParseMultipartForm(maxUpload); err != nil {
		a.importFehler(w, fmt.Sprintf(
			"Die Datei ließ sich nicht entgegennehmen. Ist sie größer als %d MB?", maxUpload>>20))
		return
	}

	datei, kopf, err := r.FormFile("datei")
	if err != nil {
		a.importFehler(w, "Bitte eine Datei auswählen.")
		return
	}
	defer datei.Close()

	if !strings.EqualFold(filepath.Ext(kopf.Filename), dateiendung) {
		a.importFehler(w, fmt.Sprintf(
			"%q ist keine %s-Datei. Das alte .xls-Format liest der Import nicht — "+
				"in Excel einmal als .xlsx speichern.", kopf.Filename, dateiendung))
		return
	}

	ergebnis, err := importer.ExcelImporter{}.Lesen(datei)
	if err != nil {
		a.importFehler(w, err.Error())
		return
	}

	bericht, err := uebernehmen(a.svc, ergebnis)
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	bericht.Dateiname = kopf.Filename
	bericht.Navigation = navigation(bereichImport)

	a.rendern(w, "import-bericht", bericht)
}

// importFehler zeigt den Datei-Dialog erneut und sagt, warum die Datei als
// Ganzes nicht zu gebrauchen war. Das ist etwas anderes als der Fehlerbericht:
// hier ist keine einzelne Zeile gescheitert, sondern gar nichts gelesen worden.
func (a *App) importFehler(w http.ResponseWriter, text string) {
	a.rendern(w, "import-formular", importbericht{
		Meldung:    text,
		Navigation: navigation(bereichImport),
	})
}

// zeilenfehlerSortieren bringt die Gründe in die Reihenfolge der Tabelle.
func zeilenfehlerSortieren(fehler []importer.Zeilenfehler) {
	slices.SortStableFunc(fehler, func(a, b importer.Zeilenfehler) int {
		return a.Zeile - b.Zeile
	})
}
