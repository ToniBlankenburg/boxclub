package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Die Handler der Dokumentablage (ADR-0007). Hochgeladen wird beim Mitglied,
// abgelegt wird am Zeitraum: ein Vertrag gehört zu der Mitgliedschaft, über die
// er geschlossen wurde, und nicht zu der Person, die ihn unterschrieben hat
// (CONTEXT.md → Vertrag).
//
// Jede der drei Aktionen antwortet mit dem Vertragsblock genau ihres Zeitraums.
// Das ist kein Detail: der Block steht unter einem halb ausgefüllten
// Mitgliedsformular, und eine Antwort, die mehr austauschte, nähme die
// Tipparbeit mit.

// maxDokumentUpload begrenzt, was der Datei-Dialog überhaupt entgegennimmt.
//
// Etwas Luft über der Grenze des Service: der multipart-Rahmen wiegt selbst ein
// paar hundert Byte, und eine Datei knapp über MaxDokumentBytes soll die
// genaue Meldung des Service bekommen ("ist 10,4 MB groß") und nicht die grobe
// von hier. Was weit darüber liegt, bricht dagegen schon hier ab, bevor es durch
// die Leitung ist.
//
// Wie beim Excel-Import ist http.MaxBytesReader das Mittel und nicht das
// Argument von ParseMultipartForm: das ist nur der Speicherpuffer, und was
// darüber hinausgeht, schreibt Go klaglos in eine Datei im Temp-Verzeichnis.
const maxDokumentUpload = service.MaxDokumentBytes + (1 << 20)

// Speicherziel fragt den Benutzer, wohin eine Datei geschrieben werden soll, und
// liefert den gewählten Pfad; der leere Pfad heißt "abgebrochen".
//
// Es ist der eine Punkt, an dem diese App das Betriebssystem braucht und nicht
// mit HTTP auskommt: das WebView von Wails kennt keine Downloads, ein
// Content-Disposition liefe also ins Leere (v2.15 behandelt sie auf keiner
// Plattform). Als Funktionstyp und nicht als direkter Aufruf des Wails-Runtime,
// damit main.go den Dialog einsetzt und ein Test etwas anderes einsetzen kann —
// die Handler bleiben so testbar, ohne ein Fenster zu öffnen.
//
// filterBeschriftung und filterMuster gehen mit, weil der Dialog inzwischen für
// mehr als eine Dateiart steht (Vertrag und Rechnung als PDF, MoneyMoney-Export
// als CSV) — ohne sie zeigte der Dialog immer denselben, fest verdrahteten
// PDF-Filter, gleich was tatsächlich gespeichert wird.
type Speicherziel func(vorschlag, filterBeschriftung, filterMuster string) (string, error)

// pdfFilter ist der Dialog-Filter für die beiden PDF-Exporte (Vertrag,
// Rechnung) — an einer Stelle benannt, damit beide dasselbe sagen.
const pdfFilterBeschriftung, pdfFilterMuster = "PDF-Dateien (*.pdf)", "*.pdf"

// SpeicherzielSetzen hinterlegt den Datei-Dialog. Er steht erst zur Verfügung,
// wenn Wails gestartet ist (OnStartup), und kann deshalb nicht schon New
// mitgegeben werden.
//
// Aufzurufen, bevor der erste Request kommt — in OnStartup ist das der Fall, dort
// steht das Fenster noch nicht. Danach wird das Feld nur noch gelesen; eine
// Absicherung dagegen, es im Betrieb zu tauschen, gibt es nicht, weil es dafür
// keinen Anlass gibt.
func (a *App) SpeicherzielSetzen(ziel Speicherziel) {
	a.speicherziel = ziel
}

// vertragDaten speisen den Vertragsblock eines Zeitraums: der Zeitraum selbst,
// an dem der Vertrag hängt, und was zuletzt mit ihm passiert ist.
type vertragDaten struct {
	Mitgliedschaft service.Mitgliedschaft
	Meldung        meldung

	// Fehler sind die Gründe, aus denen eine Datei nicht angenommen wurde —
	// dieselbe Liste wie am Mitgliedsformular, weil es dieselbe Art Auskunft
	// ist: zu groß und kein PDF können beide zugleich zutreffen.
	Fehler []string
}

// vertragAblegen nimmt ein hochgeladenes PDF entgegen und legt es an dem
// Zeitraum ab, aus dessen Block es kam. Liegt dort schon einer, ersetzt es ihn.
func (a *App) vertragAblegen(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedschaftID(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxDokumentUpload)

	if err := r.ParseMultipartForm(maxDokumentUpload); err != nil {
		// Genannt wird die Grenze des Service und nicht die etwas höhere von
		// hier: was hier abbricht, liegt in jedem Fall darüber, und zwei
		// verschiedene Zahlen für dieselbe Grenze wären nur verwirrend.
		a.vertragRendern(w, id, meldung{}, []string{
			"Die Datei ließ sich nicht entgegennehmen — sie ist größer als " +
				service.Dokumentgrenze() + "."})

		return
	}

	datei, kopf, err := r.FormFile("datei")
	if err != nil {
		a.vertragRendern(w, id, meldung{}, []string{"Bitte eine Datei auswählen."})
		return
	}
	defer datei.Close()

	inhalt, err := io.ReadAll(datei)
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	err = a.svc.VertragAblegen(id, service.NeuesDokument{Name: kopf.Filename, Inhalt: inhalt})

	var validierung *service.ValidierungsFehler
	if errors.As(err, &validierung) {
		a.vertragRendern(w, id, meldung{}, a.uebersetzeMeldungen(validierung.Meldungen))
		return
	}
	if err != nil {
		a.vertragNichtGefundenOderFehler(w, err)
		return
	}

	a.vertragRendern(w, id, meldung{Text: "Der Vertrag wurde abgelegt."}, nil)
}

// vertragEntfernen nimmt den Vertrag eines Zeitraums aus der Ablage — ohne
// Rückfrage: das Papier liegt beim Verein, und ein zweites Einscannen ist
// weniger Aufwand als ein Papierkorb, den niemand gebaut hat.
func (a *App) vertragEntfernen(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedschaftID(w, r)
	if !ok {
		return
	}

	if err := a.svc.VertragEntfernen(id); err != nil {
		a.vertragNichtGefundenOderFehler(w, err)
		return
	}

	a.vertragRendern(w, id, meldung{Text: "Der Vertrag wurde entfernt."}, nil)
}

// vertragExportieren schreibt das abgelegte PDF dorthin, wo der Benutzer es
// haben will. Ohne diesen Weg käme niemand mehr an seinen Vertrag: die
// Datenbank ist der einzige Ort, an dem er liegt (ADR-0007).
func (a *App) vertragExportieren(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedschaftID(w, r)
	if !ok {
		return
	}

	dokument, err := a.svc.VertragInhalt(id)
	if err != nil {
		a.vertragNichtGefundenOderFehler(w, err)
		return
	}

	if a.speicherziel == nil {
		// Kein Dialog hinterlegt: die App läuft außerhalb von Wails. Das ist
		// etwas anderes als ein abgebrochener Dialog und soll auch anders
		// dastehen — hier ist niemand gefragt worden.
		a.vertragRendern(w, id, meldung{
			Text:    "Der Datei-Dialog steht nicht zur Verfügung — der Vertrag lässt sich gerade nicht exportieren.",
			Warnung: true,
		}, nil)

		return
	}

	ziel, err := a.speicherziel(dokument.Name, pdfFilterBeschriftung, pdfFilterMuster)
	if err != nil {
		fehlerAntwort(w, fmt.Errorf("speicherort erfragen: %w", err))
		return
	}
	if ziel == "" {
		// Ein abgebrochener Dialog ist kein Vorgang. Der Block kommt unverändert
		// zurück, damit die laufende htmx-Anfrage ein Ziel hat.
		a.vertragRendern(w, id, meldung{}, nil)

		return
	}

	// 0600: ein Vertrag ist eine personenbezogene Unterlage und geht niemanden
	// sonst an — dieselbe Überlegung wie beim Verzeichnis der Datenbank.
	if err := os.WriteFile(ziel, dokument.Inhalt, 0o600); err != nil {
		a.vertragRendern(w, id, meldung{
			Text:    fmt.Sprintf("Der Vertrag ließ sich nicht speichern: %v", err),
			Warnung: true,
		}, nil)

		return
	}

	a.vertragRendern(w, id, meldung{Text: "Der Vertrag wurde nach " + ziel + " gespeichert."}, nil)
}

// vertragRendern zeigt den Vertragsblock eines Zeitraums mit dem Stand aus der
// Datenbank. Gelesen wird auch nach dem Ablegen: der Service bereinigt den
// Dateinamen, und der Block soll zeigen, was abgelegt ist.
func (a *App) vertragRendern(w http.ResponseWriter, id int64, m meldung, fehler []string) {
	mitgliedschaft, err := a.svc.Mitgliedschaft(id)
	if err != nil {
		a.vertragNichtGefundenOderFehler(w, err)
		return
	}

	a.rendern(w, "vertrag", vertragDaten{
		Mitgliedschaft: mitgliedschaft,
		Meldung:        m,
		Fehler:         fehler,
	})
}

// vertragNichtGefundenOderFehler beantwortet einen Service-Fehler im Kontext des
// Vertragsblocks. Was fehlt, war die Ansicht veraltet — dann führt der Weg
// zurück in die Liste, statt an der Stelle des Blocks etwas einzuwechseln, das
// dort nicht hingehört.
//
// Unterschieden wird, *was* fehlt: ein Zeitraum ohne Vertrag ist etwas anderes
// als ein Zeitraum, den es nicht gibt, und dasselbe für beides zu melden wäre
// eine falsche Auskunft. Dieselbe Unterscheidung trifft veralteteAnsichtOderFehler
// für den Lebenszyklus.
func (a *App) vertragNichtGefundenOderFehler(w http.ResponseWriter, err error) {
	var text string

	switch {
	case errors.Is(err, service.ErrKeinVertrag):
		text = "Zu diesem Zeitraum ist kein Vertrag abgelegt."
	case errors.Is(err, service.ErrNichtGefunden):
		text = "Diesen Zeitraum gibt es nicht mehr."
	default:
		fehlerAntwort(w, err)
		return
	}

	aufListeUmleiten(w)
	a.listeRendern(w, meldung{Text: text, Warnung: true})
}

// vertragsbloecke baut die Blöcke, die unter dem Mitgliedsformular stehen: einer
// je Zeitraum, jeder mit dem Vertrag, der an ihm hängt.
//
// Je Zeitraum und nicht in einer gemeinsamen Liste: wer zweimal im Verein war,
// hat zwei Verträge, und nebeneinander sähe man nicht, welcher zu welchem
// Zeitraum gehört.
func vertragsbloecke(mitgliedschaften []service.Mitgliedschaft) []vertragDaten {
	bloecke := make([]vertragDaten, 0, len(mitgliedschaften))
	for _, mitgliedschaft := range mitgliedschaften {
		bloecke = append(bloecke, vertragDaten{Mitgliedschaft: mitgliedschaft})
	}

	return bloecke
}
