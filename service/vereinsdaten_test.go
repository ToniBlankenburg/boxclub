package service_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// vereinsdatenImTest sind vollständig ausgefüllte Vereinsdaten — die Tests
// ändern davon nur, worum es ihnen geht.
func vereinsdatenImTest() service.Vereinsdaten {
	return service.Vereinsdaten{
		Name: "Boxclub Musterstadt e. V.",
		Anschrift: service.Anschrift{
			Adresse: "Kanalstraße 12", Postleitzahl: "12043", Ort: "Berlin",
		},
		Email:          "vorstand@example.org",
		Telefon:        "030 1234567",
		IBAN:           "DE02120300000000202051",
		BIC:            "BYLADEM1001",
		Kreditinstitut: "Musterbank",
		Fusszeile:      "Kein Ausweis von Umsatzsteuer\ngemäß § 19 UStG.",
	}
}

// Eine frische Datenbank hat die Zeile noch nicht — sie entsteht beim Öffnen
// (MemberService.Open). Dass sie danach da ist, zeigt dieser Test dadurch, dass
// das Lesen überhaupt gelingt: GetVereinsdaten kennt keinen Zweig für eine
// fehlende Zeile und käme sonst mit einem Fehler zurück.
//
// Darin steht nichts: die Vereinsdaten werden von Hand gepflegt und nirgends
// geseedet.
func TestGetVereinsdaten_FrischeDatenbankLiefertLeereAngaben(t *testing.T) {
	svc := neuerService(t)

	daten, err := svc.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}

	if !reflect.DeepEqual(daten, service.Vereinsdaten{}) {
		t.Errorf("Vereinsdaten = %+v, erwartet leer", daten)
	}
}

func TestSetVereinsdaten_SchreibtUndLiestZurueck(t *testing.T) {
	svc := neuerService(t)

	erwartet := vereinsdatenImTest()
	if err := svc.SetVereinsdaten(erwartet); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	daten, err := svc.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}
	if !reflect.DeepEqual(daten, erwartet) {
		t.Errorf("Vereinsdaten = %+v, erwartet %+v", daten, erwartet)
	}
}

// Gespeichert wird immer dieselbe eine Zeile: ein zweites Speichern ersetzt sie
// und legt keine zweite daneben. Nachsehen lässt sich das nur über das Lesen —
// das liefert genau eine Zeile, und zwar die zuletzt geschriebene.
func TestSetVereinsdaten_ZweitesSpeichernErsetztDieAngaben(t *testing.T) {
	svc := neuerService(t)

	if err := svc.SetVereinsdaten(vereinsdatenImTest()); err != nil {
		t.Fatalf("erstes Speichern: %v", err)
	}

	geaendert := vereinsdatenImTest()
	geaendert.Name = "Boxclub Musterstadt e. V. — Abteilung Nord"
	geaendert.IBAN = "DE02500105170137075030"
	if err := svc.SetVereinsdaten(geaendert); err != nil {
		t.Fatalf("zweites Speichern: %v", err)
	}

	daten, err := svc.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}
	if !reflect.DeepEqual(daten, geaendert) {
		t.Errorf("Vereinsdaten = %+v, erwartet %+v", daten, geaendert)
	}
}

// Jede Angabe ist freiwillig: was leer bleibt, fehlt später auf der Rechnung
// und blockiert nichts (Ticket 23).
func TestSetVereinsdaten_NimmtLeereAngabenAn(t *testing.T) {
	svc := neuerService(t)

	if err := svc.SetVereinsdaten(vereinsdatenImTest()); err != nil {
		t.Fatalf("erstes Speichern: %v", err)
	}

	// Alles wieder geleert — auch das ist ein gültiger Stand und kein Versehen,
	// das die App zurückweisen dürfte.
	if err := svc.SetVereinsdaten(service.Vereinsdaten{}); err != nil {
		t.Fatalf("Speichern leerer Angaben: %v", err)
	}

	daten, err := svc.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}
	if !reflect.DeepEqual(daten, service.Vereinsdaten{}) {
		t.Errorf("Vereinsdaten = %+v, erwartet leer", daten)
	}
}

// Die IBAN steht als reiner Text da — keine Prüfziffer, kein Format (ADR-0006).
// Was der Verein tippt, wird gespeichert, auch wenn es unvollständig ist.
func TestSetVereinsdaten_PrueftDieIBANNicht(t *testing.T) {
	svc := neuerService(t)

	daten := service.Vereinsdaten{IBAN: "DE02 1203 0000 0000 20"}
	if err := svc.SetVereinsdaten(daten); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	gelesen, err := svc.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}
	if gelesen.IBAN != daten.IBAN {
		t.Errorf("IBAN = %q, erwartet unverändert %q", gelesen.IBAN, daten.IBAN)
	}
}

// Umschließender Leerraum fällt weg — bei der Fußzeile also die leere Zeile am
// Ende, die beim Tippen im Textfeld entsteht. Innerhalb bleibt sie, wie sie ist:
// die Fußzeile ist mehrzeilig, und ihre Zeilenumbrüche sind die Angabe.
func TestSetVereinsdaten_SchneidetUmschliessendenLeerraumAb(t *testing.T) {
	svc := neuerService(t)

	if err := svc.SetVereinsdaten(service.Vereinsdaten{
		Name:      "  Boxclub Musterstadt e. V.  ",
		Fusszeile: "\nErste Zeile\nZweite Zeile\n\n",
	}); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	daten, err := svc.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}
	if daten.Name != "Boxclub Musterstadt e. V." {
		t.Errorf("Name = %q, erwartet ohne umschließenden Leerraum", daten.Name)
	}
	if daten.Fusszeile != "Erste Zeile\nZweite Zeile" {
		t.Errorf("Fußzeile = %q, erwartet ohne leere Zeilen davor und danach", daten.Fusszeile)
	}
}

// Beim zweiten Start steht die Zeile schon da und wird nicht überschrieben. Das
// ist die andere Hälfte von „legt sie an, falls sie fehlt": ein Anlegen, das
// sich nicht zurückhält, leerte bei jedem Start den Briefkopf des Vereins.
//
// Geprüft wird über zwei Läufe auf derselben Datei — dieselbe Strecke, die die
// App bei jedem Start nimmt.
func TestVereinsdaten_UeberlebenDenNeustart(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "boxclub.db")

	erster, err := service.Open(pfad)
	if err != nil {
		t.Fatalf("erster Start: %v", err)
	}
	if err := erster.SetVereinsdaten(vereinsdatenImTest()); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}
	if err := erster.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	zweiter, err := service.Open(pfad)
	if err != nil {
		t.Fatalf("zweiter Start: %v", err)
	}
	t.Cleanup(func() { zweiter.Close() })

	daten, err := zweiter.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}
	if !reflect.DeepEqual(daten, vereinsdatenImTest()) {
		t.Errorf("Vereinsdaten nach dem Neustart = %+v, erwartet unverändert %+v",
			daten, vereinsdatenImTest())
	}
}
