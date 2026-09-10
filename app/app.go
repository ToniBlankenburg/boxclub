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

	return mux
}

// formularEingabe hält die Rohwerte des Formulars, damit eine fehlerhafte
// Eingabe beim erneuten Rendern nicht verloren geht.
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

type formularDaten struct {
	Beitragsklassen []service.Beitragsklasse
	Eingabe         formularEingabe
	Fehler          []string
}

// listeDaten trägt die Mitgliederliste und optional eine Rückmeldung über eine
// gerade abgeschlossene Aktion.
type listeDaten struct {
	Eintraege []service.Listeneintrag
	Hinweis   string
}

func (a *App) mitgliederListe(w http.ResponseWriter, r *http.Request) {
	a.listeRendern(w, "")
}

// listeRendern ist die Rückkehr-Ansicht nach jeder Aktion: htmx tauscht das
// Listen-Fragment ein, ohne die Seite neu zu laden.
func (a *App) listeRendern(w http.ResponseWriter, hinweis string) {
	eintraege, err := a.svc.List()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.rendern(w, "mitglieder-liste", listeDaten{Eintraege: eintraege, Hinweis: hinweis})
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

	eingabe := formularEingabe{
		Vorname:          r.FormValue("vorname"),
		Nachname:         r.FormValue("nachname"),
		Geburtsdatum:     r.FormValue("geburtsdatum"),
		Adresse:          r.FormValue("adresse"),
		Email:            r.FormValue("email"),
		Telefon:          r.FormValue("telefon"),
		BeitragsklasseID: r.FormValue("beitragsklasse_id"),
		Eintritt:         r.FormValue("eintritt"),
	}

	neu, fehler := eingabe.alsNeuesMitglied()

	// Nur wenn die Rohwerte überhaupt parsebar waren, lohnt der Service-Aufruf.
	// Die Pflichtfeld-Regeln selbst liegen im Service; hier werden sie nur angezeigt.
	if len(fehler) == 0 {
		_, err := a.svc.Create(neu)
		if err == nil {
			a.listeRendern(w, fmt.Sprintf("%s %s wurde angelegt.", neu.Vorname, neu.Nachname))
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

	// Ein leeres Feld bleibt die Null-ID und fällt damit in die Pflichtfeld-Prüfung
	// des Service; nur ein unparsebarer Wert ist ein Adapter-Problem.
	if e.BeitragsklasseID != "" {
		id, err := strconv.ParseInt(e.BeitragsklasseID, 10, 64)
		if err != nil {
			fehler = append(fehler, "Bitte eine Beitragsklasse wählen.")
		} else {
			neu.BeitragsklasseID = id
		}
	}

	return neu, fehler
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

var templateFunktionen = template.FuncMap{
	// euro formatiert einen Cent-Betrag als "60,00 €".
	"euro": func(cents int64) string {
		return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
	},
	// datum formatiert ein Datum deutsch; nil und der Nullwert werden zu "—".
	"datum": func(d any) string {
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
			return v.Format("02.01.2006")
		default:
			return "—"
		}
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
}

// feldDaten beschreibt ein einzelnes Formularfeld für das Teil-Template "feld".
type feldDaten struct {
	Beschriftung string
	Name         string
	Typ          string
	Wert         string
	Pflicht      bool
	Breit        bool
}
