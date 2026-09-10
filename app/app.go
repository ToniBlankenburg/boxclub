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
	mux.HandleFunc("GET /api/mitglied/{id}/zahlung", a.zahlungFormular)
	mux.HandleFunc("POST /api/mitglied/{id}/zahlung", a.zahlungSpeichern)

	return mux
}

// formularEingabe hält die Rohwerte des Formulars, damit eine fehlerhafte
// Eingabe beim erneuten Rendern nicht verloren geht.
//
// Geburtsdatum und Eintritt stehen nur beim Anlegen im Formular; beim
// Bearbeiten sind sie unveränderlich und werden über formularDaten nur
// angezeigt.
type formularEingabe struct {
	Vorname          string
	Nachname         string
	Geburtsdatum     string
	Adresse          string
	Email            string
	Telefon          string
	BeitragsklasseID string
	Eintritt         string
}

// formularDaten speist das Formular-Template. Bearbeiten unterscheidet die
// beiden Modi: Neuanlage (POST auf /api/mitglied) und Änderung eines
// bestehenden Mitglieds (POST auf /api/mitglied/{id}).
type formularDaten struct {
	Bearbeiten      bool
	MitgliedID      int64
	Beitragsklassen []service.Beitragsklasse
	Eingabe         formularEingabe

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

// Werte des Zahlungsstatus-Filters, wie sie über die Adresszeile laufen. Sie
// stehen hier als Konstanten, damit Auswahlliste und Auswertung nicht
// auseinanderlaufen können.
const (
	statusAlle         = ""
	statusBezahlt      = "bezahlt"
	statusNichtBezahlt = "nicht-bezahlt"
)

// suchEingabe hält die Rohwerte der Filterleiste. Der Nullwert ist die
// Standardansicht — dieselbe, die service.Suchfilter{} beschreibt.
type suchEingabe struct {
	Query     string
	Status    string
	Klasse    string
	Ehemalige bool
}

// suchEingabeLesen sammelt Suchbegriff und Filter aus der Adresszeile ein.
func suchEingabeLesen(r *http.Request) suchEingabe {
	werte := r.URL.Query()

	return suchEingabe{
		Query: werte.Get("q"),
		// Ein Kontrollkästchen schickt seinen Wert nur, wenn es gesetzt ist.
		Ehemalige: werte.Get("ehemalige") != "",
		Status:    werte.Get("status"),
		Klasse:    werte.Get("klasse"),
	}
}

// alsSuchfilter übersetzt die Rohwerte in die Service-Eingabe. Was sich nicht
// zuordnen lässt, fällt auf den Standard zurück: solche Werte können nur aus
// einem selbstgebauten Request stammen, und eine Filterleiste ist kein Ort für
// Fehlermeldungen.
func (e suchEingabe) alsSuchfilter() service.Suchfilter {
	filter := service.Suchfilter{AuchEhemalige: e.Ehemalige}

	switch e.Status {
	case statusBezahlt:
		filter.Zahlungsstatus = service.ZahlungsfilterBezahlt
	case statusNichtBezahlt:
		filter.Zahlungsstatus = service.ZahlungsfilterNichtBezahlt
	}

	if id, err := strconv.ParseInt(e.Klasse, 10, 64); err == nil {
		filter.BeitragsklasseID = id
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

// zahlungsstatusoptionen sind die Stufen des Zahlungsstatus-Filters in der
// Reihenfolge, in der die Auswahlliste sie zeigt. Der erste Eintrag ist der
// Standard, auf den auch ein unbekannter Wert zurückfällt.
var zahlungsstatusoptionen = []filteroption{
	{Wert: statusAlle, Beschriftung: "Alle Zahlungsstatus"},
	{Wert: statusBezahlt, Beschriftung: "Bezahlt"},
	{Wert: statusNichtBezahlt, Beschriftung: "Nicht bezahlt"},
}

// listeDaten trägt das Suchergebnis, die Werte der Filterleiste und optional
// eine Rückmeldung.
type listeDaten struct {
	Eintraege       []service.Listeneintrag
	Meldung         meldung
	Suche           suchEingabe
	Beitragsklassen []service.Beitragsklasse
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

// Statusoptionen sind die Stufen des Zahlungsstatus-Filters, die gewählte
// darunter markiert.
func (d listeDaten) Statusoptionen() []filteroption {
	optionen := slices.Clone(zahlungsstatusoptionen)

	// Was sich nicht zuordnen lässt, steht auf dem Standard — dieselbe Regel,
	// nach der alsSuchfilter den Wert auswertet.
	gewaehlt := 0
	for i, o := range optionen {
		if o.Wert == d.Suche.Status {
			gewaehlt = i
		}
	}
	optionen[gewaehlt].Gewaehlt = true

	return optionen
}

// Klassenoptionen listet die Beitragsklassen so, wie die Datenbank sie führt —
// die Auswahl wächst also mit, wenn eine Klasse dazukommt.
func (d listeDaten) Klassenoptionen() []filteroption {
	optionen := []filteroption{{Beschriftung: "Alle Beitragsklassen", Gewaehlt: d.Suche.Klasse == ""}}

	for _, k := range d.Beitragsklassen {
		wert := strconv.FormatInt(k.ID, 10)
		optionen = append(optionen, filteroption{
			Wert:         wert,
			Beschriftung: k.Name,
			Gewaehlt:     d.Suche.Klasse == wert,
		})
	}

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

	klassen, err := a.svc.AktiveBeitragsklassen()
	if err != nil {
		fehlerAntwort(w, err)
		return listeDaten{}, false
	}

	return listeDaten{
		Eintraege:       eintraege,
		Meldung:         m,
		Suche:           eingabe,
		Beitragsklassen: klassen,
	}, true
}

func (a *App) mitgliedFormular(w http.ResponseWriter, r *http.Request) {
	klassen, err := a.svc.AktiveBeitragsklassen()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.rendern(w, "mitglied-formular", formularDaten{
		Beitragsklassen: klassen,
		Eingabe:         formularEingabe{Eintritt: time.Now().Format(isoDatum)},
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

	klassen, err := a.svc.AktiveBeitragsklassen()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	// Fehlerhafte Eingabe: Formular mit Werten und Meldungen zurückgeben. Bewusst
	// mit Status 200 — htmx tauscht Antworten mit Fehlerstatus standardmäßig nicht ein.
	a.rendern(w, "mitglied-formular", formularDaten{
		Beitragsklassen: klassen,
		Eingabe:         eingabe,
		Fehler:          fehler,
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

	klassen, err := a.svc.AktiveBeitragsklassen()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	if eingabe == nil {
		eingabe = &formularEingabe{
			Vorname:          m.Vorname,
			Nachname:         m.Nachname,
			Adresse:          m.Adresse,
			Email:            m.Email,
			Telefon:          m.Telefon,
			BeitragsklasseID: strconv.FormatInt(m.BeitragsklasseID, 10),
		}
	}

	// Angezeigt wird der Eintritt des Zeitraums, in dem das Mitglied gerade
	// aktiv ist; hat es keinen, bleibt das Feld leer.
	var eintritt *time.Time
	if laufend := m.LaufendeMitgliedschaft(); laufend != nil {
		eintritt = &laufend.Eintritt
	}

	a.rendern(w, "mitglied-formular", formularDaten{
		Bearbeiten:          true,
		MitgliedID:          m.ID,
		Beitragsklassen:     klassen,
		Eingabe:             *eingabe,
		GeburtsdatumAnzeige: datumAnzeige(m.Geburtsdatum),
		EintrittAnzeige:     datumAnzeige(eintritt),
		Fehler:              fehler,
	})
}

// zahlungDaten speist die Zeile, in der bezahlt_bis eingetragen wird.
type zahlungDaten struct {
	Eintrag service.Listeneintrag
	Fehler  string
}

// mitgliedZeile liefert eine einzelne Listenzeile — der Rückweg aus der
// Zahlungseingabe, wenn der Nutzer abbricht.
func (a *App) mitgliedZeile(w http.ResponseWriter, r *http.Request) {
	eintrag, ok := a.zeileLesen(w, r)
	if !ok {
		return
	}

	a.rendern(w, "mitglied-zeile", eintrag)
}

// zahlungFormular tauscht die Zeile gegen die Eingabe von bezahlt_bis.
func (a *App) zahlungFormular(w http.ResponseWriter, r *http.Request) {
	eintrag, ok := a.zeileLesen(w, r)
	if !ok {
		return
	}

	a.rendern(w, "mitglied-zahlung-formular", zahlungDaten{Eintrag: eintrag})
}

// zahlungFormularMitFehler zeigt die Eingabe erneut, mitsamt der Meldung.
//
// Der abgelehnte Rohwert wird bewusst nicht zurückgereicht: <input type="date">
// zeigt einen Wert, der kein Datum ist, ohnehin nicht an. Hierher kommt nur, wer
// den Request selbst gebaut hat — das Feld beginnt dann wieder beim
// gespeicherten Stand.
func (a *App) zahlungFormularMitFehler(w http.ResponseWriter, id int64, fehler string) {
	eintrag, err := a.svc.Eintrag(id)
	if err != nil {
		a.zeileNichtGefundenOderFehler(w, err)
		return
	}

	a.rendern(w, "mitglied-zahlung-formular", zahlungDaten{Eintrag: eintrag, Fehler: fehler})
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

// zahlungSpeichern schreibt bezahlt_bis und antwortet mit der aktualisierten
// Zeile, die htmx an Ort und Stelle einwechselt.
func (a *App) zahlungSpeichern(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	// Ein leeres Feld ist kein Fehler, sondern die Rücknahme der Angabe.
	var bezahltBis *time.Time
	if roh := r.FormValue("bezahlt_bis"); roh != "" {
		d, err := time.Parse(isoDatum, roh)
		if err != nil {
			a.zahlungFormularMitFehler(w, id, "Das ist kein gültiges Datum.")
			return
		}
		bezahltBis = &d
	}

	if err := a.svc.SetBezahltBis(id, bezahltBis); err != nil {
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

// zeileNichtGefundenOderFehler beantwortet einen Service-Fehler im
// Zeilen-Kontext. Gibt es die Zeile nicht mehr, wäre es falsch, ausgerechnet an
// ihrer Stelle etwas einzuwechseln: die Antwort bekommt per HX-Retarget ein
// neues Ziel und ersetzt die ganze Liste — sonst landete eine Liste im
// Tabellenzeilen-Element.
func (a *App) zeileNichtGefundenOderFehler(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrNichtGefunden) {
		w.Header().Set("HX-Retarget", "#inhalt")
		w.Header().Set("HX-Reswap", "innerHTML")
	}

	a.nichtGefundenOderFehler(w, err)
}

// formularEingabeLesen sammelt die Rohwerte des abgeschickten Formulars ein.
func formularEingabeLesen(r *http.Request) formularEingabe {
	return formularEingabe{
		Vorname:          r.FormValue("vorname"),
		Nachname:         r.FormValue("nachname"),
		Geburtsdatum:     r.FormValue("geburtsdatum"),
		Adresse:          r.FormValue("adresse"),
		Email:            r.FormValue("email"),
		Telefon:          r.FormValue("telefon"),
		BeitragsklasseID: r.FormValue("beitragsklasse_id"),
		Eintritt:         r.FormValue("eintritt"),
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

// alsNeuesMitglied übersetzt die Rohwerte in die Service-Eingabe und sammelt
// dabei alle Parse-Fehler, statt beim ersten abzubrechen. Leere Pflichtfelder
// meldet nicht diese Funktion, sondern der Service — sonst stünde dieselbe Regel
// an zwei Stellen.
func (e formularEingabe) alsNeuesMitglied() (service.NeuesMitglied, []string) {
	var fehler []string

	neu := service.NeuesMitglied{
		Vorname:  e.Vorname,
		Nachname: e.Nachname,
		Adresse:  e.Adresse,
		Email:    e.Email,
		Telefon:  e.Telefon,
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

	beitragsklasseID, fehlermeldung := e.beitragsklasseID()
	if fehlermeldung != "" {
		fehler = append(fehler, fehlermeldung)
	}
	neu.BeitragsklasseID = beitragsklasseID

	return neu, fehler
}

// alsPatch übersetzt die Rohwerte in eine Änderung. Das Bearbeitungsformular
// schickt immer alle bearbeitbaren Felder, deshalb sind hier auch alle gesetzt;
// Geburtsdatum und Eintritt gehören bewusst nicht dazu.
func (e formularEingabe) alsPatch() (service.MitgliedPatch, []string) {
	var fehler []string

	patch := service.MitgliedPatch{
		Vorname:  &e.Vorname,
		Nachname: &e.Nachname,
		Adresse:  &e.Adresse,
		Email:    &e.Email,
		Telefon:  &e.Telefon,
	}

	beitragsklasseID, fehlermeldung := e.beitragsklasseID()
	if fehlermeldung != "" {
		fehler = append(fehler, fehlermeldung)
	}
	patch.BeitragsklasseID = &beitragsklasseID

	return patch, fehler
}

// beitragsklasseID parst die gewählte Klasse. Ein leeres Feld bleibt die
// Null-ID und fällt damit in die Pflichtfeld-Prüfung des Service; nur ein
// unparsebarer Wert ist ein Adapter-Problem.
func (e formularEingabe) beitragsklasseID() (int64, string) {
	if e.BeitragsklasseID == "" {
		return 0, ""
	}

	id, err := strconv.ParseInt(e.BeitragsklasseID, 10, 64)
	if err != nil {
		return 0, "Bitte eine Beitragsklasse wählen."
	}

	return id, ""
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
	// euro formatiert einen Cent-Betrag als "60,00 €".
	"euro": func(cents int64) string {
		return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
	},
	"datum": datumAnzeige,
	// isodatum liefert den Wert für ein <input type="date">; nil wird zu "".
	"isodatum": func(d *time.Time) string {
		if d == nil {
			return ""
		}

		return d.Format(isoDatum)
	},
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
