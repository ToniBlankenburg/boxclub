package app

import (
	"errors"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ToniBlankenburg/boxclub/i18n"
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
	Fehler []importer.Zeilenmeldung

	// Hinweise sind dasselbe für Zeilen, die trotzdem hereingekommen sind —
	// heute die Trainingstermine, die im Stundenplan nicht zu finden waren. Sie
	// stehen im selben Bericht und in einem eigenen Abschnitt: eine Zeile mit
	// Nacharbeit ist keine gescheiterte Zeile, und beides in einer Liste ließe
	// die Zahl darüber nicht mehr zusammenpassen.
	Hinweise []importer.Zeilenmeldung

	// StundenplanLeer sagt, dass kein Trainingstermin zu vergeben ist. Ohne
	// diese Auskunft importiert jemand 200 Mitglieder und bekommt 400 Hinweise,
	// ohne zu verstehen, warum (ADR-0008).
	StundenplanLeer bool

	// Meldung sagt, warum die Datei als Ganzes nicht zu gebrauchen war — kein
	// .xlsx, kein Blatt „Verwaltung", eine fehlende Spalte. Sie steht neben den
	// Zeilengründen und nicht bei ihnen, weil dann gar nichts gelesen wurde und
	// die vier Zahlen nichts aussagen.
	Meldung string

	Navigation []navigationseintrag
}

// Erfolgreich sagt, ob nichts zu korrigieren blieb. Ein Hinweis zählt dazu: die
// Zeile ist zwar da, aber ihre Trainingszeiten sind es nicht.
func (b importbericht) Erfolgreich() bool {
	return b.Gescheitert == 0 && len(b.Hinweise) == 0
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

	// Die Hinweise zur Hand, nach Zeile: weist der Service eine Zeile doch noch
	// ab, gehen ihre Hinweise zu den Gründen und aus dieser Sammlung heraus.
	// Sonst stünde dieselbe Zeile in beiden Abschnitten des Berichts — einmal
	// als gescheitert und einmal als übernommen.
	offeneHinweise := make(map[int][]string, len(ergebnis.Hinweise))
	for _, meldung := range ergebnis.Hinweise {
		offeneHinweise[meldung.Zeile] = meldung.Meldungen
	}

	for _, zeilensatz := range ergebnis.Saetze {
		wirkung, err := svc.Uebernehmen(zeilensatz.Satz)

		var vf *service.ValidierungsFehler
		if errors.As(err, &vf) {
			bericht.Fehler = append(bericht.Fehler, importer.Zeilenmeldung{
				Zeile:     zeilensatz.Zeile,
				Meldungen: append(slices.Clone(vf.Meldungen), offeneHinweise[zeilensatz.Zeile]...),
			})
			delete(offeneHinweise, zeilensatz.Zeile)
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

	// Die übrig gebliebenen Hinweise in der Reihenfolge, in der der Importer sie
	// gefunden hat — das ist schon die Reihenfolge der Tabelle.
	for _, meldung := range ergebnis.Hinweise {
		if _, offen := offeneHinweise[meldung.Zeile]; offen {
			bericht.Hinweise = append(bericht.Hinweise, meldung)
		}
	}

	// Sortiert nach Zeilennummer, damit der Bericht dieselbe Reihenfolge hat wie
	// die Tabelle: die Fehler aus dem Einsetzen kommen sonst gesammelt hinter
	// denen aus dem Lesen.
	nachZeileSortieren(bericht.Fehler)

	return bericht, nil
}

// stundenplan holt den Terminkatalog, gegen den der Importer die Freitexte der
// Trainingsspalten hält — einschließlich der archivierten Termine, damit der
// Bericht „den gibt es nicht mehr" sagen kann und nicht „den gibt es nicht".
func (a *App) stundenplan() (importer.Stundenplan, error) {
	termine, err := a.svc.ListTrainingstermine(true)
	if err != nil {
		return importer.Stundenplan{}, err
	}

	return importer.StundenplanAus(termine), nil
}

// importFormular zeigt den Datei-Dialog.
func (a *App) importFormular(w http.ResponseWriter, r *http.Request) {
	plan, err := a.stundenplan()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.importFormularZeigen(w, plan, "")
}

// importAusfuehren nimmt die hochgeladene Mappe entgegen, liest sie und setzt
// sie ein.
func (a *App) importAusfuehren(w http.ResponseWriter, r *http.Request) {
	// Der Stundenplan zuerst: er entscheidet, was aus den Trainingsspalten wird,
	// und steht auch über jedem Dialog, der nach einer abgewiesenen Datei
	// wiederkommt.
	plan, err := a.stundenplan()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	if err := r.ParseMultipartForm(maxUpload); err != nil {
		a.importFormularZeigen(w, plan,
			i18n.Text(a.Sprache(), "import.fehler.upload_zu_gross", maxUpload>>20))
		return
	}

	datei, kopf, err := r.FormFile("datei")
	if err != nil {
		a.importFormularZeigen(w, plan, i18n.Text(a.Sprache(), "import.fehler.keine_datei"))
		return
	}
	defer datei.Close()

	if !strings.EqualFold(filepath.Ext(kopf.Filename), dateiendung) {
		a.importFormularZeigen(w, plan,
			i18n.Text(a.Sprache(), "import.fehler.falsche_endung", kopf.Filename, dateiendung))
		return
	}

	ergebnis, err := importer.ExcelImporter{}.Lesen(datei, plan, a.Sprache())
	if err != nil {
		a.importFormularZeigen(w, plan, err.Error())
		return
	}

	bericht, err := uebernehmen(a.svc, ergebnis)
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	bericht.Dateiname = kopf.Filename
	bericht.StundenplanLeer = plan.Leer()
	bericht.Navigation = a.navigation(bereichImport)

	a.rendern(w, "import-bericht", bericht)
}

// importFormularZeigen zeigt den Datei-Dialog — mit einer Meldung, wenn eine
// Datei als Ganzes nicht zu gebrauchen war. Das ist etwas anderes als der
// Fehlerbericht: dort ist eine einzelne Zeile gescheitert, hier ist gar nichts
// gelesen worden.
func (a *App) importFormularZeigen(w http.ResponseWriter, plan importer.Stundenplan, meldung string) {
	a.rendern(w, "import-formular", importbericht{
		Meldung:         meldung,
		StundenplanLeer: plan.Leer(),
		Navigation:      a.navigation(bereichImport),
	})
}

// nachZeileSortieren bringt die gescheiterten Zeilen in die Reihenfolge der
// Tabelle. Die Hinweise brauchen das nicht: sie kommen allesamt aus dem Lesen
// und stehen schon in dieser Reihenfolge.
func nachZeileSortieren(meldungen []importer.Zeilenmeldung) {
	slices.SortStableFunc(meldungen, func(a, b importer.Zeilenmeldung) int {
		return a.Zeile - b.Zeile
	})
}
