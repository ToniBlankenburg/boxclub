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
	"strconv"
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
	mux.HandleFunc("GET /api/mitglied/formular", a.mitgliedFormular)
	mux.HandleFunc("POST /api/mitglied", a.mitgliedAnlegen)
	mux.HandleFunc("GET /api/mitglied/{id}/formular", a.mitgliedBearbeitenFormular)
	mux.HandleFunc("POST /api/mitglied/{id}", a.mitgliedAktualisieren)

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

// listeDaten trägt die Mitgliederliste und optional eine Rückmeldung.
type listeDaten struct {
	Eintraege []service.Listeneintrag
	Meldung   meldung
}

func (a *App) mitgliederListe(w http.ResponseWriter, r *http.Request) {
	a.listeRendern(w, meldung{})
}

// listeRendern ist die Rückkehr-Ansicht nach jeder Aktion: htmx tauscht das
// Listen-Fragment ein, ohne die Seite neu zu laden.
func (a *App) listeRendern(w http.ResponseWriter, m meldung) {
	eintraege, err := a.svc.List()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.rendern(w, "mitglieder-liste", listeDaten{Eintraege: eintraege, Meldung: m})
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
	id, err := pfadID(r)
	if err != nil {
		http.Error(w, "ungültige Mitglied-ID", http.StatusBadRequest)
		return
	}

	a.bearbeitenFormularRendern(w, id, nil, nil)
}

func (a *App) mitgliedAktualisieren(w http.ResponseWriter, r *http.Request) {
	id, err := pfadID(r)
	if err != nil {
		http.Error(w, "ungültige Mitglied-ID", http.StatusBadRequest)
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

// pfadID liest die Mitglied-ID aus dem Routen-Platzhalter.
func pfadID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
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
