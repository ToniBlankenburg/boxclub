package service_test

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// suchBestand ist der Mitgliederbestand, an dem die Such- und Filtertests
// arbeiten. Er ist bewusst klein und so geschnitten, dass jede Filterdimension
// für sich beobachtbar ist: beide Rückstandswerte, jede Trainingsfrequenz von
// null bis drei und je ein aktives wie ein ausgetretenes Mitglied kommen darin
// vor. Umlaute und ein Bindestrich stecken in den Namen; die Beiträge sind
// verschieden, damit eine vertauschte Zeile auffiele.
type suchBestand struct {
	// IDs der angelegten Mitglieder, benannt nach ihrem Nachnamen.
	Berger, Oeztuerk, MeierSchmidt, Wagner, Klein int64
}

func suchbestandAnlegen(t *testing.T, svc *service.MemberService) suchBestand {
	t.Helper()

	var b suchBestand

	anlegen := func(vorname, nachname, email, telefon string, beitragCents int64, slots ...string) int64 {
		t.Helper()

		id, err := svc.Create(service.NeuesMitglied{
			Vorname:        vorname,
			Nachname:       nachname,
			Email:          email,
			Telefon:        telefon,
			BeitragCents:   beitragCents,
			Eintritt:       datum(t, "2026-01-05"),
			Trainingsslots: slots,
		})
		if err != nil {
			t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
		}

		return id
	}

	imRueckstand := func(id int64, notiz string) {
		t.Helper()

		if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: notiz}); err != nil {
			t.Fatalf("SetRueckstand: %v", err)
		}
	}

	// Anna Berger und Jörg Meier-Schmidt bleiben in Ordnung — der Normalfall
	// beim Lastschrifteinzug, und zugleich der Nullwert nach der Anlage.
	//
	// Die Trainingsslots sind so verteilt, dass jede Frequenz genau einmal unter
	// den Aktiven vorkommt: Meier-Schmidt trainiert gar nicht, Öztürk einmal,
	// Berger zweimal, Wagner dreimal.
	b.Berger = anlegen("Anna", "Berger", "anna.berger@example.org", "030 1234567", 8000,
		"Montag 18:00 Uhr", "Mittwoch 19:30 Uhr")

	b.Oeztuerk = anlegen("Mehmet", "Öztürk", "m.oeztuerk@example.org", "0171 9876543", 6000,
		"Samstag 10:30 Uhr")
	imRueckstand(b.Oeztuerk, "Rücklastschrift Oktober")

	b.MeierSchmidt = anlegen("Jörg", "Meier-Schmidt", "joerg.meier-schmidt@example.org", "030 5550101", 0)

	b.Wagner = anlegen("Paul", "Wagner", "paul.wagner@example.org", "030 7778899", 13500,
		"Montag 18:00 Uhr", "Mittwoch 19:30 Uhr", "Samstag 10:30 Uhr")
	imRueckstand(b.Wagner, "Rücklastschrift September, angeschrieben am 05.10.")

	// Nina Klein ist ausgetreten — sonst würde sie in denselben Filter fallen
	// wie Wagner, und der Aktivitätsfilter bliebe unbewiesen. Ihre zwei Slots
	// teilt sie mit Berger: so trifft der Frequenzfilter allein noch nicht die
	// Aktivität.
	b.Klein = anlegen("Nina", "Klein", "nina.klein@example.org", "0160 4443322", 4500,
		"Dienstag 19:30 Uhr", "Samstag 10:30 Uhr")
	imRueckstand(b.Klein, "Rücklastschrift Juni, beim Austritt noch offen")
	if err := svc.MarkExit(b.Klein, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("MarkExit: %v", err)
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

// Der Rückstandsfilter ist zweiwertig plus "alle" — ein neutraler Zustand, den
// beide Stufen ausschließen würden, existiert nicht mehr (ADR-0006). Jedes
// Mitglied fällt deshalb in genau eine der beiden Stufen.
func TestSearch_FiltertNachRueckstand(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	faelle := []struct {
		name     string
		filter   service.Rueckstandsfilter
		erwartet []string
	}{
		{"alle", service.RueckstandsfilterAlle, alleAktiven},
		{"in Ordnung", service.RueckstandsfilterInOrdnung, []string{"Berger, Anna", "Meier-Schmidt, Jörg"}},
		{"im Rückstand", service.RueckstandsfilterImRueckstand, []string{"Öztürk, Mehmet", "Wagner, Paul"}},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			gefunden := suchen(t, svc, "", service.Suchfilter{Rueckstand: f.filter})
			if !slices.Equal(gefunden, f.erwartet) {
				t.Errorf("Rückstandsfilter %v = %v, erwartet %v", f.name, gefunden, f.erwartet)
			}
		})
	}

	// Die beiden Stufen zerlegen den Bestand vollständig: zusammen ergeben sie
	// wieder alle Aktiven, ohne Überschneidung.
	inOrdnung := suchen(t, svc, "", service.Suchfilter{Rueckstand: service.RueckstandsfilterInOrdnung})
	rueckstaendig := suchen(t, svc, "", service.Suchfilter{Rueckstand: service.RueckstandsfilterImRueckstand})
	if len(inOrdnung)+len(rueckstaendig) != len(alleAktiven) {
		t.Errorf("beide Stufen zusammen = %d Einträge, erwartet %d — es gibt keinen dritten Zustand",
			len(inOrdnung)+len(rueckstaendig), len(alleAktiven))
	}
}

// Der Frequenzfilter fragt nach einer abgeleiteten Größe: gespeichert sind nur
// die Slots (CONTEXT.md → Trainingsfrequenz). Er ersetzt den Klassenfilter, den
// ADR-0005 mit den Beitragsklassen abgeräumt hat.
func TestSearch_FiltertNachTrainingsfrequenz(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	faelle := []struct {
		name     string
		filter   service.Frequenzfilter
		erwartet []string
	}{
		{"alle", service.FrequenzfilterAlle, alleAktiven},
		{"1× pro Woche", service.FrequenzfilterEinmal, []string{"Öztürk, Mehmet"}},
		{"2× pro Woche", service.FrequenzfilterZweimal, []string{"Berger, Anna"}},
		{"3× pro Woche", service.FrequenzfilterDreimal, []string{"Wagner, Paul"}},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			gefunden := suchen(t, svc, "", service.Suchfilter{Frequenz: f.filter})
			if !slices.Equal(gefunden, f.erwartet) {
				t.Errorf("Frequenzfilter %s = %v, erwartet %v", f.name, gefunden, f.erwartet)
			}
		})
	}

	// Meier-Schmidt hat keinen Slot und fällt damit durch jede Stufe: null Slots
	// sind "keine Frequenz" und nicht "1×". Nur "alle" zeigt ihn.
	for _, f := range []service.Frequenzfilter{
		service.FrequenzfilterEinmal, service.FrequenzfilterZweimal, service.FrequenzfilterDreimal,
	} {
		gefunden := suchen(t, svc, "meier-schmidt", service.Suchfilter{Frequenz: f})
		if len(gefunden) != 0 {
			t.Errorf("Frequenzfilter %d = %v, erwartet kein Ergebnis für ein Mitglied ohne Slot", int(f), gefunden)
		}
	}
}

// Der Frequenzfilter greift mit Rückstand und Aktivität zusammen — alle drei
// Dimensionen müssen zutreffen.
func TestSearch_KombiniertFrequenzMitRueckstandUndAktivitaet(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	// Berger und Klein trainieren beide zweimal; Klein ist ausgetreten.
	gefunden := suchen(t, svc, "", service.Suchfilter{Frequenz: service.FrequenzfilterZweimal})
	if !slices.Equal(gefunden, []string{"Berger, Anna"}) {
		t.Errorf("2× aktiv = %v, erwartet [Berger, Anna]", gefunden)
	}

	gefunden = suchen(t, svc, "", service.Suchfilter{
		Frequenz:      service.FrequenzfilterZweimal,
		AuchEhemalige: true,
	})
	if !slices.Equal(gefunden, []string{"Berger, Anna", "Klein, Nina"}) {
		t.Errorf("2× inkl. Ehemaliger = %v, erwartet [Berger, Anna Klein, Nina]", gefunden)
	}

	// Von den beiden ist nur Klein im Rückstand.
	gefunden = suchen(t, svc, "", service.Suchfilter{
		Frequenz:      service.FrequenzfilterZweimal,
		Rueckstand:    service.RueckstandsfilterImRueckstand,
		AuchEhemalige: true,
	})
	if !slices.Equal(gefunden, []string{"Klein, Nina"}) {
		t.Errorf("2× im Rückstand inkl. Ehemaliger = %v, erwartet [Klein, Nina]", gefunden)
	}

	// Und der Suchbegriff muss zusätzlich treffen: Wagner trainiert dreimal und
	// ist im Rückstand, Öztürk nur einmal.
	dreimalImRueckstand := service.Suchfilter{
		Frequenz:   service.FrequenzfilterDreimal,
		Rueckstand: service.RueckstandsfilterImRueckstand,
	}
	gefunden = suchen(t, svc, "wagner", dreimalImRueckstand)
	if !slices.Equal(gefunden, []string{"Wagner, Paul"}) {
		t.Errorf("Search(\"wagner\", 3× im Rückstand) = %v, erwartet [Wagner, Paul]", gefunden)
	}
	if gefunden := suchen(t, svc, "öztürk", dreimalImRueckstand); len(gefunden) != 0 {
		t.Errorf("Search(\"öztürk\", 3× im Rückstand) = %v, erwartet kein Ergebnis", gefunden)
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

// Der Fall aus dem Ticket: "aktive Mitglieder im Rückstand".
func TestSearch_KombiniertSucheUndFilter(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	aktiveImRueckstand := service.Suchfilter{Rueckstand: service.RueckstandsfilterImRueckstand}

	// Klein passt in jede Filterdimension außer der Aktivität und muss deshalb
	// fehlen.
	gefunden := suchen(t, svc, "", aktiveImRueckstand)
	if !slices.Equal(gefunden, []string{"Öztürk, Mehmet", "Wagner, Paul"}) {
		t.Errorf("aktive im Rückstand = %v, erwartet [Öztürk, Mehmet Wagner, Paul]", gefunden)
	}

	// Derselbe Filter, jetzt auch mit den Ehemaligen: Klein kommt dazu.
	mitEhemaligen := aktiveImRueckstand
	mitEhemaligen.AuchEhemalige = true

	gefunden = suchen(t, svc, "", mitEhemaligen)
	if !slices.Equal(gefunden, []string{"Klein, Nina", "Öztürk, Mehmet", "Wagner, Paul"}) {
		t.Errorf("derselbe Filter inkl. Ehemaliger = %v, erwartet [Klein, Nina Öztürk, Mehmet Wagner, Paul]", gefunden)
	}

	// Query und Filter müssen beide zutreffen.
	gefunden = suchen(t, svc, "wagner", aktiveImRueckstand)
	if !slices.Equal(gefunden, []string{"Wagner, Paul"}) {
		t.Errorf("Search(\"wagner\", aktive im Rückstand) = %v, erwartet [Wagner, Paul]", gefunden)
	}

	// Berger trifft den Query, fällt aber durch den Rückstandsfilter.
	gefunden = suchen(t, svc, "berger", aktiveImRueckstand)
	if len(gefunden) != 0 {
		t.Errorf("Search(\"berger\", aktive im Rückstand) = %v, erwartet kein Ergebnis", gefunden)
	}
}

// Ein leerer Query grenzt nichts ein: das Ergebnis ist dasselbe, als wäre nur
// gefiltert worden — auch dann, wenn im Suchfeld nur Leerzeichen stehen.
func TestSearch_LeererQueryEntsprichtFilterOhneSuche(t *testing.T) {
	svc := neuerService(t)
	suchbestandAnlegen(t, svc)

	filter := []service.Suchfilter{
		{},
		{Rueckstand: service.RueckstandsfilterInOrdnung},
		{Rueckstand: service.RueckstandsfilterImRueckstand},
		{AuchEhemalige: true},
		{Rueckstand: service.RueckstandsfilterImRueckstand, AuchEhemalige: true},
		{Frequenz: service.FrequenzfilterZweimal},
		{Frequenz: service.FrequenzfilterZweimal, AuchEhemalige: true},
		{Frequenz: service.FrequenzfilterDreimal, Rueckstand: service.RueckstandsfilterImRueckstand},
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

	if !f.AuchEhemalige && e.Austritt != nil {
		t.Errorf("%s: ausgetreten, gehört ohne AuchEhemalige nicht ins Ergebnis", e.Nachname)
	}

	switch f.Rueckstand {
	case service.RueckstandsfilterInOrdnung:
		if e.Rueckstand.Offen {
			t.Errorf("%s: im Rückstand, erwartet in Ordnung", e.Nachname)
		}
	case service.RueckstandsfilterImRueckstand:
		if !e.Rueckstand.Offen {
			t.Errorf("%s: in Ordnung, erwartet im Rückstand", e.Nachname)
		}
	}

	if f.Frequenz != service.FrequenzfilterAlle && int(e.Trainingsfrequenz()) != int(f.Frequenz) {
		t.Errorf("%s: trainiert %d×, erwartet %d×", e.Nachname, int(e.Trainingsfrequenz()), int(f.Frequenz))
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
		Rueckstand:    service.RueckstandsfilterImRueckstand,
		AuchEhemalige: true,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(liste) != 0 {
		t.Errorf("Search = %+v, erwartet leer bei frischer Datenbank", liste)
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

// Die Mitglieds-ID ist die Nummer, unter der der Verein sein Mitglied kennt
// (CONTEXT.md → Mitglieds-ID). Sie direkt eintippen zu können ist der schnellste
// Weg zu einer bekannten Zeile.
//
// Anders als die übrigen Suchfelder trifft sie **genau** und nicht als
// Teilzeichenkette: bei 200 Mitgliedern würde eine einzelne Ziffer sonst ein
// Dutzend Zeilen zurückgeben, und die Nummer wäre als Sprungmarke wertlos.
func TestSearch_FindetMitgliedUeberSeineMitgliedsID(t *testing.T) {
	svc := neuerService(t)

	// Bewusst ohne E-Mail und Telefon: eine Ziffernfolge darf hier nur über die
	// ID treffen können, sonst belegte der Test nichts.
	namenDerReihe := []string{
		"Berger", "Öztürk", "Meier", "Wagner", "Klein", "Sommer",
		"Adler", "Vogel", "Brandt", "Kumar", "Petrov", "Delgado",
	}

	ids := make(map[string]int64, len(namenDerReihe))
	for _, nachname := range namenDerReihe {
		ids[nachname] = mitgliedAnlegen(t, svc, "Test", nachname)
	}

	for _, nachname := range namenDerReihe {
		id := ids[nachname]
		gefunden := suchen(t, svc, strconv.FormatInt(id, 10), service.Suchfilter{})
		if !slices.Equal(gefunden, []string{nachname + ", Test"}) {
			t.Errorf("Search(%d) = %v, erwartet [%s, Test]", id, gefunden, nachname)
		}
	}

	// Eine Nummer, die es nicht gibt, trifft nichts — und eine Teilzahl trifft
	// nicht die längeren Nummern, die mit ihr beginnen.
	for _, query := range []string{"99", "0", "4711"} {
		if gefunden := suchen(t, svc, query, service.Suchfilter{}); len(gefunden) != 0 {
			t.Errorf("Search(%q) = %v, erwartet kein Ergebnis", query, gefunden)
		}
	}
}

// Die ID-Suche hebelt die übrigen Filter nicht aus: sie ist ein zusätzliches
// Suchfeld, kein Sprung an der Ansicht vorbei.
func TestSearch_MitgliedsIDRespektiertDieFilter(t *testing.T) {
	svc := neuerService(t)

	// Ohne Telefon und E-Mail, damit eine Ziffer nur über die ID treffen kann.
	aktiv := mitgliedAnlegen(t, svc, "Paul", "Wagner")
	ehemalig := mitgliedAnlegen(t, svc, "Nina", "Klein")
	if err := svc.MarkExit(ehemalig, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}
	if err := svc.SetRueckstand(aktiv, service.Rueckstand{Offen: true, Notiz: "Rücklastschrift Oktober"}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	alsText := func(id int64) string { return strconv.FormatInt(id, 10) }

	// Die ID einer ausgetretenen Person findet sie nur, wenn die Ansicht
	// Ehemalige einschließt.
	if gefunden := suchen(t, svc, alsText(ehemalig), service.Suchfilter{}); len(gefunden) != 0 {
		t.Errorf("Search(%d) = %v, erwartet kein Ergebnis ohne AuchEhemalige", ehemalig, gefunden)
	}
	gefunden := suchen(t, svc, alsText(ehemalig), service.Suchfilter{AuchEhemalige: true})
	if !slices.Equal(gefunden, []string{"Klein, Nina"}) {
		t.Errorf("Search(%d, auch Ehemalige) = %v, erwartet [Klein, Nina]", ehemalig, gefunden)
	}

	// Und ebenso wenig am Rückstandsfilter vorbei.
	gefunden = suchen(t, svc, alsText(aktiv), service.Suchfilter{Rueckstand: service.RueckstandsfilterImRueckstand})
	if !slices.Equal(gefunden, []string{"Wagner, Paul"}) {
		t.Errorf("Search(%d, im Rückstand) = %v, erwartet [Wagner, Paul]", aktiv, gefunden)
	}
	gefunden = suchen(t, svc, alsText(aktiv), service.Suchfilter{Rueckstand: service.RueckstandsfilterInOrdnung})
	if len(gefunden) != 0 {
		t.Errorf("Search(%d, in Ordnung) = %v, erwartet kein Ergebnis", aktiv, gefunden)
	}
}
