// Package app ist die Adapter-Schicht zwischen dem htmx-Frontend im WebView und
// dem MemberService. Die Handler sind bewusst dünn: Formularwerte parsen, an den
// Service delegieren, das Ergebnis als HTML-Fragment rendern. Fachlogik gehört
// ausschließlich in service/ (siehe ADR-0002).
package app

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
	"github.com/ToniBlankenburg/boxclub/templates"
)

// isoDatum ist das Format, das <input type="date"> liefert und erwartet.
const isoDatum = "2006-01-02"

// App bündelt die HTTP-Handler, die der Wails-Assetserver an das Frontend
// ausliefert.
type App struct {
	svc *service.MemberService
	tpl *template.Template
}

// New parst die Fragment-Templates und bindet sie an den übergebenen Service.
func New(svc *service.MemberService) (*App, error) {
	tpl, err := template.New("boxclub").Funcs(templateFunktionen).ParseFS(templates.FS, "*.html")
	if err != nil {
		return nil, fmt.Errorf("templates parsen: %w", err)
	}

	return &App{svc: svc, tpl: tpl}, nil
}

// Handler liefert das Routing für die htmx-Aufrufe des Frontends. Der
// Wails-Assetserver reicht alle Nicht-GET-Requests sowie GET-Requests, die
// keine statische Datei treffen, an diesen Handler durch.
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/mitglieder", a.mitgliederListe)
	mux.HandleFunc("GET /api/mitglieder/ergebnis", a.mitgliederErgebnis)
	mux.HandleFunc("GET /api/mitglied/formular", a.mitgliedFormular)
	mux.HandleFunc("POST /api/mitglied", a.mitgliedAnlegen)
	mux.HandleFunc("GET /api/mitglied/{id}/formular", a.mitgliedBearbeitenFormular)
	mux.HandleFunc("POST /api/mitglied/{id}", a.mitgliedAktualisieren)
	mux.HandleFunc("GET /api/mitglied/{id}/zeile", a.mitgliedZeile)
	mux.HandleFunc("GET /api/mitglied/{id}/rueckstand", a.rueckstandFormular)
	mux.HandleFunc("POST /api/mitglied/{id}/rueckstand", a.rueckstandSpeichern)
	mux.HandleFunc("GET /api/mitglied/{id}/kuendigung", a.kuendigungFormular)
	mux.HandleFunc("POST /api/mitglied/{id}/kuendigung", a.kuendigungEintragen)
	mux.HandleFunc("POST /api/mitglied/{id}/ruhend", a.ruhendSchalten)
	mux.HandleFunc("GET /api/mitglied/{id}/wiedereintritt", a.wiedereintrittFormular)
	mux.HandleFunc("POST /api/mitglied/{id}/wiedereintritt", a.wiedereintrittEintragen)
	mux.HandleFunc("GET /api/trainingstermine", a.trainingstermineListe)
	mux.HandleFunc("GET /api/trainingstermin/formular", a.trainingsterminFormular)
	mux.HandleFunc("POST /api/trainingstermin", a.trainingsterminAnlegen)
	mux.HandleFunc("GET /api/trainingstermin/{id}/formular", a.trainingsterminBearbeitenFormular)
	mux.HandleFunc("POST /api/trainingstermin/{id}", a.trainingsterminAktualisieren)
	mux.HandleFunc("POST /api/trainingstermin/{id}/archiv", a.trainingsterminArchivSchalten)
	mux.HandleFunc("GET /api/import", a.importFormular)
	mux.HandleFunc("POST /api/import", a.importAusfuehren)

	return mux
}

// Bereiche der App — die Ebene, auf der die Kopfzeilen-Navigation umschaltet.
const (
	bereichMitglieder       = "mitglieder"
	bereichTrainingstermine = "trainingstermine"
	bereichImport           = "import"
)

// navigationseintrag ist ein Eintrag der Bereichsnavigation. Schluessel ist der
// Bereich, den der Eintrag öffnet; er dient nur dem Vergleich in navigation und
// erscheint nicht in der Ausgabe.
type navigationseintrag struct {
	Schluessel   string
	Beschriftung string
	Pfad         string
	Aktiv        bool
}

// bereiche sind die Bereiche in der Reihenfolge, in der die Navigation sie
// zeigt. Wie bei den Filteroptionen steht die Liste in Go, damit Beschriftungen
// und Pfade nicht als Textliterale ins Template wandern.
var bereiche = []navigationseintrag{
	{Schluessel: bereichMitglieder, Beschriftung: "Mitglieder", Pfad: "/api/mitglieder"},
	{Schluessel: bereichTrainingstermine, Beschriftung: "Trainingstermine", Pfad: "/api/trainingstermine"},
	{Schluessel: bereichImport, Beschriftung: "Excel-Import", Pfad: "/api/import"},
}

// navigation liefert die Navigationseinträge mit dem angegebenen Bereich als
// aktivem — dasselbe Muster wie Rueckstandsoptionen: die feste Liste kopieren und
// darin markieren.
//
// Mitgeschickt wird sie von jeder Antwort, die eine ganze Bereichsansicht
// ersetzt. Antworten innerhalb eines Bereichs (Formulare, einzelne Zeilen)
// lassen sie weg; die Markierung in der Kopfzeile bleibt dann stehen, wie sie ist.
func navigation(aktiv string) []navigationseintrag {
	eintraege := slices.Clone(bereiche)
	for i := range eintraege {
		eintraege[i].Aktiv = eintraege[i].Schluessel == aktiv
	}

	return eintraege
}

// formularEingabe hält die Rohwerte des Formulars, damit eine fehlerhafte
// Eingabe beim erneuten Rendern nicht verloren geht.
//
// Geburtsdatum und Eintritt stehen nur beim Anlegen im Formular; beim
// Bearbeiten sind sie unveränderlich und werden über formularDaten nur
// angezeigt.
type formularEingabe struct {
	Vorname      string
	Nachname     string
	Geburtsdatum string
	// Die drei Felder der Anschrift kommen einzeln aus dem Formular und gehen
	// als service.Anschrift weiter (CONTEXT.md → Anschrift).
	Adresse      string
	Postleitzahl string
	Ort          string
	Email        string
	Telefon      string
	// IBAN, Geschlecht und Digital sind freier Text und gehen unverändert
	// weiter: geprüft wird an keinem der drei etwas (ADR-0006, Spec → Digital).
	IBAN       string
	Geschlecht string
	Digital    string
	// GoogleBewertung kommt aus einem Kontrollkästchen und ist deshalb schon
	// hier ein Wahrheitswert — parsen lässt sich daran nichts.
	GoogleBewertung bool
	// Beitrag ist der getippte Euro-Betrag, so wie er im Feld steht — umgerechnet
	// wird er im Service (service.BeitragAusEuro).
	Beitrag string
	// Anmeldegebuehr ist der getippte Euro-Betrag der einmaligen Gebühr; anders
	// als der Beitrag darf das Feld leer bleiben (service.AnmeldegebuehrAusEuro).
	Anmeldegebuehr string
	// Anmeldedatum und Eintritt stehen beide als ISO-Text im Formular. Der
	// Eintritt ist nach der Anlage nicht mehr änderbar, das Anmeldedatum schon
	// — es begrenzt keinen Zeitraum, sondern hält einen Vorgang fest.
	Anmeldedatum string
	Eintritt     string

	// Trainingsslots sind die Rohwerte der Slot-Felder — immer so viele, wie das
	// Formular Felder zeigt. Welche davon leer sind und damit kein Slot, und wie
	// viele es höchstens sein dürfen, entscheidet der Service.
	Trainingsslots []string
}

// trainingsslotfeld ist ein Slot-Feld des Formulars. Die Nummer steht in der
// Beschriftung und macht die Felder für die Sprachausgabe unterscheidbar.
type trainingsslotfeld struct {
	Nummer int
	Wert   string
}

// Trainingsslotfelder sind die Slot-Felder des Formulars: immer
// service.MaxTrainingsslots Stück, die leeren eingeschlossen. Mehr als drei
// Termine gibt es nicht, deshalb stehen von vornherein alle da — Hinzufügen
// heißt ein Feld ausfüllen, Entfernen heißt es leeren. Ein Hinzufügen-Knopf,
// der ein viertes Feld erzeugen könnte, wäre ein Versprechen, das der Service
// zu Recht bricht.
func (e formularEingabe) Trainingsslotfelder() []trainingsslotfeld {
	felder := make([]trainingsslotfeld, service.MaxTrainingsslots)
	for i := range felder {
		felder[i].Nummer = i + 1
		if i < len(e.Trainingsslots) {
			felder[i].Wert = e.Trainingsslots[i]
		}
	}

	return felder
}

// formularDaten speist das Formular-Template. Bearbeiten unterscheidet die
// beiden Modi: Neuanlage (POST auf /api/mitglied) und Änderung eines
// bestehenden Mitglieds (POST auf /api/mitglied/{id}).
type formularDaten struct {
	Bearbeiten bool
	MitgliedID int64
	Eingabe    formularEingabe

	// Anzeigewerte der Felder, die beim Bearbeiten festliegen — fertig
	// formatiert, weil sie nur gelesen und nicht zurückgeschickt werden.
	GeburtsdatumAnzeige string
	EintrittAnzeige     string

	Fehler []string
}

// meldung ist die Rückmeldung über eine gerade abgeschlossene Aktion. Warnung
// unterscheidet "hat geklappt" von "so nicht" — beides erscheint an derselben
// Stelle über der Liste, aber nicht in derselben Farbe.
type meldung struct {
	Text    string
	Warnung bool
}

// Werte des Rückstandsfilters, wie sie über die Adresszeile laufen. Sie stehen
// hier als Konstanten, damit Auswahlliste und Auswertung nicht auseinanderlaufen
// können.
const (
	rueckstandAlle         = ""
	rueckstandImRueckstand = "im-rueckstand"
	rueckstandInOrdnung    = "in-ordnung"
)

// frequenzAlle ist der Wert des Frequenzfilters für "nicht eingrenzen" — der
// leere Wert und damit der Standard. Die übrigen Werte sind die Frequenz als
// Ziffer und stehen deshalb nicht einzeln hier: sie entstehen aus derselben
// Zählung, aus der auch die Auswahlliste entsteht.
const frequenzAlle = ""

// suchEingabe hält die Rohwerte der Filterleiste. Der Nullwert ist die
// Standardansicht — dieselbe, die service.Suchfilter{} beschreibt.
type suchEingabe struct {
	Query      string
	Rueckstand string
	Frequenz   string
	Ehemalige  bool
}

// suchEingabeLesen sammelt Suchbegriff und Filter aus der Adresszeile ein.
func suchEingabeLesen(r *http.Request) suchEingabe {
	werte := r.URL.Query()

	return suchEingabe{
		Query: werte.Get("q"),
		// Ein Kontrollkästchen schickt seinen Wert nur, wenn es gesetzt ist.
		Ehemalige:  werte.Get("ehemalige") != "",
		Rueckstand: werte.Get("rueckstand"),
		Frequenz:   werte.Get("frequenz"),
	}
}

// alsSuchfilter übersetzt die Rohwerte in die Service-Eingabe. Was sich nicht
// zuordnen lässt, fällt auf den Standard zurück: solche Werte können nur aus
// einem selbstgebauten Request stammen, und eine Filterleiste ist kein Ort für
// Fehlermeldungen.
func (e suchEingabe) alsSuchfilter() service.Suchfilter {
	filter := service.Suchfilter{AuchEhemalige: e.Ehemalige}

	switch e.Rueckstand {
	case rueckstandImRueckstand:
		filter.Rueckstand = service.RueckstandsfilterImRueckstand
	case rueckstandInOrdnung:
		filter.Rueckstand = service.RueckstandsfilterInOrdnung
	}

	// Der Wert des Frequenzfilters ist die Frequenz als Ziffer. Alles andere —
	// "alle", Leerraum, ein selbstgebauter Request — bleibt beim Standard.
	if stufe, err := strconv.Atoi(e.Frequenz); err == nil && stufe >= 1 && stufe <= service.MaxTrainingsslots {
		filter.Frequenz = service.Frequenzfilter(stufe)
	}

	return filter
}

// filteroption ist ein Eintrag einer Auswahlliste. Die Liste wird in Go
// gebaut, damit die Wert-Konstanten nicht als Textliterale ins Template wandern.
type filteroption struct {
	Wert         string
	Beschriftung string
	Gewaehlt     bool
}

// rueckstandsoptionen sind die Stufen des Rückstandsfilters in der Reihenfolge,
// in der die Auswahlliste sie zeigt. Der erste Eintrag ist der Standard, auf den
// auch ein unbekannter Wert zurückfällt.
//
// "Im Rückstand" steht vor "In Ordnung": es ist die Ausnahmeliste, die der
// Verein tatsächlich abarbeitet (ADR-0006).
var rueckstandsoptionen = []filteroption{
	{Wert: rueckstandAlle, Beschriftung: "Alle Mitglieder"},
	{Wert: rueckstandImRueckstand, Beschriftung: "Im Rückstand"},
	{Wert: rueckstandInOrdnung, Beschriftung: "In Ordnung"},
}

// frequenzoptionen sind die Stufen des Frequenzfilters, eine je möglicher
// Frequenz. Sie werden gezählt und nicht aufgezählt, damit Wert, Beschriftung
// und Auswertung aus derselben Quelle kommen; die Beschriftung holt sich die
// Liste aus dem Service, wo das Vokabular für die Frequenz liegt.
//
// Eine Stufe für "keine Frequenz" gibt es nicht — gefragt wird nach den
// Trainierenden einer Frequenz, und wer keinen Slot hat, ist keine solche Gruppe.
var frequenzoptionen = frequenzoptionenBauen()

func frequenzoptionenBauen() []filteroption {
	optionen := []filteroption{{Wert: frequenzAlle, Beschriftung: "Jede Frequenz"}}
	for stufe := 1; stufe <= service.MaxTrainingsslots; stufe++ {
		optionen = append(optionen, filteroption{
			Wert:         strconv.Itoa(stufe),
			Beschriftung: service.Trainingsfrequenz(stufe).Bezeichnung(),
		})
	}

	return optionen
}

// listeDaten trägt das Suchergebnis, die Werte der Filterleiste und optional
// eine Rückmeldung.
type listeDaten struct {
	Eintraege  []service.Listeneintrag
	Meldung    meldung
	Suche      suchEingabe
	Navigation []navigationseintrag
}

// Gefiltert sagt, ob überhaupt eingegrenzt wurde. Ein leeres Ergebnis liest
// sich dann anders: "nichts gefunden" statt "noch nichts erfasst".
//
// Ein Suchbegriff aus lauter Leerraum zählt nicht — er grenzt auch im Service
// nichts ein.
func (d listeDaten) Gefiltert() bool {
	eingabe := d.Suche
	eingabe.Query = strings.TrimSpace(eingabe.Query)

	return eingabe != (suchEingabe{})
}

// Rueckstandsoptionen sind die Stufen des Rückstandsfilters, die gewählte
// darunter markiert.
func (d listeDaten) Rueckstandsoptionen() []filteroption {
	return gewaehlteOption(rueckstandsoptionen, d.Suche.Rueckstand)
}

// Frequenzoptionen sind die Stufen des Frequenzfilters, die gewählte darunter
// markiert.
func (d listeDaten) Frequenzoptionen() []filteroption {
	return gewaehlteOption(frequenzoptionen, d.Suche.Frequenz)
}

// gewaehlteOption kopiert eine Auswahlliste und markiert darin den Eintrag zum
// übergebenen Wert.
//
// Was sich nicht zuordnen lässt, steht auf dem ersten Eintrag — dieselbe Regel,
// nach der alsSuchfilter den Wert auswertet.
func gewaehlteOption(vorlage []filteroption, wert string) []filteroption {
	optionen := slices.Clone(vorlage)

	gewaehlt := 0
	for i, o := range optionen {
		if o.Wert == wert {
			gewaehlt = i
		}
	}
	optionen[gewaehlt].Gewaehlt = true

	return optionen
}

// mitgliederListe liefert die vollständige Listenansicht samt Filterleiste.
func (a *App) mitgliederListe(w http.ResponseWriter, r *http.Request) {
	a.listeMitFilterRendern(w, suchEingabeLesen(r), meldung{})
}

// mitgliederErgebnis liefert nur den Ergebnisteil. Suche und Filter tauschen
// ihn allein aus — die Filterleiste selbst bleibt stehen, sonst verlöre das
// Suchfeld bei jedem Tastendruck den Fokus.
func (a *App) mitgliederErgebnis(w http.ResponseWriter, r *http.Request) {
	daten, ok := a.listeDatenLesen(w, suchEingabeLesen(r), meldung{})
	if !ok {
		return
	}

	a.rendern(w, "mitglieder-ergebnis", daten)
}

// listeRendern ist die Rückkehr-Ansicht nach jeder Aktion: htmx tauscht das
// Listen-Fragment ein, ohne die Seite neu zu laden.
//
// Zurück geht es bewusst in die Standardansicht: das gerade geänderte Mitglied
// soll sichtbar sein und nicht hinter einem noch gesetzten Filter verschwinden.
func (a *App) listeRendern(w http.ResponseWriter, m meldung) {
	a.listeMitFilterRendern(w, suchEingabe{}, m)
}

func (a *App) listeMitFilterRendern(w http.ResponseWriter, eingabe suchEingabe, m meldung) {
	daten, ok := a.listeDatenLesen(w, eingabe, m)
	if !ok {
		return
	}

	a.rendern(w, "mitglieder-liste", daten)
}

// listeDatenLesen holt Ergebnis und Auswahllisten. Ist das Ergebnis nicht ok,
// wurde die Antwort bereits geschrieben.
func (a *App) listeDatenLesen(w http.ResponseWriter, eingabe suchEingabe, m meldung) (listeDaten, bool) {
	eintraege, err := a.svc.Search(eingabe.Query, eingabe.alsSuchfilter())
	if err != nil {
		fehlerAntwort(w, err)
		return listeDaten{}, false
	}

	return listeDaten{
		Eintraege:  eintraege,
		Meldung:    m,
		Suche:      eingabe,
		Navigation: navigation(bereichMitglieder),
	}, true
}

func (a *App) mitgliedFormular(w http.ResponseWriter, r *http.Request) {
	a.rendern(w, "mitglied-formular", formularDaten{
		Eingabe: formularEingabe{Eintritt: time.Now().Format(isoDatum)},
	})
}

func (a *App) mitgliedAnlegen(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	eingabe := formularEingabeLesen(r)

	neu, fehler := eingabe.alsNeuesMitglied()

	// Nur wenn die Rohwerte überhaupt parsebar waren, lohnt der Service-Aufruf.
	// Die Pflichtfeld-Regeln selbst liegen im Service; hier werden sie nur angezeigt.
	if len(fehler) == 0 {
		_, err := a.svc.Create(neu)
		if err == nil {
			a.listeRendern(w, meldung{Text: fmt.Sprintf("%s %s wurde angelegt.", neu.Vorname, neu.Nachname)})
			return
		}

		var validierung *service.ValidierungsFehler
		if !errors.As(err, &validierung) {
			fehlerAntwort(w, err)
			return
		}
		fehler = validierung.Meldungen
	}

	// Fehlerhafte Eingabe: Formular mit Werten und Meldungen zurückgeben. Bewusst
	// mit Status 200 — htmx tauscht Antworten mit Fehlerstatus standardmäßig nicht ein.
	a.rendern(w, "mitglied-formular", formularDaten{
		Eingabe: eingabe,
		Fehler:  fehler,
	})
}

// mitgliedBearbeitenFormular liefert das mit den aktuellen Stammdaten
// vorbefüllte Formular für ein bestehendes Mitglied.
func (a *App) mitgliedBearbeitenFormular(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	a.bearbeitenFormularRendern(w, id, nil, nil)
}

func (a *App) mitgliedAktualisieren(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	eingabe := formularEingabeLesen(r)

	patch, fehler := eingabe.alsPatch()

	if len(fehler) == 0 {
		err := a.svc.Update(id, patch)
		if err == nil {
			a.listeRendern(w, meldung{Text: fmt.Sprintf("%s %s wurde gespeichert.", eingabe.Vorname, eingabe.Nachname)})
			return
		}

		var validierung *service.ValidierungsFehler
		if !errors.As(err, &validierung) {
			a.nichtGefundenOderFehler(w, err)
			return
		}
		fehler = validierung.Meldungen
	}

	a.bearbeitenFormularRendern(w, id, &eingabe, fehler)
}

// bearbeitenFormularRendern rendert das Bearbeitungsformular. Ist eingabe nil,
// kommen die Werte aus dem gespeicherten Mitglied (erster Aufruf); andernfalls
// aus der abgelehnten Eingabe, damit die Tipparbeit nicht verloren geht.
func (a *App) bearbeitenFormularRendern(w http.ResponseWriter, id int64, eingabe *formularEingabe, fehler []string) {
	m, err := a.svc.Get(id)
	if err != nil {
		a.nichtGefundenOderFehler(w, err)
		return
	}

	// Welcher Zeitraum maßgeblich ist, entscheidet der Service — auch bei einem
	// ausgetretenen Mitglied, dessen letzter Eintritt und Beitrag hier stehen
	// sollen. Denselben Zeitraum ändert Update auch wieder.
	var (
		eintritt       *time.Time
		beitrag        string
		anmeldedatum   string
		anmeldegebuehr string
		slots          []string
	)
	if letzte := m.LetzteMitgliedschaft(); letzte != nil {
		eintritt = &letzte.Eintritt
		beitrag = service.BeitragAlsEuro(letzte.BeitragCents)
		anmeldedatum = isoDatumsWert(letzte.Anmeldung.Datum)
		anmeldegebuehr = service.AnmeldegebuehrAlsEuro(letzte.Anmeldung.GebuehrCents)
		slots = letzte.Trainingsslots
	}

	if eingabe == nil {
		eingabe = &formularEingabe{
			Vorname:         m.Vorname,
			Nachname:        m.Nachname,
			Adresse:         m.Anschrift.Adresse,
			Postleitzahl:    m.Anschrift.Postleitzahl,
			Ort:             m.Anschrift.Ort,
			Email:           m.Email,
			Telefon:         m.Telefon,
			IBAN:            m.IBAN,
			Geschlecht:      m.Geschlecht,
			GoogleBewertung: bool(m.GoogleBewertung),
			Digital:         m.Digital,
			Beitrag:         beitrag,
			Anmeldegebuehr:  anmeldegebuehr,
			Anmeldedatum:    anmeldedatum,
			Trainingsslots:  slots,
		}
	}

	a.rendern(w, "mitglied-formular", formularDaten{
		Bearbeiten:          true,
		MitgliedID:          m.ID,
		Eingabe:             *eingabe,
		GeburtsdatumAnzeige: datumAnzeige(m.Geburtsdatum),
		EintrittAnzeige:     datumAnzeige(eintritt),
		Fehler:              fehler,
	})
}

// mitgliedZeile liefert eine einzelne Listenzeile — der Rückweg aus der
// Rückstandseingabe, wenn der Nutzer abbricht.
func (a *App) mitgliedZeile(w http.ResponseWriter, r *http.Request) {
	eintrag, ok := a.zeileLesen(w, r)
	if !ok {
		return
	}

	a.rendern(w, "mitglied-zeile", eintrag)
}

// rueckstandFormular tauscht die Zeile gegen die Pflege von Kennzeichen und
// Notiz. Beides steht im selben Formular: Setzen, Aufheben und das Nachtragen
// der Notiz sind für den Nutzer derselbe Vorgang.
func (a *App) rueckstandFormular(w http.ResponseWriter, r *http.Request) {
	eintrag, ok := a.zeileLesen(w, r)
	if !ok {
		return
	}

	a.rendern(w, "mitglied-rueckstand-formular", eintrag)
}

// zeileLesen holt die Zeile zur ID aus dem Pfad. Ist das Ergebnis nicht ok,
// wurde die Antwort bereits geschrieben.
func (a *App) zeileLesen(w http.ResponseWriter, r *http.Request) (service.Listeneintrag, bool) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return service.Listeneintrag{}, false
	}

	return a.zeileLesenZuID(w, id)
}

// zeileLesenZuID ist dasselbe zur bereits bekannten ID — der Weg, den die
// Fehlerpfade nehmen: wer dorthin kommt, hat die ID schon aus dem Pfad gelesen
// und will sie nicht ein zweites Mal holen.
func (a *App) zeileLesenZuID(w http.ResponseWriter, id int64) (service.Listeneintrag, bool) {
	eintrag, err := a.svc.Eintrag(id)
	if err != nil {
		a.zeileNichtGefundenOderFehler(w, err)
		return service.Listeneintrag{}, false
	}

	return eintrag, true
}

// rueckstandSpeichern schreibt Kennzeichen und Notiz und antwortet mit der
// aktualisierten Zeile, die htmx an Ort und Stelle einwechselt.
//
// Hier kann nichts ungültig sein: das Kennzeichen ist ein Kontrollkästchen, die
// Notiz freier Text. Einen Fehlerpfad ins Formular zurück braucht es deshalb
// nicht.
func (a *App) rueckstandSpeichern(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	if err := a.svc.SetRueckstand(id, service.Rueckstand{
		// Ein Kontrollkästchen schickt seinen Wert nur, wenn es gesetzt ist.
		Offen: r.FormValue("rueckstand") != "",
		Notiz: r.FormValue("notiz"),
	}); err != nil {
		a.zeileNichtGefundenOderFehler(w, err)
		return
	}

	eintrag, err := a.svc.Eintrag(id)
	if err != nil {
		a.zeileNichtGefundenOderFehler(w, err)
		return
	}

	a.rendern(w, "mitglied-zeile", eintrag)
}

// kuendigungDaten speist die Zeile, in der eine Kündigung erfasst wird: zwei
// Datumsfelder nebeneinander — der Tag der Kündigung und der Tag, zu dem der
// Austritt wirksam wird.
type kuendigungDaten struct {
	Eintrag          service.Listeneintrag
	Kuendigungsdatum string
	Austritt         string
	Fehler           []string
}

// wiedereintrittDaten speist die Zeile, in der ein Wiedereintritt datiert wird.
// Hier genügt ein Datum: ein Eintritt beginnt einen Zeitraum, er beendet keinen.
type wiedereintrittDaten struct {
	Eintrag service.Listeneintrag
	Datum   string
	Fehler  []string
}

// Die Aktion steht in der Route und nicht im abgeschickten Formular: so trifft
// ein Klick aus einer veralteten Ansicht auf den Fehler des Service
// (ErrNichtAktiv bzw. ErrBereitsAktiv), statt still das Gegenteil zu tun.
func (a *App) kuendigungFormular(w http.ResponseWriter, r *http.Request) {
	eintrag, ok := a.zeileLesen(w, r)
	if !ok {
		return
	}

	a.rendern(w, "mitglied-kuendigung-formular", kuendigungsformular(eintrag, nil))
}

// kuendigungsformular belegt beide Eingaben vor: das Kündigungsdatum mit dem
// bereits erfassten, sonst mit dem heutigen Tag — gekündigt wird meist an dem
// Tag, an dem man es einträgt. Das Austrittsfeld zeigt den erfassten Termin
// und, wenn keiner erfasst ist, den regulären nach der Satzungsfrist
// (service.RegulaererAustritt).
//
// Der erfasste Termin hat dabei Vorrang, und das ist seit der Kündigungsfrist
// nötig: dort führt die Schaltfläche der Zeile zurück in dieses Formular, und
// SetKuendigung schreibt beide Spalten zusammen. Käme das Feld leer oder mit
// einem frisch gerechneten Wert herauf, überschriebe eine Korrektur am
// Kündigungsdatum den vereinbarten Austritt.
//
// Gerechnet wird einmal beim Öffnen und aus dem Datum, das oben steht. Ändert
// der Nutzer danach das Kündigungsdatum, zieht der Austritt **nicht** nach —
// dafür bräuchte es einen eigenen Endpunkt, und der überschriebe dann auch den
// Termin, den jemand gerade von Hand eingetippt hat.
//
// Bleiben soll der Termin trotzdem löschbar: ein geleertes Feld heißt weiterhin
// "Kündigung liegt vor, Termin noch offen" (siehe kuendigungLesen). Der
// Vorschlag nimmt dem Nutzer das Rechnen ab, nicht die Entscheidung.
func kuendigungsformular(eintrag service.Listeneintrag, fehler []string) kuendigungDaten {
	kuendigungsdatum := time.Now()
	if eintrag.Kuendigungsdatum != nil {
		kuendigungsdatum = *eintrag.Kuendigungsdatum
	}

	austritt := eintrag.Austritt
	if austritt == nil {
		regulaer := service.RegulaererAustritt(kuendigungsdatum)
		austritt = &regulaer
	}

	return kuendigungDaten{
		Eintrag:          eintrag,
		Kuendigungsdatum: kuendigungsdatum.Format(isoDatum),
		Austritt:         austritt.Format(isoDatum),
		Fehler:           fehler,
	}
}

// kuendigungEintragen erfasst die Kündigung. Welche Datumsangaben fachlich
// zulässig sind, entscheidet der Service; hier werden sie nur gelesen.
func (a *App) kuendigungEintragen(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	k, fehler := kuendigungLesen(r)
	if len(fehler) > 0 {
		a.kuendigungFormularMitFehler(w, id, fehler)
		return
	}

	if err := a.svc.SetKuendigung(id, k); err != nil {
		var validierung *service.ValidierungsFehler
		if errors.As(err, &validierung) {
			a.kuendigungFormularMitFehler(w, id, validierung.Meldungen)
			return
		}

		a.veralteteAnsichtOderFehler(w, err)
		return
	}

	eintrag, err := a.svc.Eintrag(id)
	if err != nil {
		a.veralteteAnsichtOderFehler(w, err)
		return
	}

	// Solange der Austritt nicht erreicht ist, endet nichts: das Mitglied bleibt
	// in der Standardansicht — auch mit erfasstem Termin, denn in der
	// Kündigungsfrist trainiert und zahlt es weiter. Die Zeile kommt dann mit
	// ihrem neuen Status zurück. Über beides entscheidet der Service, nicht das
	// Vorhandensein eines Datums.
	if !eintrag.Status().Ausgetreten() {
		a.rendern(w, "mitglied-zeile", eintrag)
		return
	}

	// Ist der Austritt dagegen schon erreicht — heute oder früher —, ist die
	// Zeile nicht mehr der richtige Platz für die Antwort: in der
	// Standardansicht gibt es sie nicht mehr. Deshalb kommt die ganze Liste
	// zurück und sagt zugleich, wo das Mitglied jetzt steht.
	//
	// k.Austritt ist hier gesetzt: ausgetreten wird nur, wer ein erreichtes
	// Austrittsdatum hat, und geschrieben hat es genau dieser Aufruf.
	aufListeUmleiten(w)
	a.listeRendern(w, meldung{Text: fmt.Sprintf(
		"Für %s %s ist der Austritt zum %s erfasst; die Zeile steht jetzt unter »Auch Ehemalige«.",
		eintrag.Vorname, eintrag.Nachname, datumAnzeige(*k.Austritt))})
}

// kuendigungLesen sammelt die beiden Datumsfelder ein. Ein leeres Feld heißt
// "nicht erfasst" — unlesbar ist deshalb nur, was gefüllt und trotzdem kein
// Datum ist. Ob auch beide zugleich leer bleiben dürfen, entscheidet der
// Service; diese Regel steht nur dort.
func kuendigungLesen(r *http.Request) (service.Kuendigung, []string) {
	var (
		k      service.Kuendigung
		fehler []string
	)

	if roh := r.FormValue("kuendigungsdatum"); roh != "" {
		d, err := time.Parse(isoDatum, roh)
		if err != nil {
			fehler = append(fehler, "Kündigungsdatum ist kein gültiges Datum.")
		} else {
			k.Datum = &d
		}
	}

	if roh := r.FormValue("austritt"); roh != "" {
		d, err := time.Parse(isoDatum, roh)
		if err != nil {
			fehler = append(fehler, "Austrittsdatum ist kein gültiges Datum.")
		} else {
			k.Austritt = &d
		}
	}

	return k, fehler
}

// kuendigungFormularMitFehler zeigt die Eingabe erneut, mitsamt den Meldungen
// des Service.
//
// Die abgelehnten Rohwerte werden bewusst nicht zurückgereicht: <input
// type="date"> zeigt einen Wert, der kein Datum ist, ohnehin nicht an — die
// Felder beginnen deshalb wieder bei ihrer Vorbelegung.
func (a *App) kuendigungFormularMitFehler(w http.ResponseWriter, id int64, fehler []string) {
	eintrag, ok := a.zeileLesenZuID(w, id)
	if !ok {
		return
	}

	a.rendern(w, "mitglied-kuendigung-formular", kuendigungsformular(eintrag, fehler))
}

// ruhendSchalten setzt das Ruhend-Kennzeichen oder nimmt es zurück und antwortet
// mit der aktualisierten Zeile. Ein Formular braucht das nicht: die Zeile zeigt
// den aktuellen Wert, die Schaltfläche schickt den gewünschten mit.
func (a *App) ruhendSchalten(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	if err := a.svc.SetRuhend(id, r.FormValue("ruhend") != ""); err != nil {
		a.veralteteAnsichtOderFehler(w, err)
		return
	}

	eintrag, err := a.svc.Eintrag(id)
	if err != nil {
		a.veralteteAnsichtOderFehler(w, err)
		return
	}

	a.rendern(w, "mitglied-zeile", eintrag)
}

// wiedereintrittFormular tauscht die Zeile gegen die Datumseingabe. Vorbelegt
// ist der heutige Tag — der häufigste Fall ist "ab sofort".
func (a *App) wiedereintrittFormular(w http.ResponseWriter, r *http.Request) {
	eintrag, ok := a.zeileLesen(w, r)
	if !ok {
		return
	}

	a.rendern(w, "mitglied-wiedereintritt-formular", wiedereintrittDaten{
		Eintrag: eintrag,
		Datum:   time.Now().Format(isoDatum),
	})
}

// wiedereintrittEintragen eröffnet den neuen Zeitraum. Ob das Datum fachlich
// zulässig ist, entscheidet der Service; hier wird es nur geparst.
func (a *App) wiedereintrittEintragen(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	d, err := time.Parse(isoDatum, r.FormValue("datum"))
	if err != nil {
		a.wiedereintrittFormularMitFehler(w, id, []string{"Das ist kein gültiges Datum."})
		return
	}

	if err := a.svc.Rejoin(id, d); err != nil {
		var validierung *service.ValidierungsFehler
		if errors.As(err, &validierung) {
			a.wiedereintrittFormularMitFehler(w, id, validierung.Meldungen)
			return
		}

		a.veralteteAnsichtOderFehler(w, err)
		return
	}

	eintrag, err := a.svc.Eintrag(id)
	if err != nil {
		a.veralteteAnsichtOderFehler(w, err)
		return
	}

	// Die Zeile stand unter »Auch Ehemalige« und gehört jetzt woandershin —
	// deshalb antwortet auch hier die ganze Liste.
	aufListeUmleiten(w)
	a.listeRendern(w, meldung{Text: fmt.Sprintf("%s %s ist zum %s wieder eingetreten.",
		eintrag.Vorname, eintrag.Nachname, datumAnzeige(d))})
}

// wiedereintrittFormularMitFehler zeigt die Datumseingabe erneut, mitsamt den
// Meldungen des Service — wie beim Kündigungsformular ohne den abgelehnten
// Rohwert.
func (a *App) wiedereintrittFormularMitFehler(w http.ResponseWriter, id int64, fehler []string) {
	eintrag, ok := a.zeileLesenZuID(w, id)
	if !ok {
		return
	}

	a.rendern(w, "mitglied-wiedereintritt-formular", wiedereintrittDaten{
		Eintrag: eintrag,
		Datum:   time.Now().Format(isoDatum),
		Fehler:  fehler,
	})
}

// veralteteAnsichtOderFehler beantwortet einen Fehler, der keine Frage des
// Datums ist: die Ansicht, aus der geklickt wurde, kannte den Lebenszyklus des
// Mitglieds nicht mehr richtig. Zurück geht es dann in die Liste, die den
// aktuellen Stand zeigt und sagt, was los war.
func (a *App) veralteteAnsichtOderFehler(w http.ResponseWriter, err error) {
	var text string

	switch {
	case errors.Is(err, service.ErrNichtGefunden):
		text = "Dieses Mitglied gibt es nicht mehr."
	case errors.Is(err, service.ErrNichtAktiv):
		text = "Dieses Mitglied ist bereits ausgetreten."
	case errors.Is(err, service.ErrBereitsAktiv):
		text = "Dieses Mitglied ist bereits aktiv."
	default:
		fehlerAntwort(w, err)
		return
	}

	aufListeUmleiten(w)
	a.listeRendern(w, meldung{Text: text, Warnung: true})
}

// zeileNichtGefundenOderFehler beantwortet einen Service-Fehler im
// Zeilen-Kontext. Gibt es die Zeile nicht mehr, wäre es falsch, ausgerechnet an
// ihrer Stelle etwas einzuwechseln.
func (a *App) zeileNichtGefundenOderFehler(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrNichtGefunden) {
		aufListeUmleiten(w)
	}

	a.nichtGefundenOderFehler(w, err)
}

// aufListeUmleiten gibt der Antwort ein neues Ziel: statt der Zeile, aus der die
// Aktion kam, ersetzt sie die ganze Liste. Nötig, wo die Zeile danach nicht mehr
// existiert — sonst landete eine ganze Liste im Tabellenzeilen-Element.
func aufListeUmleiten(w http.ResponseWriter) {
	w.Header().Set("HX-Retarget", "#inhalt")
	w.Header().Set("HX-Reswap", "innerHTML")
}

// formularEingabeLesen sammelt die Rohwerte des abgeschickten Formulars ein.
func formularEingabeLesen(r *http.Request) formularEingabe {
	return formularEingabe{
		Vorname:      r.FormValue("vorname"),
		Nachname:     r.FormValue("nachname"),
		Geburtsdatum: r.FormValue("geburtsdatum"),
		Adresse:      r.FormValue("adresse"),
		Postleitzahl: r.FormValue("postleitzahl"),
		Ort:          r.FormValue("ort"),
		Email:        r.FormValue("email"),
		Telefon:      r.FormValue("telefon"),
		IBAN:         r.FormValue("iban"),
		Geschlecht:   r.FormValue("geschlecht"),
		Digital:      r.FormValue("digital"),
		// Ein Kontrollkästchen schickt seinen Wert nur, wenn es gesetzt ist.
		GoogleBewertung: r.FormValue("google_bewertung") != "",
		Beitrag:         r.FormValue("beitrag"),
		Anmeldegebuehr:  r.FormValue("anmeldegebuehr"),
		Anmeldedatum:    r.FormValue("anmeldedatum"),
		Eintritt:        r.FormValue("eintritt"),
		// Alle gleichnamigen Slot-Felder auf einmal, in der Reihenfolge des
		// Formulars. Die leeren kommen mit; sie auszusortieren ist Sache des
		// Service, der auch die Obergrenze kennt.
		Trainingsslots: r.Form["trainingsslot"],
	}
}

// mitgliedID liest die Mitglied-ID aus dem Routen-Platzhalter.
func mitgliedID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return pfadID(w, r, "Mitglied-ID")
}

// terminID liest die Trainingstermin-ID aus dem Routen-Platzhalter.
func terminID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return pfadID(w, r, "Trainingstermin-ID")
}

// pfadID liest die ID aus dem Routen-Platzhalter. Ist das Ergebnis nicht ok,
// wurde die Antwort bereits geschrieben — eine ID, die keine Zahl ist, kann nur
// aus einem selbstgebauten Request stammen und ist deshalb, anders als eine
// unbekannte ID, tatsächlich ein Fehlerstatus wert.
func pfadID(w http.ResponseWriter, r *http.Request, was string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "ungültige "+was, http.StatusBadRequest)
		return 0, false
	}

	return id, true
}

// alsAnschrift bündelt die drei Adressfelder des Formulars. Anders als Beitrag
// und Datum ist hier nichts zu parsen — die Werte gehen unverändert weiter.
func (e formularEingabe) alsAnschrift() service.Anschrift {
	return service.Anschrift{
		Adresse:      e.Adresse,
		Postleitzahl: e.Postleitzahl,
		Ort:          e.Ort,
	}
}

// alsAnmeldung bündelt Anmeldedatum und Anmeldegebühr — die beiden Angaben, die
// an der Mitgliedschaft und nicht am Mitglied hängen. Beide sind freiwillig:
// ein leeres Datumsfeld heißt "nicht erfasst", ein leeres Gebührenfeld "keine
// erhoben". Unlesbar ist deshalb nur, was gefüllt und trotzdem kein Wert ist.
func (e formularEingabe) alsAnmeldung() (service.Anmeldung, []string) {
	var (
		anmeldung service.Anmeldung
		fehler    []string
	)

	if e.Anmeldedatum != "" {
		d, err := time.Parse(isoDatum, e.Anmeldedatum)
		if err != nil {
			fehler = append(fehler, "Anmeldedatum ist kein gültiges Datum.")
		} else {
			anmeldung.Datum = &d
		}
	}

	cents, err := service.AnmeldegebuehrAusEuro(e.Anmeldegebuehr)
	if err != nil {
		fehler = append(fehler, err.Error())
	}
	anmeldung.GebuehrCents = cents

	return anmeldung, fehler
}

// alsNeuesMitglied übersetzt die Rohwerte in die Service-Eingabe und sammelt
// dabei alle Parse-Fehler, statt beim ersten abzubrechen. Leere Pflichtfelder
// meldet nicht diese Funktion, sondern der Service — sonst stünde dieselbe Regel
// an zwei Stellen.
func (e formularEingabe) alsNeuesMitglied() (service.NeuesMitglied, []string) {
	var fehler []string

	anmeldung, anmeldungFehler := e.alsAnmeldung()
	fehler = append(fehler, anmeldungFehler...)

	neu := service.NeuesMitglied{
		Vorname:         e.Vorname,
		Nachname:        e.Nachname,
		Anschrift:       e.alsAnschrift(),
		Email:           e.Email,
		Telefon:         e.Telefon,
		IBAN:            e.IBAN,
		Geschlecht:      e.Geschlecht,
		GoogleBewertung: service.GoogleBewertung(e.GoogleBewertung),
		Digital:         e.Digital,
		Anmeldung:       anmeldung,
		Trainingsslots:  e.Trainingsslots,
	}

	if e.Geburtsdatum != "" {
		d, err := time.Parse(isoDatum, e.Geburtsdatum)
		if err != nil {
			fehler = append(fehler, "Geburtsdatum ist kein gültiges Datum.")
		} else {
			neu.Geburtsdatum = &d
		}
	}

	if e.Eintritt != "" {
		d, err := time.Parse(isoDatum, e.Eintritt)
		if err != nil {
			fehler = append(fehler, "Eintrittsdatum ist kein gültiges Datum.")
		} else {
			neu.Eintritt = d
		}
	}

	cents, err := service.BeitragAusEuro(e.Beitrag)
	if err != nil {
		fehler = append(fehler, err.Error())
	}
	neu.BeitragCents = cents

	return neu, fehler
}

// alsPatch übersetzt die Rohwerte in eine Änderung. Das Bearbeitungsformular
// schickt immer alle bearbeitbaren Felder, deshalb sind hier auch alle gesetzt;
// Geburtsdatum und Eintritt gehören bewusst nicht dazu.
func (e formularEingabe) alsPatch() (service.MitgliedPatch, []string) {
	var fehler []string

	anschrift := e.alsAnschrift()
	googleBewertung := service.GoogleBewertung(e.GoogleBewertung)
	patch := service.MitgliedPatch{
		Vorname:         &e.Vorname,
		Nachname:        &e.Nachname,
		Anschrift:       &anschrift,
		Email:           &e.Email,
		Telefon:         &e.Telefon,
		IBAN:            &e.IBAN,
		Geschlecht:      &e.Geschlecht,
		GoogleBewertung: &googleBewertung,
		Digital:         &e.Digital,
		// Das Formular schickt die Slot-Felder immer mit, auch die leeren: was
		// darin steht, ist die vollständige Aussage darüber, wann das Mitglied
		// künftig trainiert.
		Trainingsslots: &e.Trainingsslots,
	}

	// Was sich nicht lesen lässt, bleibt ungesetzt: ein halb verstandener Wert
	// stünde sonst stillschweigend auf 0 € bzw. auf "kein Datum", und beides ist
	// eine gültige Aussage, die niemand getroffen hat. Gemeldet werden beide
	// Betragsfelder auf einmal — wer sich in beiden vertippt hat, soll das nicht
	// nacheinander erfahren.
	anmeldung, anmeldungFehler := e.alsAnmeldung()
	if len(anmeldungFehler) > 0 {
		fehler = append(fehler, anmeldungFehler...)
	} else {
		patch.Anmeldung = &anmeldung
	}

	cents, err := service.BeitragAusEuro(e.Beitrag)
	if err != nil {
		fehler = append(fehler, err.Error())
	} else {
		patch.BeitragCents = &cents
	}

	return patch, fehler
}

func (a *App) rendern(w http.ResponseWriter, name string, daten any) {
	// Erst in einen Puffer rendern: schlägt das Template mittendrin fehl, ist
	// sonst bereits halbes HTML samt Status 200 unterwegs.
	var puffer bytes.Buffer
	if err := a.tpl.ExecuteTemplate(&puffer, name, daten); err != nil {
		fehlerAntwort(w, fmt.Errorf("template %q rendern: %w", name, err))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Jedes Fragment zeigt einen Stand aus der Datenbank und ist damit Daten,
	// kein Inhalt. Ohne diese Kopfzeile darf das WebView es zwischenspeichern
	// und liefert ein erneut geöffnetes Formular aus dem Cache — mit den Werten
	// von vor der letzten Änderung. Wer es dann speichert, schreibt den alten
	// Stand zurück; die Änderung ist weg, und einen Fehler hat niemand gesehen.
	//
	// Die Kopfzeile steht hier und nicht in den einzelnen Handlern: sie gilt für
	// jedes Fragment, und eine neue Route soll sie nicht erst wieder brauchen.
	w.Header().Set("Cache-Control", "no-store")

	puffer.WriteTo(w)
}

func fehlerAntwort(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// nichtGefundenOderFehler beantwortet einen Service-Fehler. Eine ID, zu der es
// nichts (mehr) gibt, ist kein Serverfehler, sondern eine veraltete Ansicht:
// dann kehrt das Fragment zur Liste zurück und sagt, was los war. Ein
// Fehlerstatus wäre hier das falsche Mittel — htmx tauscht solche Antworten
// nicht ein, und der Klick bliebe für den Nutzer wirkungslos.
func (a *App) nichtGefundenOderFehler(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrNichtGefunden) {
		a.listeRendern(w, meldung{Text: "Dieses Mitglied gibt es nicht mehr.", Warnung: true})
		return
	}

	fehlerAntwort(w, err)
}

// isoDatumsWert schreibt ein optionales Datum so, wie <input type="date"> es
// erwartet; nil wird zum leeren Feld. Das ist die Umkehrung des Parsens in
// formularEingabe.alsAnmeldung.
func isoDatumsWert(d *time.Time) string {
	if d == nil {
		return ""
	}

	return d.Format(isoDatum)
}

// datumAnzeige formatiert ein Datum deutsch; nil und der Nullwert werden zu "—".
func datumAnzeige(d any) string {
	switch v := d.(type) {
	case time.Time:
		if v.IsZero() {
			return "—"
		}
		return v.Format("02.01.2006")
	case *time.Time:
		if v == nil {
			return "—"
		}
		return datumAnzeige(*v)
	default:
		return "—"
	}
}

var templateFunktionen = template.FuncMap{
	// euro formatiert einen Cent-Betrag als "60,00 €" — dieselbe Schreibweise,
	// die das Formular zur Bearbeitung anbietet, nur mit Währungszeichen.
	"euro": func(cents int64) string {
		return service.BeitragAlsEuro(cents) + " €"
	},
	"datum": datumAnzeige,
	// feld bündelt die Argumente für das Teil-Template "feld"; html/template
	// kennt keine benannten Parameter.
	"feld": func(beschriftung, name, typ, wert string, pflicht, breit bool) feldDaten {
		return feldDaten{
			Beschriftung: beschriftung,
			Name:         name,
			Typ:          typ,
			Wert:         wert,
			Pflicht:      pflicht,
			Breit:        breit,
		}
	},
	// anzeigefeld bündelt die Argumente für das Teil-Template "feld-nur-lesen".
	"anzeigefeld": func(beschriftung, wert string) anzeigeDaten {
		return anzeigeDaten{Beschriftung: beschriftung, Wert: wert}
	},
	// vorschlagsfeld bündelt die Argumente für das Teil-Template
	// "feld-mit-vorschlaegen".
	"vorschlagsfeld": func(beschriftung, name, wert string, vorschlaege []string) vorschlagsfeldDaten {
		return vorschlagsfeldDaten{
			Beschriftung: beschriftung,
			Name:         name,
			Wert:         wert,
			Vorschlaege:  vorschlaege,
		}
	},
	// geschlechtVorschlaege reicht die Eintipphilfe aus dem Service ins
	// Template — die Werte stehen dort, wo das Vokabular liegt.
	"geschlechtVorschlaege": service.GeschlechtVorschlaege,
	// auswahlfeld bündelt die Argumente für das Teil-Template "feld-auswahl".
	"auswahlfeld": func(beschriftung, name string, optionen []filteroption, pflicht bool) auswahlfeldDaten {
		return auswahlfeldDaten{
			Beschriftung: beschriftung,
			Name:         name,
			Optionen:     optionen,
			Pflicht:      pflicht,
		}
	},
	// kontrollkaestchen bündelt die Argumente für das Teil-Template
	// "feld-kontrollkaestchen". Text beschreibt, was ein Haken bedeutet, und
	// kommt deshalb aus dem Service.
	"kontrollkaestchen": func(beschriftung, name, text string, gesetzt bool) kontrollkaestchenDaten {
		return kontrollkaestchenDaten{
			Beschriftung: beschriftung,
			Name:         name,
			Text:         text,
			Gesetzt:      gesetzt,
		}
	},
	// beschriftungHatBewertet ist der Text am Google-Haken: der Zustand, den ein
	// gesetzter Haken bedeutet. Er kommt aus dem Service, damit Liste und
	// Formular dieselben Worte benutzen (service.GoogleBewertung).
	"beschriftungHatBewertet": service.GoogleBewertung(true).Bezeichnung,
}

// vorschlagsfeldDaten beschreibt ein Freitextfeld mit Eintipphilfe für das
// Teil-Template "feld-mit-vorschlaegen". Die Vorschläge schränken nicht ein:
// gespeichert wird, was der Nutzer schreibt.
type vorschlagsfeldDaten struct {
	Beschriftung string
	Name         string
	Wert         string
	Vorschlaege  []string
}

// auswahlfeldDaten beschreibt eine Auswahlliste für das Teil-Template
// "feld-auswahl". Anders als die Vorschläge eines Freitextfelds schränkt sie
// ein: gespeichert wird ausschließlich, was in der Liste steht.
type auswahlfeldDaten struct {
	Beschriftung string
	Name         string
	Optionen     []filteroption
	Pflicht      bool
}

// kontrollkaestchenDaten beschreibt einen zweiwertigen Haken für das
// Teil-Template "feld-kontrollkaestchen".
type kontrollkaestchenDaten struct {
	Beschriftung string
	Name         string
	Text         string
	Gesetzt      bool
}

// anzeigeDaten beschreibt eine Angabe, die nur gelesen wird, für das
// Teil-Template "feld-nur-lesen".
type anzeigeDaten struct {
	Beschriftung string
	Wert         string
}

// feldDaten beschreibt ein einzelnes Eingabefeld für das Teil-Template "feld".
type feldDaten struct {
	Beschriftung string
	Name         string
	Typ          string
	Wert         string
	Pflicht      bool
	Breit        bool
}
