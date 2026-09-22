package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ToniBlankenburg/boxclub/i18n"
	"github.com/ToniBlankenburg/boxclub/service"
)

// Die Handler des Stundenplans (ADR-0008). Sie sind so dünn wie die übrigen:
// Formularwerte einsammeln, an den MemberService geben, das Ergebnis als
// HTML-Fragment zurückgeben. Geprüft wird nichts hier — weder die Uhrzeit noch
// das Verhältnis von Beginn und Ende; beides steht in
// service.Trainingsterminangabe.

// terminEingabe hält die Rohwerte des Terminformulars, damit eine abgelehnte
// Eingabe beim erneuten Rendern nicht verloren geht.
//
// Der Wochentag steht auch hier als Text: er kommt als Wert eines <select>
// herauf, und eine abgelehnte Eingabe soll die getroffene Auswahl behalten —
// auch die leere.
type terminEingabe struct {
	Wochentag   string
	Beginn      string
	Ende        string
	Bezeichnung string
}

// terminEingabeLesen sammelt die Rohwerte des abgeschickten Formulars ein.
func terminEingabeLesen(r *http.Request) terminEingabe {
	return terminEingabe{
		Wochentag:   r.FormValue("wochentag"),
		Beginn:      r.FormValue("beginn"),
		Ende:        r.FormValue("ende"),
		Bezeichnung: r.FormValue("bezeichnung"),
	}
}

// alsAngabe übersetzt die Rohwerte in die Service-Eingabe. Ein Wochentag, der
// keine Zahl ist, wird zum Nullwert — und der heißt im Service "nicht
// ausgewählt", also genau das, was ein leeres <select> aussagt.
func (e terminEingabe) alsAngabe() service.Trainingsterminangabe {
	tag, _ := strconv.Atoi(e.Wochentag)

	return service.Trainingsterminangabe{
		Wochentag:   service.Wochentag(tag),
		Beginn:      e.Beginn,
		Ende:        e.Ende,
		Bezeichnung: e.Bezeichnung,
	}
}

// terminEingabeAus befüllt das Formular aus einem gespeicherten Termin.
func terminEingabeAus(t service.Trainingstermin) terminEingabe {
	return terminEingabe{
		Wochentag:   strconv.Itoa(int(t.Wochentag)),
		Beginn:      string(t.Beginn),
		Ende:        string(t.Ende),
		Bezeichnung: t.Bezeichnung,
	}
}

// wochentagsoptionenBauen sind die Einträge der Wochentagsauswahl. Wie die
// übrigen Filteroptionen entstehen sie in Go aus der Aufzählung des Service,
// damit Wert und Reihenfolge aus derselben Quelle kommen — die Beschriftung
// selbst kommt aber aus dem Katalog (wochentag.*) und nicht aus
// service.Wochentag.Bezeichnung(): die dient dem Excel-Abgleich (ADR-0008)
// und der Kurzschreibweise eines Termins und muss deshalb unabhängig von der
// Anzeigesprache bleiben, während diese Auswahlliste reine
// Formularbeschriftung ist (Ticket 04).
//
// Der erste Eintrag ist leer: ein Termin ohne gewählten Tag soll als solcher
// abgeschickt und vom Service abgewiesen werden können. Stünde dort Montag
// vorbelegt, bekäme jeder Termin, bei dem der Tag vergessen wurde, still einen.
func wochentagsoptionenBauen(sprache i18n.Sprache) []filteroption {
	optionen := []filteroption{{Wert: "", Beschriftung: i18n.Text(sprache, "trainingstermine.wochentag_waehlen")}}
	for _, tag := range service.Wochentage() {
		optionen = append(optionen, filteroption{
			Wert:         strconv.Itoa(int(tag)),
			Beschriftung: i18n.Text(sprache, "wochentag."+strconv.Itoa(int(tag))),
		})
	}

	return optionen
}

// terminformularDaten speist das Terminformular. Bearbeiten unterscheidet die
// beiden Modi: Neuanlage (POST auf /api/trainingstermin) und Änderung eines
// bestehenden Termins (POST auf /api/trainingstermin/{id}).
type terminformularDaten struct {
	Bearbeiten bool
	TerminID   int64
	Eingabe    terminEingabe

	// Sprache ist die aktuelle Anzeigesprache — gebraucht, um die
	// Wochentagsauswahl erst hier aufzulösen (i18n.Text), statt sie wie vor
	// Ticket 04 als deutsches Literal mitzuführen.
	Sprache i18n.Sprache

	// Archiviert sagt, ob der bearbeitete Termin derzeit aus dem Stundenplan
	// genommen ist. Das Formular ändert daran nichts und weist nur darauf hin —
	// wer ihn zurückholen will, tut das in der Liste.
	Archiviert bool

	// AuchArchivierte ist der Stand der Einblendung, aus der das Formular
	// geöffnet wurde. Er fährt mit, damit Abbrechen und Speichern in dieselbe
	// Ansicht zurückführen und nicht in eine, in der der gerade bearbeitete
	// Termin fehlt.
	AuchArchivierte bool

	Fehler []string
}

// Wochentagsoptionen sind die sieben Tage, der gewählte darunter markiert.
func (d terminformularDaten) Wochentagsoptionen() []filteroption {
	return gewaehlteOption(wochentagsoptionenBauen(d.Sprache), d.Eingabe.Wochentag)
}

// trainingstermineDaten trägt den Stundenplan, den Stand der Einblendung und
// optional eine Rückmeldung.
type trainingstermineDaten struct {
	Termine         []terminzeile
	AuchArchivierte bool
	Meldung         meldung
	Navigation      []navigationseintrag
}

// terminzeile ist ein Termin, wie die Liste ihn zeigt: der Termin selbst, seine
// Teilnehmer und die Adresse, die ihn archiviert oder zurückholt.
//
// Die Adresse entsteht hier und nicht im Template, weil in ihr zwei Dinge
// zusammenkommen — der gewünschte neue Zustand und der Stand der Einblendung —
// und die Fallunterscheidung in einem Attribut nicht mehr zu lesen wäre.
type terminzeile struct {
	service.Trainingstermin
	Teilnehmer []teilnehmerchip
	ArchivPfad string
}

// teilnehmerchip ist ein Teilnehmer, wie die Liste ihn zeigt: der Name und die
// Initialen für das Kürzel davor. Die Initialen sind reine Darstellung und
// stehen deshalb hier und nicht im Service.
type teilnehmerchip struct {
	service.Teilnehmer
	Initialen string
}

// initialenAus bildet das Kürzel aus dem ersten Buchstaben von Vor- und
// Nachname. Fehlt einer der beiden, bleibt der andere allein stehen.
func initialenAus(vorname, nachname string) string {
	var kuerzel []rune
	for _, name := range []string{vorname, nachname} {
		if erster, _ := utf8.DecodeRuneInString(strings.TrimSpace(name)); erster != utf8.RuneError {
			kuerzel = append(kuerzel, unicode.ToUpper(erster))
		}
	}

	return string(kuerzel)
}

// Anzahl ist die Zahl der Teilnehmer — dieselbe Auskunft wie
// Teilnehmerliste.Anzahl, hier, weil die Zeile die Liste nicht mehr als Ganzes
// trägt.
func (z terminzeile) Anzahl() int {
	return len(z.Teilnehmer)
}

// terminzeilen ergänzt jeden Termin um seine Teilnehmer und seine Archiv-Adresse.
//
// Mitgeschickt wird der *gewünschte* Zustand und nicht der aktuelle: so tut ein
// Klick aus einer veralteten Ansicht nicht das Gegenteil — dasselbe Muster wie
// beim Ruhend-Kennzeichen. Der Stand der Einblendung fährt mit, damit die
// Antwort dieselbe Ansicht zeigt, aus der geklickt wurde.
func terminzeilen(listen []service.Teilnehmerliste, auchArchivierte bool) []terminzeile {
	zeilen := make([]terminzeile, 0, len(listen))
	for _, liste := range listen {
		t := liste.Termin

		werte := url.Values{}
		if !t.Archiviert {
			werte.Set(parameterArchivieren, "1")
		}
		if auchArchivierte {
			werte.Set(parameterArchivierte, "1")
		}

		pfad := fmt.Sprintf("/api/trainingstermin/%d/archiv", t.ID)
		if len(werte) > 0 {
			pfad += "?" + werte.Encode()
		}

		chips := make([]teilnehmerchip, 0, len(liste.Teilnehmer))
		for _, teilnehmer := range liste.Teilnehmer {
			chips = append(chips, teilnehmerchip{
				Teilnehmer: teilnehmer,
				Initialen:  initialenAus(teilnehmer.Vorname, teilnehmer.Nachname),
			})
		}

		zeilen = append(zeilen, terminzeile{Trainingstermin: t, Teilnehmer: chips, ArchivPfad: pfad})
	}

	return zeilen
}

// Die beiden Parameter dieser Ansicht. Sie sehen einander ähnlich und meinen
// Verschiedenes, deshalb stehen sie hier beieinander: das Adjektiv beschreibt,
// was die Liste zeigt, das Verb, was die Schaltfläche tun soll.
const (
	// parameterArchivierte blendet die archivierten Termine in der Liste ein.
	parameterArchivierte = "archivierte"
	// parameterArchivieren ist der gewünschte Zustand beim Schalten.
	parameterArchivieren = "archivieren"
)

// Anhaengsel ist der Zusatz, den jede Adresse dieser Ansicht braucht, um die
// Einblendung der archivierten Termine über eine Aktion hinweg zu behalten.
// Er steht hier und nicht im Template, damit die Schreibweise des Parameters
// nicht in fünf Attributen einzeln steht.
func (d trainingstermineDaten) Anhaengsel() string {
	return anhaengsel(d.AuchArchivierte)
}

// Anhaengsel ist derselbe Zusatz für das Formular: Speichern und Abbrechen
// führen damit in die Ansicht zurück, aus der es geöffnet wurde.
func (d terminformularDaten) Anhaengsel() string {
	return anhaengsel(d.AuchArchivierte)
}

func anhaengsel(auchArchivierte bool) string {
	if auchArchivierte {
		return "?" + parameterArchivierte + "=1"
	}

	return ""
}

// auchArchivierteLesen liest die Einblendung aus der Adresszeile. Ein
// Kontrollkästchen schickt seinen Wert nur, wenn es gesetzt ist — genau das
// macht "nur der gepflegte Stundenplan" zum Standard.
func auchArchivierteLesen(r *http.Request) bool {
	return r.URL.Query().Get(parameterArchivierte) != ""
}

// trainingstermineListe ist die Bereichsansicht: der Stundenplan in
// Wochenreihenfolge.
func (a *App) trainingstermineListe(w http.ResponseWriter, r *http.Request) {
	a.trainingstermineRendern(w, auchArchivierteLesen(r), meldung{})
}

// trainingstermineRendern ist die Rückkehr-Ansicht nach jeder Aktion.
func (a *App) trainingstermineRendern(w http.ResponseWriter, auchArchivierte bool, m meldung) {
	listen, err := a.svc.Teilnehmerlisten(auchArchivierte)
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.rendern(w, "trainingstermine", trainingstermineDaten{
		Termine:         terminzeilen(listen, auchArchivierte),
		AuchArchivierte: auchArchivierte,
		Meldung:         m,
		Navigation:      a.navigation(bereichTrainingstermine),
	})
}

// trainingsterminFormular liefert das leere Formular für einen neuen Termin.
func (a *App) trainingsterminFormular(w http.ResponseWriter, r *http.Request) {
	a.rendern(w, "trainingstermin-formular", terminformularDaten{
		Sprache:         a.Sprache(),
		AuchArchivierte: auchArchivierteLesen(r),
	})
}

func (a *App) trainingsterminAnlegen(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	auchArchivierte := auchArchivierteLesen(r)
	eingabe := terminEingabeLesen(r)

	id, err := a.svc.CreateTrainingstermin(eingabe.alsAngabe())
	if err == nil {
		a.trainingstermineRendern(w, auchArchivierte,
			a.terminMeldung(id, "trainingstermine.angelegt_mit_name", "trainingstermine.angelegt_ohne_name"))
		return
	}

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		fehlerAntwort(w, err)
		return
	}

	// Fehlerhafte Eingabe: Formular mit Werten und Meldungen zurückgeben. Wie
	// beim Mitgliedsformular bewusst mit Status 200 — htmx tauscht Antworten mit
	// Fehlerstatus standardmäßig nicht ein.
	a.rendern(w, "trainingstermin-formular", terminformularDaten{
		Sprache:         a.Sprache(),
		Eingabe:         eingabe,
		AuchArchivierte: auchArchivierte,
		Fehler:          validierung.Meldungen,
	})
}

// trainingsterminBearbeitenFormular liefert das mit dem gespeicherten Termin
// vorbefüllte Formular.
func (a *App) trainingsterminBearbeitenFormular(w http.ResponseWriter, r *http.Request) {
	id, ok := terminID(w, r)
	if !ok {
		return
	}

	termin, err := a.svc.GetTrainingstermin(id)
	if err != nil {
		a.terminNichtGefundenOderFehler(w, r, err)
		return
	}

	a.rendern(w, "trainingstermin-formular", terminformularDaten{
		Sprache:         a.Sprache(),
		Bearbeiten:      true,
		TerminID:        termin.ID,
		Eingabe:         terminEingabeAus(termin),
		Archiviert:      termin.Archiviert,
		AuchArchivierte: auchArchivierteLesen(r),
	})
}

func (a *App) trainingsterminAktualisieren(w http.ResponseWriter, r *http.Request) {
	id, ok := terminID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	auchArchivierte := auchArchivierteLesen(r)
	eingabe := terminEingabeLesen(r)

	err := a.svc.UpdateTrainingstermin(id, eingabe.alsAngabe())
	if err == nil {
		a.trainingstermineRendern(w, auchArchivierte,
			a.terminMeldung(id, "trainingstermine.gespeichert_mit_name", "trainingstermine.gespeichert_ohne_name"))
		return
	}

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		a.terminNichtGefundenOderFehler(w, r, err)
		return
	}

	// Ob der Termin archiviert ist, steht nicht im Formular und ist deshalb aus
	// der abgelehnten Eingabe nicht abzulesen. Einen zweiten Lesezugriff ist der
	// Hinweis nicht wert: fehlt er einmal, sagt ihn die Liste danach wieder.
	a.rendern(w, "trainingstermin-formular", terminformularDaten{
		Sprache:         a.Sprache(),
		Bearbeiten:      true,
		TerminID:        id,
		Eingabe:         eingabe,
		AuchArchivierte: auchArchivierte,
		Fehler:          validierung.Meldungen,
	})
}

// trainingsterminArchivSchalten nimmt einen Termin aus dem Stundenplan oder
// holt ihn zurück. Wie beim Ruhend-Kennzeichen schickt die Schaltfläche den
// gewünschten Wert mit und nicht den aktuellen: so tut ein Klick aus einer
// veralteten Ansicht nicht das Gegenteil.
//
// Zurück kommt die ganze Liste und nicht die Zeile: ein archivierter Termin
// verschwindet aus ihr, solange die Archivierten ausgeblendet sind.
func (a *App) trainingsterminArchivSchalten(w http.ResponseWriter, r *http.Request) {
	id, ok := terminID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	archiviert := r.FormValue(parameterArchivieren) != ""

	// Erst die Meldung bauen, dann schalten: sie nennt den Termin beim Namen,
	// und gelesen wird er dafür ohnehin.
	schluesselMitName, schluesselOhneName := "trainingstermine.archiviert_mit_name", "trainingstermine.archiviert_ohne_name"
	if !archiviert {
		schluesselMitName, schluesselOhneName = "trainingstermine.reaktiviert_mit_name", "trainingstermine.reaktiviert_ohne_name"
	}
	m := a.terminMeldung(id, schluesselMitName, schluesselOhneName)

	if err := a.svc.SetTrainingsterminArchiviert(id, archiviert); err != nil {
		a.terminNichtGefundenOderFehler(w, r, err)
		return
	}

	a.trainingstermineRendern(w, auchArchivierteLesen(r), m)
}

// terminMeldung nennt den Termin und was mit ihm geschehen ist. Gelesen wird er
// dafür noch einmal — die Anzeige entsteht im Service (Trainingstermin.Anzeige),
// damit in der Meldung dieselbe Schreibweise steht wie in der Liste. Scheitert
// das Lesen, bleibt die Meldung ohne Namen: sie ist eine Rückmeldung und kein
// Ergebnis, für das sich ein Fehlerbild lohnte — schluesselOhneName trägt
// diesen Fall als eigene, namenlose Formulierung.
func (a *App) terminMeldung(id int64, schluesselMitName, schluesselOhneName string) meldung {
	sprache := a.Sprache()

	termin, err := a.svc.GetTrainingstermin(id)
	if err != nil {
		return meldung{Text: i18n.Text(sprache, schluesselOhneName)}
	}

	return meldung{Text: i18n.Text(sprache, schluesselMitName, termin.Anzeige())}
}

// trainingsterminSerienmail bereitet die Serienmail an die Teilnehmerliste
// eines Trainingstermins vor (CONTEXT.md → Teilnehmerliste) — anders als bei
// der Mitgliederliste ohne eigene Auswahl: die Empfänger sind, wer gerade für
// den Termin angemeldet ist.
func (a *App) trainingsterminSerienmail(w http.ResponseWriter, r *http.Request) {
	id, ok := terminID(w, r)
	if !ok {
		return
	}

	emails, err := a.svc.TeilnehmerEmails(id)
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.rendern(w, "serienmail-ergebnis", service.SerienmailVorbereiten(emails))
}

// terminNichtGefundenOderFehler beantwortet einen Service-Fehler im
// Stundenplan. Eine ID, zu der es nichts mehr gibt, ist kein Serverfehler,
// sondern eine veraltete Ansicht — dann kehrt das Fragment zur Liste zurück und
// sagt, was los war.
func (a *App) terminNichtGefundenOderFehler(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, service.ErrNichtGefunden) {
		a.trainingstermineRendern(w, auchArchivierteLesen(r),
			meldung{Text: i18n.Text(a.Sprache(), "trainingstermine.nicht_gefunden"), Warnung: true})
		return
	}

	fehlerAntwort(w, err)
}
