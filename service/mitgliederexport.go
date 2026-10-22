package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Der Export der Mitgliederliste (CONTEXT.md → Sicherung, Abgrenzung) ist eine
// Excel-Datei zum Weitergeben und Auswerten: eine Zeile je Mitgliedschaft,
// ausgetretene eingeschlossen, die Angaben der Person in jeder ihrer Zeilen.
// Dokumente gehören nicht dazu, und zurückspielen lässt sich daraus nichts.

// MitgliederExportZeile ist eine Zeile des Exports.
type MitgliederExportZeile struct {
	MitgliedID       int64
	MitgliedschaftID int64

	Vorname      string
	Nachname     string
	Geburtsdatum *time.Time
	Anschrift    Anschrift
	Email        string
	Telefon      string
	Geschlecht   string

	GoogleBewertung GoogleBewertung
	Rueckstand      Rueckstand

	Kontoinhaber string
	IBAN         string

	Eintritt         time.Time
	Kuendigungsdatum *time.Time
	Austritt         *time.Time
	BeitragCents     int64

	AnmeldungDatum *time.Time
	GebuehrCents   int64

	// Trainingstermine sind die Anzeigetexte der angemeldeten Termine, durch
	// Zeilenumbruch getrennt — auch die archivierten, sie zählen zur
	// Vereinbarung (ADR-0008).
	Trainingstermine string
}

// MitgliederExport liest alle Mitgliedschaften aller Mitglieder. Es geht über
// Get und teilt damit dessen Lesart, statt sie in einer zweiten Abfrage
// nachzubauen; bei ein paar hundert Mitgliedern ist das ohne Belang.
func (s *MemberService) MitgliederExport() ([]MitgliederExportZeile, error) {
	reihen, err := s.db.Query(`SELECT id FROM mitglied ORDER BY nachname, vorname, id`)
	if err != nil {
		return nil, fmt.Errorf("mitglieder lesen: %w", err)
	}

	var ids []int64
	for reihen.Next() {
		var id int64
		if err := reihen.Scan(&id); err != nil {
			reihen.Close()
			return nil, fmt.Errorf("mitglied lesen: %w", err)
		}
		ids = append(ids, id)
	}
	if err := reihen.Err(); err != nil {
		reihen.Close()
		return nil, fmt.Errorf("mitglieder lesen: %w", err)
	}
	reihen.Close()

	var zeilen []MitgliederExportZeile
	for _, id := range ids {
		m, err := s.Get(id)
		if err != nil {
			return nil, fmt.Errorf("mitglied %d lesen: %w", id, err)
		}

		for _, ms := range m.Mitgliedschaften {
			termine := make([]string, len(ms.Trainingstermine))
			for i, t := range ms.Trainingstermine {
				termine[i] = t.Anzeige()
			}

			zeilen = append(zeilen, MitgliederExportZeile{
				MitgliedID: m.ID, MitgliedschaftID: ms.ID,
				Vorname: m.Vorname, Nachname: m.Nachname, Geburtsdatum: m.Geburtsdatum,
				Anschrift: m.Anschrift, Email: m.Email, Telefon: m.Telefon, Geschlecht: m.Geschlecht,
				GoogleBewertung: m.GoogleBewertung, Rueckstand: m.Rueckstand,
				Kontoinhaber: m.Kontoinhaber, IBAN: m.IBAN,
				Eintritt: ms.Eintritt, Kuendigungsdatum: ms.Kuendigungsdatum, Austritt: ms.Austritt,
				BeitragCents:   ms.BeitragCents,
				AnmeldungDatum: ms.Anmeldung.Datum, GebuehrCents: ms.Anmeldung.GebuehrCents,
				Trainingstermine: strings.Join(termine, "\n"),
			})
		}
	}

	return zeilen, nil
}

// mitgliederExcelBlatt ist der Name des einen Blatts der Datei.
const mitgliederExcelBlatt = "Mitglieder"

// mitgliederExcelKopf sind die Spaltenüberschriften in der Reihenfolge, in der
// mitgliederExcelZeile die Zellen liefert.
var mitgliederExcelKopf = []string{
	"Mitglieds-ID", "Vorname", "Nachname", "Geburtstag", "Adresse", "Postleitzahl", "Ort",
	"E-Mail", "Telefon", "Geschlecht", "Bewertung", "Rückstand", "Rückstand Notiz",
	"Kontoinhaber", "IBAN", "Eintritt", "Gekündigt", "Austritt", "Beitrag",
	"Anmeldedatum", "Anmeldegebühr", "Training",
}

// MitgliederExcel rendert die Zeilen als .xlsx. Datumsangaben stehen als
// Datumsangabe da (TT.MM.JJJJ), Beträge als Euro mit Komma.
func MitgliederExcel(zeilen []MitgliederExportZeile) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	if err := f.SetSheetName("Sheet1", mitgliederExcelBlatt); err != nil {
		return nil, fmt.Errorf("blatt benennen: %w", err)
	}

	kopf := make([]any, len(mitgliederExcelKopf))
	for i, k := range mitgliederExcelKopf {
		kopf[i] = k
	}
	if err := f.SetSheetRow(mitgliederExcelBlatt, "A1", &kopf); err != nil {
		return nil, fmt.Errorf("kopf schreiben: %w", err)
	}

	for i, z := range zeilen {
		zelle, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			return nil, fmt.Errorf("zelle berechnen: %w", err)
		}

		reihe := mitgliederExcelZeile(z)
		if err := f.SetSheetRow(mitgliederExcelBlatt, zelle, &reihe); err != nil {
			return nil, fmt.Errorf("zeile %d schreiben: %w", i+1, err)
		}
	}

	var puffer bytes.Buffer
	if err := f.Write(&puffer); err != nil {
		return nil, fmt.Errorf("datei erzeugen: %w", err)
	}

	return puffer.Bytes(), nil
}

func mitgliederExcelZeile(z MitgliederExportZeile) []any {
	rueckstand := "in Ordnung"
	if z.Rueckstand.Offen {
		rueckstand = "im Rückstand"
	}

	return []any{
		z.MitgliedID, z.Vorname, z.Nachname, datumText(z.Geburtsdatum),
		z.Anschrift.Adresse, z.Anschrift.Postleitzahl, z.Anschrift.Ort,
		z.Email, z.Telefon, z.Geschlecht, z.GoogleBewertung.Bezeichnung(),
		rueckstand, z.Rueckstand.Notiz,
		z.Kontoinhaber, z.IBAN,
		z.Eintritt.Format("02.01.2006"), datumText(z.Kuendigungsdatum), datumText(z.Austritt),
		BeitragAlsEuro(z.BeitragCents),
		datumText(z.AnmeldungDatum), BeitragAlsEuro(z.GebuehrCents),
		z.Trainingstermine,
	}
}

func datumText(d *time.Time) string {
	if d == nil {
		return ""
	}

	return d.Format("02.01.2006")
}
