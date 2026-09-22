package service

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Der MoneyMoney-Export (ADR-0013) erzeugt eine CSV-Zeile je Mitgliedschaft,
// die diesen Monat tatsächlich Beitrag einzieht — damit stößt der Verein den
// Lastschrifteinzug über MoneyMoney an, statt ihn von Hand einzutippen. Die
// Spalten entsprechen dem Blatt "Money Money" der alten Excel-Tabelle:
// Mandatsreferenz, Zahlungspflichtiger, IBAN, BIC, Betrag, Verwendungszweck,
// Unterschrieben am.
//
// Das ist der eine Punkt, an dem ADR-0006 ("kein SEPA-Export") nicht mehr
// uneingeschränkt gilt — siehe ADR-0013 für die Begründung und die genaue
// Abgrenzung.

// MoneyMoneyZeile ist eine Zeile des Exports.
type MoneyMoneyZeile struct {
	// MitgliedID ist zugleich die Mandatsreferenz der Zeile — wie schon in der
	// alten Excel-Tabelle keine echte SEPA-Mandatsreferenz, sondern die
	// laufende Mitglieds-Nummer (CONTEXT.md → Mitglieds-ID).
	MitgliedID int64

	Zahlungspflichtiger string
	IBAN                string
	BetragCents         int64
	Verwendungszweck    string
	UnterschriebenAm    time.Time
}

// MoneyMoneyExport ist das Ergebnis eines Exportlaufs.
type MoneyMoneyExport struct {
	Zeilen []MoneyMoneyZeile

	// MitgliedschaftIDs sind die Zeiträume, deren Zeile die Anmeldegebühr
	// enthält. AnmeldegebuehrenAlsEingezogenMarkieren erwartet genau diese
	// Liste — aufzurufen, nachdem der Export tatsächlich gespeichert wurde,
	// nicht schon beim bloßen Erzeugen der Zeilen.
	MitgliedschaftIDs []int64
}

// moneyMoneyAbfrage liest die maßgebliche, laufende Mitgliedschaft je
// Mitglied — dieselbe Unterabfrage wie in eintraegeAbfrage, nur ohne die
// Ehemaligen: für sie geht keine Lastschrift mehr los, laeuftNoch schließt
// sie deshalb schon in der Abfrage aus statt erst beim Klassifizieren in Go.
const moneyMoneyAbfrage = `
	SELECT m.id, m.vorname, m.nachname, m.iban,
		ms.id, ms.eintritt, ms.kuendigungsdatum, ms.austritt, ms.ruhend,
		ms.beitrag_monatlich_cents,
		ms.anmeldedatum, ms.anmeldegebuehr_cents, ms.anmeldegebuehr_eingezogen
	FROM mitglied m
	JOIN mitgliedschaft ms ON ms.id = (
		SELECT id FROM mitgliedschaft
		WHERE mitglied_id = m.id AND ` + laeuftNoch + `
		ORDER BY eintritt DESC, id DESC
		LIMIT 1
	)`

// MoneyMoneyExportieren liefert die Zeilen des CSV-Exports: eine je
// Mitgliedschaft, die diesen Monat auch tatsächlich einzieht — dieselbe Menge
// wie im Monatssoll des Dashboards (aktiv und in Kündigungsfrist, ohne
// ruhende und ohne noch nicht begonnene, siehe Monatsuebersicht), nur als
// Einzelzeilen statt als Summe.
//
// Eine offene Anmeldegebühr reist in derselben Zeile mit: sie ist ein
// einmaliger Betrag, aber der Verein zieht nicht zweimal im selben Monat ein.
// "Offen" heißt hier, was Anmeldung.GebuehrEingezogen sagt — nicht ein
// bestimmter Kalendermonat, damit eine einmal verpasste Gebühr beim nächsten
// Lauf nicht verloren geht.
func (s *MemberService) MoneyMoneyExportieren() (MoneyMoneyExport, error) {
	verein, err := s.GetVereinsdaten()
	if err != nil {
		return MoneyMoneyExport{}, fmt.Errorf("vereinsdaten für export lesen: %w", err)
	}

	// jetzt liefert sowohl den Tag für die Abfrage (als isoDatum-Text) als
	// auch den Monat für den Verwendungszweck — beides aus demselben Zeitpunkt,
	// statt den Text danach wieder zurück in ein time.Time zu parsen.
	jetzt := time.Now()
	heute := jetzt.Format(isoDatum)

	rows, err := s.db.Query(moneyMoneyAbfrage, heute)
	if err != nil {
		return MoneyMoneyExport{}, fmt.Errorf("moneymoney-export lesen: %w", err)
	}
	defer rows.Close()

	var export MoneyMoneyExport
	for rows.Next() {
		var (
			mitgliedID, mitgliedschaftID int64
			vorname, nachname, iban      string
			eintrittText                 string
			kuendigungsdatum, austritt   sql.NullString
			ruhend                       bool
			beitragCents, gebuehrCents   int64
			anmeldedatum                 sql.NullString
			gebuehrEingezogen            bool
		)
		if err := rows.Scan(&mitgliedID, &vorname, &nachname, &iban,
			&mitgliedschaftID, &eintrittText, &kuendigungsdatum, &austritt, &ruhend,
			&beitragCents, &anmeldedatum, &gebuehrCents, &gebuehrEingezogen); err != nil {
			return MoneyMoneyExport{}, fmt.Errorf("moneymoney-zeile lesen: %w", err)
		}

		eintritt, err := time.Parse(isoDatum, eintrittText)
		if err != nil {
			return MoneyMoneyExport{}, fmt.Errorf("eintritt von mitgliedschaft %d: %w", mitgliedschaftID, err)
		}
		kuendigung, err := ausDatumsText(kuendigungsdatum)
		if err != nil {
			return MoneyMoneyExport{}, fmt.Errorf("kündigungsdatum von mitgliedschaft %d: %w", mitgliedschaftID, err)
		}
		austrittDatum, err := ausDatumsText(austritt)
		if err != nil {
			return MoneyMoneyExport{}, fmt.Errorf("austritt von mitgliedschaft %d: %w", mitgliedschaftID, err)
		}
		anmeldungsdatum, err := ausDatumsText(anmeldedatum)
		if err != nil {
			return MoneyMoneyExport{}, fmt.Errorf("anmeldedatum von mitgliedschaft %d: %w", mitgliedschaftID, err)
		}

		status := statusAus(eintritt, kuendigung, austrittDatum, heute)

		// Dieselbe Regel wie im Monatssoll (zaehltInsMonatssoll,
		// service/dashboard.go) — Ausgetreten kommt aus der Abfrage
		// (laeuftNoch) hier zwar gar nicht erst mit, aber die Funktion trotzdem
		// zu rufen hält beide Stellen an einer Regel fest statt an zweien.
		if !zaehltInsMonatssoll(status, ruhend) {
			continue
		}

		betragCents := beitragCents
		gebuehrEnthalten := gebuehrCents > 0 && !gebuehrEingezogen
		if gebuehrEnthalten {
			betragCents += gebuehrCents
			export.MitgliedschaftIDs = append(export.MitgliedschaftIDs, mitgliedschaftID)
		}

		unterschriebenAm := eintritt
		if anmeldungsdatum != nil {
			unterschriebenAm = *anmeldungsdatum
		}

		export.Zeilen = append(export.Zeilen, MoneyMoneyZeile{
			MitgliedID:          mitgliedID,
			Zahlungspflichtiger: strings.TrimSpace(vorname + " " + nachname),
			IBAN:                iban,
			BetragCents:         betragCents,
			Verwendungszweck: moneyMoneyVerwendungszweck(
				verein.Name, verein.MoneyMoneyVerwendungszweck, gebuehrEnthalten, jetzt),
			UnterschriebenAm: unterschriebenAm,
		})
	}
	if err := rows.Err(); err != nil {
		return MoneyMoneyExport{}, fmt.Errorf("moneymoney-export lesen: %w", err)
	}

	return export, nil
}

// moneyMoneyVerwendungszweck baut den Buchungstext: normalerweise Vereinsname
// und der Beitragsmonat, ergänzt um die Anmeldegebühr, wenn die Zeile sie
// enthält — genau wie im Blatt "Money Money" der alten Excel-Tabelle
// ("Anmeldegebühr + Beitrag 10/2026").
//
// Ist in den Vereinsdaten ein eigener Text hinterlegt (ADR-0016), tritt er an
// die Stelle von "Vereinsname Beitrag MM/JJJJ" — das "Anmeldegebühr + "-Präfix
// bleibt aber in jedem Fall automatisch: es ist Faktenstand dieser einen
// Zeile (ob die Mitgliedschaft eine offene Anmeldegebühr enthält), kein
// Stiltext, den der eigene Text überschreiben könnte.
func moneyMoneyVerwendungszweck(vereinsname, eigenerText string, gebuehrEnthalten bool, monat time.Time) string {
	zweck := eigenerText
	if zweck == "" {
		zweck = fmt.Sprintf("Beitrag %02d/%d", monat.Month(), monat.Year())
		if vereinsname != "" {
			zweck = vereinsname + " " + zweck
		}
	}

	if gebuehrEnthalten {
		zweck = "Anmeldegebühr + " + zweck
	}

	return zweck
}

// AnmeldegebuehrenAlsEingezogenMarkieren trägt für die angegebenen Zeiträume
// ein, dass ihre Anmeldegebühr eingezogen wurde. Aufzurufen ist das erst,
// nachdem ein MoneyMoney-Export (MoneyMoneyExport.MitgliedschaftIDs)
// tatsächlich gespeichert wurde — vorher gälte eine Gebühr als eingezogen,
// deren Datei nie irgendwo ankam.
func (s *MemberService) AnmeldegebuehrenAlsEingezogenMarkieren(mitgliedschaftIDs []int64) error {
	if len(mitgliedschaftIDs) == 0 {
		return nil
	}

	platzhalter := strings.Repeat(", ?", len(mitgliedschaftIDs)-1)
	werte := make([]any, len(mitgliedschaftIDs))
	for i, id := range mitgliedschaftIDs {
		werte[i] = id
	}

	if _, err := s.db.Exec(
		`UPDATE mitgliedschaft SET anmeldegebuehr_eingezogen = 1 WHERE id IN (?`+platzhalter+`)`,
		werte...); err != nil {
		return fmt.Errorf("anmeldegebühren als eingezogen markieren: %w", err)
	}

	return nil
}

// MoneyMoneyCSV setzt die Zeilen des Exports als CSV — Semikolon-getrennt,
// wie im deutschsprachigen Raum üblich und hier auch nötig, weil der Betrag
// selbst ein Komma als Dezimaltrennzeichen trägt (BeitragAlsEuro). Kopfzeile
// und Spaltenreihenfolge entsprechen dem Blatt "Money Money" der alten
// Excel-Tabelle.
func MoneyMoneyCSV(export MoneyMoneyExport) ([]byte, error) {
	var puffer bytes.Buffer

	schreiber := csv.NewWriter(&puffer)
	schreiber.Comma = ';'
	// RFC 4180 verlangt CRLF als Zeilenende; Go schreibt ohne diese Zeile nur
	// LF. Ob MoneyMoney das eine so viel eher liest wie das andere, ist nicht
	// bekannt — CRLF ist aber der Standard und die sicherere Wahl.
	schreiber.UseCRLF = true

	if err := schreiber.Write([]string{
		"Mandatsreferenz", "Zahlungspflichtiger", "IBAN", "BIC",
		"Betrag", "Verwendungszweck", "Unterschrieben am",
	}); err != nil {
		return nil, fmt.Errorf("moneymoney-csv kopfzeile schreiben: %w", err)
	}

	for _, z := range export.Zeilen {
		if err := schreiber.Write([]string{
			strconv.FormatInt(z.MitgliedID, 10),
			z.Zahlungspflichtiger,
			z.IBAN,
			// BIC: das Mitglied trägt keines — die Anschrift auf der Rechnung
			// braucht es, die Lastschrift bei einer deutschen IBAN nicht mehr.
			"",
			BeitragAlsEuro(z.BetragCents),
			z.Verwendungszweck,
			z.UnterschriebenAm.Format(deutschesDatum),
		}); err != nil {
			return nil, fmt.Errorf("moneymoney-csv zeile schreiben: %w", err)
		}
	}

	schreiber.Flush()
	if err := schreiber.Error(); err != nil {
		return nil, fmt.Errorf("moneymoney-csv abschließen: %w", err)
	}

	return puffer.Bytes(), nil
}
