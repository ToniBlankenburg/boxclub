package service

import "fmt"

// Teilnehmer ist ein Eintrag der Teilnehmerliste: nur der Name und die ID, mit
// der die Oberfläche ins Mitglied führt. Mehr steht bewusst nicht darin — die
// Liste ist eine Auskunft über die Anmeldung, keine Verwaltungsansicht
// (CONTEXT.md → Teilnehmerliste).
type Teilnehmer struct {
	MitgliedID int64
	Vorname    string
	Nachname   string
}

// Teilnehmerliste sind die Mitglieder, die für einen Trainingstermin angemeldet
// sind (CONTEXT.md → Teilnehmerliste). Sie wird abgelesen und nirgends geführt.
type Teilnehmerliste struct {
	Termin Trainingstermin

	// Teilnehmer stehen alphabetisch nach Nachname, dann Vorname — die Ordnung
	// der Mitgliederliste.
	Teilnehmer []Teilnehmer
}

// Anzahl ist die Zahl der Teilnehmer. Eine Obergrenze gibt es dazu nicht
// (ADR-0008): die Zahl ist eine Auskunft, kein Fassungsvermögen.
func (l Teilnehmerliste) Anzahl() int {
	return len(l.Teilnehmer)
}

// Teilnehmerlisten liefert zu jedem Trainingstermin seine Teilnehmerliste, in
// der Wochenreihenfolge des Stundenplans. Jeder Termin steht genau einmal da,
// auch der, für den niemand angemeldet ist. Archivierte Termine stehen nur mit
// auchArchivierte (wie bei ListTrainingstermine).
//
// Die Teilnehmer sind die Zeilen der Mitgliederliste, einmal nach Termin
// gruppiert: dieselbe Auskunft über Status und Anmeldung, keine zweite Regel
// daneben. Ausgetretene stehen dort nicht, und ein früherer Zeitraum zählt
// nicht, weil die Zeile die maßgebliche Mitgliedschaft liest.
func (s *MemberService) Teilnehmerlisten(auchArchivierte bool) ([]Teilnehmerliste, error) {
	termine, err := s.ListTrainingstermine(auchArchivierte)
	if err != nil {
		return nil, err
	}

	eintraege, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("teilnehmerlisten: mitglieder lesen: %w", err)
	}

	// Die Zeilen kommen nach Namen sortiert; gruppiert bleibt diese Reihenfolge.
	nachTermin := make(map[int64][]Teilnehmer, len(termine))
	for _, e := range eintraege {
		// Wer ruht, trainiert gerade nicht: die Anmeldung steht, die Liste
		// führt ihn aber nicht (CONTEXT.md → Teilnehmerliste).
		if e.Ruhend {
			continue
		}

		for _, termin := range e.Trainingstermine {
			nachTermin[termin.ID] = append(nachTermin[termin.ID], Teilnehmer{
				MitgliedID: e.MitgliedID,
				Vorname:    e.Vorname,
				Nachname:   e.Nachname,
			})
		}
	}

	listen := make([]Teilnehmerliste, 0, len(termine))
	for _, termin := range termine {
		listen = append(listen, Teilnehmerliste{Termin: termin, Teilnehmer: nachTermin[termin.ID]})
	}

	return listen, nil
}
