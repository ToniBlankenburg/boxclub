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
	mux.HandleFunc("GET /api/mitglied/{id}/austritt", a.austrittFormular)
	mux.HandleFunc("POST /api/mitglied/{id}/austritt", a.austrittEintragen)
	mux.HandleFunc("GET /api/mitglied/{id}/wiedereintritt", a.wiedereintrittFormular)
	mux.HandleFunc("POST /api/mitglied/{id}/wiedereintritt", a.wiedereintrittEintragen)

	return mux
}

// Bereiche der App — die Ebene, auf der die Kopfzeilen-Navigation umschaltet.
// Seit dem Wegfall der Beitragsklassen-Ansicht (ADR-0005) gibt es nur noch
// einen; die Umschaltung bleibt, weil die nächsten Bereiche anstehen.
const bereichMitglieder = "mitglieder"

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
	// Beitrag ist der getippte Euro-Betrag, so wie er im Feld steht — umgerechnet
	// wird er im Service (service.BeitragAusEuro).
	Beitrag  string
	Eintritt string
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

// suchEingabe hält die Rohwerte der Filterleiste. Der Nullwert ist die
// Standardansicht — dieselbe, die service.Suchfilter{} beschreibt.
type suchEingabe struct {
	Query      string
	Rueckstand string
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
	optionen := slices.Clone(rueckstandsoptionen)

	// Was sich nicht zuordnen lässt, steht auf dem Standard — dieselbe Regel,
	// nach der alsSuchfilter den Wert auswertet.
	gewaehlt := 0
	for i, o := range optionen {
		if o.Wert == d.Suche.Rueckstand {
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
		eintritt *time.Time
		beitrag  string
	)
	if letzte := m.LetzteMitgliedschaft(); letzte != nil {
		eintritt = &letzte.Eintritt
		beitrag = service.BeitragAlsEuro(letzte.BeitragCents)
	}

	if eingabe == nil {
		eingabe = &formularEingabe{
			Vorname:      m.Vorname,
			Nachname:     m.Nachname,
			Adresse:      m.Anschrift.Adresse,
			Postleitzahl: m.Anschrift.Postleitzahl,
			Ort:          m.Anschrift.Ort,
			Email:        m.Email,
			Telefon:      m.Telefon,
			Beitrag:      beitrag,
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

// mitgliedschaftDaten speist die Zeile, in der ein Aus- oder Wiedereintritt
// datiert wird. Beide teilen sich Formular und Handler; welche der beiden
// Aktionen gemeint ist, entscheidet allein Wiedereintritt.
type mitgliedschaftDaten struct {
	Eintrag        service.Listeneintrag
	Wiedereintritt bool
	Datum          string
	Fehler         []string
}

// Die Aktion steht in der Route und nicht im abgeschickten Formular: so trifft
// ein Klick aus einer veralteten Ansicht auf den Fehler des Service
// (ErrNichtAktiv bzw. ErrBereitsAktiv), statt still das Gegenteil zu tun.
func (a *App) austrittFormular(w http.ResponseWriter, r *http.Request) {
	a.mitgliedschaftFormular(w, r, false)
}

func (a *App) wiedereintrittFormular(w http.ResponseWriter, r *http.Request) {
	a.mitgliedschaftFormular(w, r, true)
}

func (a *App) austrittEintragen(w http.ResponseWriter, r *http.Request) {
	a.mitgliedschaftAendern(w, r, false)
}

func (a *App) wiedereintrittEintragen(w http.ResponseWriter, r *http.Request) {
	a.mitgliedschaftAendern(w, r, true)
}

// mitgliedschaftFormular tauscht die Zeile gegen die Datumseingabe. Vorbelegt
// ist der heutige Tag — der häufigste Fall ist "ab sofort".
func (a *App) mitgliedschaftFormular(w http.ResponseWriter, r *http.Request, wiedereintritt bool) {
	eintrag, ok := a.zeileLesen(w, r)
	if !ok {
		return
	}

	a.rendern(w, "mitglied-mitgliedschaft-formular", mitgliedschaftDaten{
		Eintrag:        eintrag,
		Wiedereintritt: wiedereintritt,
		Datum:          time.Now().Format(isoDatum),
	})
}

// mitgliedschaftAendern trägt den Aus- bzw. Wiedereintritt ein. Ob das Datum
// fachlich zulässig ist, entscheidet der Service; hier wird es nur geparst.
func (a *App) mitgliedschaftAendern(w http.ResponseWriter, r *http.Request, wiedereintritt bool) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	roh := r.FormValue("datum")

	d, err := time.Parse(isoDatum, roh)
	if err != nil {
		a.mitgliedschaftFormularMitFehler(w, id, wiedereintritt, []string{"Das ist kein gültiges Datum."})
		return
	}

	if wiedereintritt {
		err = a.svc.Rejoin(id, d)
	} else {
		err = a.svc.MarkExit(id, d)
	}
	if err != nil {
		var validierung *service.ValidierungsFehler
		if errors.As(err, &validierung) {
			a.mitgliedschaftFormularMitFehler(w, id, wiedereintritt, validierung.Meldungen)
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

	// Nach dem Austritt ist die Zeile nicht mehr der richtige Platz für die
	// Antwort: in der Standardansicht gibt es sie gar nicht mehr. Deshalb kommt
	// die ganze Liste zurück — und sagt zugleich, wo das Mitglied jetzt steht.
	text := fmt.Sprintf("%s %s ist zum %s ausgetreten und steht jetzt unter »Auch Ehemalige«.",
		eintrag.Vorname, eintrag.Nachname, datumAnzeige(d))
	if wiedereintritt {
		text = fmt.Sprintf("%s %s ist zum %s wieder eingetreten.",
			eintrag.Vorname, eintrag.Nachname, datumAnzeige(d))
	}

	aufListeUmleiten(w)
	a.listeRendern(w, meldung{Text: text})
}

// mitgliedschaftFormularMitFehler zeigt die Datumseingabe erneut, mitsamt den
// Meldungen des Service.
//
// Der abgelehnte Rohwert wird bewusst nicht zurückgereicht: <input type="date">
// zeigt einen Wert, der kein Datum ist, ohnehin nicht an — das Feld beginnt
// deshalb wieder beim heutigen Tag.
func (a *App) mitgliedschaftFormularMitFehler(w http.ResponseWriter, id int64, wiedereintritt bool, fehler []string) {
	eintrag, err := a.svc.Eintrag(id)
	if err != nil {
		a.zeileNichtGefundenOderFehler(w, err)
		return
	}

	a.rendern(w, "mitglied-mitgliedschaft-formular", mitgliedschaftDaten{
		Eintrag:        eintrag,
		Wiedereintritt: wiedereintritt,
		Datum:          time.Now().Format(isoDatum),
		Fehler:         fehler,
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
		Beitrag:      r.FormValue("beitrag"),
		Eintritt:     r.FormValue("eintritt"),
	}
}

// mitgliedID liest die Mitglied-ID aus dem Routen-Platzhalter. Ist das Ergebnis
// nicht ok, wurde die Antwort bereits geschrieben — eine ID, die keine Zahl ist,
// kann nur aus einem selbstgebauten Request stammen und ist deshalb, anders als
// eine unbekannte ID, tatsächlich ein Fehlerstatus wert.
func mitgliedID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "ungültige Mitglied-ID", http.StatusBadRequest)
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

// alsNeuesMitglied übersetzt die Rohwerte in die Service-Eingabe und sammelt
// dabei alle Parse-Fehler, statt beim ersten abzubrechen. Leere Pflichtfelder
// meldet nicht diese Funktion, sondern der Service — sonst stünde dieselbe Regel
// an zwei Stellen.
func (e formularEingabe) alsNeuesMitglied() (service.NeuesMitglied, []string) {
	var fehler []string

	neu := service.NeuesMitglied{
		Vorname:   e.Vorname,
		Nachname:  e.Nachname,
		Anschrift: e.alsAnschrift(),
		Email:     e.Email,
		Telefon:   e.Telefon,
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
	patch := service.MitgliedPatch{
		Vorname:   &e.Vorname,
		Nachname:  &e.Nachname,
		Anschrift: &anschrift,
		Email:     &e.Email,
		Telefon:   &e.Telefon,
	}

	// Ein unlesbarer Beitrag hält den ganzen Patch auf: er würde sonst
	// stillschweigend auf 0 € stehen bleiben, und 0 € ist ein gültiger Betrag.
	cents, err := service.BeitragAusEuro(e.Beitrag)
	if err != nil {
		return patch, append(fehler, err.Error())
	}
	patch.BeitragCents = &cents

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
