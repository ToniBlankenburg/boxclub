package service_test

import (
	"bytes"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
	"github.com/xuri/excelize/v2"
)

// exportZeileZu findet die Zeile einer Mitgliedschaft — die Reihenfolge des
// Exports ist nicht Teil der Zusicherung.
func exportZeileZu(t *testing.T, zeilen []service.MitgliederExportZeile, mitgliedschaftID int64) service.MitgliederExportZeile {
	t.Helper()

	for _, z := range zeilen {
		if z.MitgliedschaftID == mitgliedschaftID {
			return z
		}
	}

	t.Fatalf("keine Exportzeile für Mitgliedschaft %d unter %d Zeilen", mitgliedschaftID, len(zeilen))
	return service.MitgliederExportZeile{}
}

func TestMitgliederExport_EineZeileJeMitgliedschaftAuchFuerAusgetretene(t *testing.T) {
	svc := neuerService(t)

	aktiv := moneyMoneyMitglied(t, svc, "Anna", "Berger", "DE02120300000000202051")
	ehemalig := moneyMoneyMitglied(t, svc, "Bruno", "Cordes", "DE02120300000000202052")
	austritt := heuteVersetzt(-10)
	if err := svc.SetKuendigung(ehemalig, service.Kuendigung{Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	zeilen, err := svc.MitgliederExport()
	if err != nil {
		t.Fatalf("MitgliederExport: %v", err)
	}
	if len(zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet 2 (aktiv und ausgetreten)", len(zeilen))
	}

	z := exportZeileZu(t, zeilen, laufendeMitgliedschaftID(t, svc, aktiv))
	if z.Vorname != "Anna" || z.Nachname != "Berger" {
		t.Errorf("Name = %q %q, erwartet Anna Berger", z.Vorname, z.Nachname)
	}
	if z.IBAN != "DE02120300000000202051" {
		t.Errorf("IBAN = %q", z.IBAN)
	}
	if z.BeitragCents != beitragImTest {
		t.Errorf("BeitragCents = %d, erwartet %d", z.BeitragCents, beitragImTest)
	}
	if z.Austritt != nil {
		t.Errorf("Austritt = %v, erwartet keinen", z.Austritt)
	}
}

func TestMitgliederExport_NenntTrainingstermineAlsText(t *testing.T) {
	svc := neuerService(t)

	terminID, err := svc.CreateTrainingstermin(service.Trainingsterminangabe{
		Wochentag: service.Montag, Beginn: "18:00", Ende: "19:30", Bezeichnung: "Anfänger",
	})
	if err != nil {
		t.Fatalf("CreateTrainingstermin: %v", err)
	}

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Anna", Nachname: "Berger", BeitragCents: beitragImTest,
		Eintritt: heuteVersetzt(-100), TrainingsterminIDs: []int64{terminID},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	zeilen, err := svc.MitgliederExport()
	if err != nil {
		t.Fatalf("MitgliederExport: %v", err)
	}

	z := exportZeileZu(t, zeilen, laufendeMitgliedschaftID(t, svc, id))
	if z.Trainingstermine != "Montag 18:00 – 19:30 · Anfänger" {
		t.Errorf("Trainingstermine = %q", z.Trainingstermine)
	}
}

func TestMitgliederExcel_LiefertLesbareDatei(t *testing.T) {
	svc := neuerService(t)
	moneyMoneyMitglied(t, svc, "Anna", "Berger", "DE02120300000000202051")

	zeilen, err := svc.MitgliederExport()
	if err != nil {
		t.Fatalf("MitgliederExport: %v", err)
	}

	daten, err := service.MitgliederExcel(zeilen)
	if err != nil {
		t.Fatalf("MitgliederExcel: %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(daten))
	if err != nil {
		t.Fatalf("Datei nicht lesbar: %v", err)
	}
	defer f.Close()

	reihen, err := f.GetRows("Mitglieder")
	if err != nil {
		t.Fatalf("GetRows: %v", err)
	}
	if len(reihen) != 2 {
		t.Fatalf("%d Reihen, erwartet Kopf und eine Zeile", len(reihen))
	}

	var gefunden bool
	for _, zelle := range reihen[1] {
		if zelle == "Berger" {
			gefunden = true
		}
	}
	if !gefunden {
		t.Errorf("Nachname steht nicht in der Datenzeile: %v", reihen[1])
	}
}
