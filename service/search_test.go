package service_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// suchBestand ist der Mitgliederbestand, an dem die Such- und Filtertests
// arbeiten. Er ist bewusst klein und so geschnitten, dass jede Filterdimension
// für sich beobachtbar ist: beide Beitragsklassen, alle drei Zahlungsstatus
// (inklusive "nicht gesetzt") und je ein aktives wie ein ausgetretenes Mitglied
// kommen darin vor. Umlaute und ein Bindestrich stecken in den Namen.
type suchBestand struct {
	EinMalWoche  service.Beitragsklasse
	ZweiMalWoche service.Beitragsklasse

	// IDs der angelegten Mitglieder, benannt nach ihrem Nachnamen.
	Berger, Oeztuerk, MeierSchmidt, Wagner, Klein int64
}

func suchbestandAnlegen(t *testing.T, svc *service.MemberService) suchBestand {
	t.Helper()

	klassen := beitragsklassen(t, svc)
	b := suchBestand{EinMalWoche: klassen[0], ZweiMalWoche: klassen[1]}

	anlegen := func(vorname, nachname, email, telefon string, klasse service.Beitragsklasse) int64 {
		t.Helper()

		id, err := svc.Create(service.NeuesMitglied{
			Vorname:          vorname,
			Nachname:         nachname,
			Email:            email,
			Telefon:          telefon,
			BeitragsklasseID: klasse.ID,
			Eintritt:         datum(t, "2026-01-05"),
		})
		if err != nil {
			t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
		}

		return id
	}

	bezahltBis := func(id int64, tage int) {
		t.Helper()

		d := heuteVersetzt(tage)
		if err := svc.SetBezahltBis(id, &d); err != nil {
			t.Fatalf("SetBezahltBis: %v", err)
		}
	}

	b.Berger = anlegen("Anna", "Berger", "anna.berger@example.org", "030 1234567", b.ZweiMalWoche)
	bezahltBis(b.Berger, 30)

	b.Oeztuerk = anlegen("Mehmet", "Öztürk", "m.oeztuerk@example.org", "0171 9876543", b.EinMalWoche)
	bezahltBis(b.Oeztuerk, -5)

	// Jörg Meier-Schmidt bleibt ohne Zahlungsangabe: "nicht gesetzt".
	b.MeierSchmidt = anlegen("Jörg", "Meier-Schmidt", "joerg.meier-schmidt@example.org", "030 5550101", b.ZweiMalWoche)

	b.Wagner = anlegen("Paul", "Wagner", "paul.wagner@example.org", "030 7778899", b.ZweiMalWoche)
	bezahltBis(b.Wagner, -3)

	// Nina Klein ist ausgetreten — sonst würde sie in denselben Filter fallen
	// wie Wagner, und der Aktivitätsfilter bliebe unbewiesen.
	b.Klein = anlegen("Nina", "Klein", "nina.klein@example.org", "0160 4443322", b.ZweiMalWoche)
	bezahltBis(b.Klein, -20)
	if err := svc.AustrittFuerTest(b.Klein, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("AustrittFuerTest: %v", err)
	}

	return b
}

// alleAktiven sind die Namen aller aktiven Mitglieder des Bestands, sortiert wie
// die Liste sie liefert.
var alleAktiven = []string{"Berger, Anna", "Meier-Schmidt, Jörg", "Öztürk, Mehmet", "Wagner, Paul"}

// namen verdichtet ein Suchergebnis auf "Nachname, Vorname" — die Reihenfolge
// bleibt dabei erhalten, denn die Sortierung gehört mit zum Ergebnis.
func namen(liste []service.Listeneintrag) []string {
	gefunden := make([]string, 0, len(liste))
	for _, e := range liste {
		gefunden = append(gefunden, e.Nachname+", "+e.Vorname)
	}

	return gefunden
}

// suchen ruft Search auf und verdichtet das Ergebnis auf die Namen.
func suchen(t *testing.T, svc *service.MemberService, query string, filter service.Suchfilter) []string {
	t.Helper()

	liste, err := svc.Search(query, filter)
	if err != nil {
		t.Fatalf("Search(%q, %+v): %v", query, filter, err)
	}

	return namen(liste)
}

func TestSearch_FindetTeiltrefferInNameEmailUndTelefon(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	faelle := []struct {
		name     string
		query    string
		erwartet []string
	}{
		{"Nachname", "berg", []string{"Berger, Anna"}},
		{"Vorname", "anna", []string{"Berger, Anna"}},
		{"Vorname in Großschreibung", "ANNA", []string{"Berger, Anna"}},
		{"Vorname exakt geschrieben", "Anna", []string{"Berger, Anna"}},
		{"E-Mail-Teil", "paul.wagner@", []string{"Wagner, Paul"}},
		{"E-Mail-Domain trifft alle", "@example.org", alleAktiven},
		{"Telefonnummer", "9876543", []string{"Öztürk, Mehmet"}},
		{"Telefon-Vorwahl trifft mehrere", "030 ", []string{"Berger, Anna", "Meier-Schmidt, Jörg", "Wagner, Paul"}},
		{"leerer Query trifft alle Aktiven", "", alleAktiven},
		{"nur Leerzeichen trifft alle Aktiven", "   ", alleAktiven},
		{"ohne Treffer", "gibtesnicht", nil},
		// Prozent und Unterstrich sind in SQL LIKE Platzhalter. Hier sind sie
		// gewöhnliche Zeichen und dürfen nicht plötzlich alles treffen.
		{"Prozentzeichen ist kein Platzhalter", "%", nil},
		{"Unterstrich ist kein Platzhalter", "_", nil},
		{"Platzhalter im Begriff", "b%r", nil},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			gefunden := suchen(t, svc, f.query, service.Suchfilter{})
			if !slices.Equal(gefunden, f.erwartet) {
				t.Errorf("Search(%q) = %v, erwartet %v", f.query, gefunden, f.erwartet)
			}
		})
	}
}

// Die Suche muss auch dort noch case-insensitiv treffen, wo die
// Groß-/Kleinschreibung nicht aus dem ASCII-Bereich stammt — ein deutscher
// Verein tippt "öztürk" so oft wie "Öztürk".
func TestSearch_FindetUmlauteUndBindestriche(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	faelle := []struct {
		name     string
		query    string
		erwartet []string
	}{
		{"Umlaut wie geschrieben", "Öztürk", []string{"Öztürk, Mehmet"}},
		{"Umlaut klein getippt", "öztürk", []string{"Öztürk, Mehmet"}},
		{"Umlaut groß getippt", "ÖZTÜRK", []string{"Öztürk, Mehmet"}},
		{"Umlaut im Vornamen", "jörg", []string{"Meier-Schmidt, Jörg"}},
		{"Doppelname mit Bindestrich", "meier-schmidt", []string{"Meier-Schmidt, Jörg"}},
		{"zweiter Namensteil", "-Schmidt", []string{"Meier-Schmidt, Jörg"}},
		{"Umlaut in der E-Mail-Umschrift", "oeztuerk", []string{"Öztürk, Mehmet"}},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			gefunden := suchen(t, svc, f.query, service.Suchfilter{})
			if !slices.Equal(gefunden, f.erwartet) {
				t.Errorf("Search(%q) = %v, erwartet %v", f.query, gefunden, f.erwartet)
			}
		})
	}
}

// Ein Mitglied ohne Zahlungsangabe ist nicht "nicht bezahlt" — es darf im
// Mahn-Filter nicht auftauchen (CONTEXT.md → Statusanzeige).
func TestSearch_FiltertNachZahlungsstatus(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	faelle := []struct {
		name     string
		filter   service.Zahlungsfilter
		erwartet []string
	}{
		{"alle", service.ZahlungsfilterAlle, alleAktiven},
		{"bezahlt", service.ZahlungsfilterBezahlt, []string{"Berger, Anna"}},
		{"nicht bezahlt", service.ZahlungsfilterNichtBezahlt, []string{"Öztürk, Mehmet", "Wagner, Paul"}},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			gefunden := suchen(t, svc, "", service.Suchfilter{Zahlungsstatus: f.filter})
			if !slices.Equal(gefunden, f.erwartet) {
				t.Errorf("Zahlungsfilter %v = %v, erwartet %v", f.name, gefunden, f.erwartet)
			}
		})
	}
}

func TestSearch_FiltertNachBeitragsklasse(t *testing.T) {
	svc := neuerService(t)
	b := suchbestandAnlegen(t, svc)

	faelle := []struct {
		name     string
		klasseID int64
		erwartet []string
	}{
		{"alle Klassen", 0, alleAktiven},
		{b.EinMalWoche.Name, b.EinMalWoche.ID, []string{"Öztürk, Mehmet"}},
		{b.ZweiMalWoche.Name, b.ZweiMalWoche.ID, []string{"Berger, Anna", "Meier-Schmidt, Jörg", "Wagner, Paul"}},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			gefunden := suchen(t, svc, "", service.Suchfilter{BeitragsklasseID: f.klasseID})
			if !slices.Equal(gefunden, f.erwartet) {
				t.Errorf("Beitragsklassenfilter %q = %v, erwartet %v", f.name, gefunden, f.erwartet)
			}
		})
	}
}

func TestSearch_ZeigtStandardmaessigNurAktiveUndAufWunschAuchEhemalige(t *testing.T) {
	svc := neuerService(t)
	b := suchbestandAnlegen(t, svc)

	gefunden := suchen(t, svc, "", service.Suchfilter{})
	if !slices.Equal(gefunden, alleAktiven) {
		t.Errorf("Standardansicht = %v, erwartet nur die Aktiven %v", gefunden, alleAktiven)
	}

	mitEhemaligen := []string{"Berger, Anna", "Klein, Nina", "Meier-Schmidt, Jörg", "Öztürk, Mehmet", "Wagner, Paul"}

	liste, err := svc.Search("", service.Suchfilter{AuchEhemalige: true})
	if err != nil {
		t.Fatalf("Search mit AuchEhemalige: %v", err)
	}
	if !slices.Equal(namen(liste), mitEhemaligen) {
		t.Errorf("Search mit AuchEhemalige = %v, erwartet %v", namen(liste), mitEhemaligen)
	}

	// Die ehemalige Mitgliedschaft trägt ihr Austrittsdatum mit; die laufenden
	// bleiben offen. Nur so lässt sich die Zeile als "ehemalig" kennzeichnen.
	austritt := datum(t, "2026-06-30")
	for _, e := range liste {
		switch {
		case e.MitgliedID == b.Klein:
			if e.Austritt == nil || !e.Austritt.Equal(austritt) {
				t.Errorf("Austritt von Klein = %v, erwartet %v", e.Austritt, austritt)
			}
		case e.Austritt != nil:
			t.Errorf("Austritt von %s %s = %v, erwartet nil (läuft noch)", e.Vorname, e.Nachname, e.Austritt)
		}
	}

	// Der Eintritt der beendeten Mitgliedschaft steht weiterhin in der Zeile.
	for _, e := range liste {
		if e.MitgliedID == b.Klein && !e.Eintritt.Equal(datum(t, "2026-01-05")) {
			t.Errorf("Eintritt von Klein = %v, erwartet 2026-01-05", e.Eintritt)
		}
	}
}

// Der Fall aus dem Ticket: "aktive Erwachsen 2×/Woche mit unbezahltem Status".
func TestSearch_KombiniertSucheUndFilter(t *testing.T) {
	svc := neuerService(t)
	b := suchbestandAnlegen(t, svc)

	aktiveZweiMalUnbezahlt := service.Suchfilter{
		Zahlungsstatus:   service.ZahlungsfilterNichtBezahlt,
		BeitragsklasseID: b.ZweiMalWoche.ID,
	}

	// Klein passt in jede Filterdimension außer der Aktivität und muss deshalb
	// fehlen; Öztürk ist unbezahlt, aber in der anderen Klasse.
	gefunden := suchen(t, svc, "", aktiveZweiMalUnbezahlt)
	if !slices.Equal(gefunden, []string{"Wagner, Paul"}) {
		t.Errorf("aktive 2×/Woche mit unbezahltem Status = %v, erwartet [Wagner, Paul]", gefunden)
	}

	// Dieselben Filter, jetzt auch mit den Ehemaligen: Klein kommt dazu.
	mitEhemaligen := aktiveZweiMalUnbezahlt
	mitEhemaligen.AuchEhemalige = true

	gefunden = suchen(t, svc, "", mitEhemaligen)
	if !slices.Equal(gefunden, []string{"Klein, Nina", "Wagner, Paul"}) {
		t.Errorf("dieselben Filter inkl. Ehemaliger = %v, erwartet [Klein, Nina Wagner, Paul]", gefunden)
	}

	// Query und Filter müssen beide zutreffen.
	gefunden = suchen(t, svc, "wagner", aktiveZweiMalUnbezahlt)
	if !slices.Equal(gefunden, []string{"Wagner, Paul"}) {
		t.Errorf("Search(\"wagner\", aktive 2× unbezahlt) = %v, erwartet [Wagner, Paul]", gefunden)
	}

	// Berger trifft den Query, fällt aber durch den Zahlungsfilter.
	gefunden = suchen(t, svc, "berger", aktiveZweiMalUnbezahlt)
	if len(gefunden) != 0 {
		t.Errorf("Search(\"berger\", aktive 2× unbezahlt) = %v, erwartet kein Ergebnis", gefunden)
	}
}

// Ein leerer Query grenzt nichts ein: das Ergebnis ist dasselbe, als wäre nur
// gefiltert worden — auch dann, wenn im Suchfeld nur Leerzeichen stehen.
func TestSearch_LeererQueryEntsprichtFilterOhneSuche(t *testing.T) {
	svc := neuerService(t)
	b := suchbestandAnlegen(t, svc)

	filter := []service.Suchfilter{
		{},
		{Zahlungsstatus: service.ZahlungsfilterBezahlt},
		{Zahlungsstatus: service.ZahlungsfilterNichtBezahlt},
		{BeitragsklasseID: b.EinMalWoche.ID},
		{BeitragsklasseID: b.ZweiMalWoche.ID},
		{AuchEhemalige: true},
		{Zahlungsstatus: service.ZahlungsfilterNichtBezahlt, BeitragsklasseID: b.ZweiMalWoche.ID, AuchEhemalige: true},
	}

	for _, f := range filter {
		ohneQuery := suchen(t, svc, "", f)

		for _, query := range []string{" ", "   ", "\t"} {
			mitLeerraum := suchen(t, svc, query, f)
			if !slices.Equal(mitLeerraum, ohneQuery) {
				t.Errorf("Search(%q, %+v) = %v, erwartet dasselbe wie ohne Query: %v",
					query, f, mitLeerraum, ohneQuery)
			}
		}

		// Und die Filter greifen dabei tatsächlich: jeder Treffer erfüllt sie.
		liste, err := svc.Search("", f)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		for _, e := range liste {
			pruefeFilter(t, e, f)
		}
	}
}

func pruefeFilter(t *testing.T, e service.Listeneintrag, f service.Suchfilter) {
	t.Helper()

	if f.BeitragsklasseID != 0 && e.Beitragsklasse.ID != f.BeitragsklasseID {
		t.Errorf("%s: Beitragsklasse %d, erwartet %d", e.Nachname, e.Beitragsklasse.ID, f.BeitragsklasseID)
	}
	if !f.AuchEhemalige && e.Austritt != nil {
		t.Errorf("%s: ausgetreten, gehört ohne AuchEhemalige nicht ins Ergebnis", e.Nachname)
	}

	switch f.Zahlungsstatus {
	case service.ZahlungsfilterBezahlt:
		if e.Zahlungsstatus != service.ZahlungsstatusBezahlt {
			t.Errorf("%s: Zahlungsstatus %v, erwartet bezahlt", e.Nachname, e.Zahlungsstatus)
		}
	case service.ZahlungsfilterNichtBezahlt:
		if e.Zahlungsstatus != service.ZahlungsstatusNichtBezahlt {
			t.Errorf("%s: Zahlungsstatus %v, erwartet nicht bezahlt", e.Nachname, e.Zahlungsstatus)
		}
	}
}

// List ist die Standardansicht: kein Query, keine Filter. Beide Wege müssen
// dasselbe liefern — sonst zeigte die Reset-Aktion etwas anderes als der
// erste Aufruf.
func TestSearch_MitStandardwertenEntsprichtList(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	gesucht, err := svc.Search("", service.Suchfilter{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(liste) != len(gesucht) {
		t.Fatalf("Search = %d Einträge, List = %d", len(gesucht), len(liste))
	}
	for i := range liste {
		// Verglichen wird über die Textform: der Eintrag trägt Zeiger auf
		// Datumswerte, die zwischen zwei Abfragen naturgemäß verschieden sind.
		if fmt.Sprintf("%+v", liste[i]) != fmt.Sprintf("%+v", gesucht[i]) {
			t.Errorf("Eintrag %d: Search = %+v, List = %+v", i, gesucht[i], liste[i])
		}
	}
}

// Ohne Mitglieder liefert jede Kombination ein leeres Ergebnis statt eines Fehlers.
func TestSearch_OhneMitgliederIstLeer(t *testing.T) {
	svc := neuerService(t)

	liste, err := svc.Search("berger", service.Suchfilter{
		Zahlungsstatus: service.ZahlungsfilterBezahlt,
		AuchEhemalige:  true,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(liste) != 0 {
		t.Errorf("Search = %+v, erwartet leer bei frischer Datenbank", liste)
	}
}

// Eine unbekannte Beitragsklasse ist kein Fehler, sondern schlicht kein Treffer.
func TestSearch_UnbekannteBeitragsklasseLiefertKeineTreffer(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	gefunden := suchen(t, svc, "", service.Suchfilter{BeitragsklasseID: 9999})
	if len(gefunden) != 0 {
		t.Errorf("Search mit unbekannter Beitragsklasse = %v, erwartet kein Ergebnis", gefunden)
	}
}

// Die Sortierung nach deutschen Regeln gilt auch für gefilterte Ergebnisse:
// Öztürk steht zwischen Meier-Schmidt und Wagner, nicht hinter beiden.
func TestSearch_SortiertErgebnisNachDeutschenRegeln(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	gefunden := suchen(t, svc, "@example.org", service.Suchfilter{AuchEhemalige: true})
	erwartet := []string{"Berger, Anna", "Klein, Nina", "Meier-Schmidt, Jörg", "Öztürk, Mehmet", "Wagner, Paul"}
	if !slices.Equal(gefunden, erwartet) {
		t.Errorf("Reihenfolge = %v, erwartet %v", gefunden, erwartet)
	}
}
