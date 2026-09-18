package service

import "fmt"

// Monatsuebersicht ist die Berechnung hinter dem Dashboard (CONTEXT.md →
// Monatssoll, ADR-0009): das Monatssoll des laufenden Monats, Mitgliederzahlen
// nach Zustand und die Zahl der Mitglieder im Rückstand. Sie speichert nichts
// Neues — jeder Aufruf liest Beitrag, Status und Ruhend neu aus derselben
// Quelle wie die Mitgliederliste.
//
// Es ist ausdrücklich ein Soll und kein Ist: was tatsächlich eingezogen wurde,
// weiß die Bank, nicht die App. Deshalb gibt es hier auch keinen Verlauf über
// mehrere Monate — die Datenbank führt je Mitgliedschaft nur den heutigen
// Beitrag, und eine Kurve daraus wäre eine Hochrechnung im Gewand einer
// Messung.
type Monatsuebersicht struct {
	// MonatssollCents ist die Summe der Beiträge aller Mitgliedschaften, die
	// diesen Monat einziehen: aktiv und in Kündigungsfrist, aber weder ruhend
	// noch noch nicht begonnen noch schon ausgetreten.
	MonatssollCents int64

	// RuhendCents steht dem Monatssoll getrennt daneben (Ticket 26): die
	// Anzahl dazu steht schon in Mitgliederzahlen.Ruhend und wird hier nicht
	// verdoppelt. Zusammen zählen sie genau die Mitgliedschaften, die ohne
	// ihre Ruhendstellung ins Monatssoll eingegangen wären — ohne diesen
	// Ausweis stünde die Summe unerklärt kleiner da, als die Mitgliederzahl
	// vermuten lässt.
	RuhendCents int64

	// Mitgliederzahlen zählt jedes Mitglied genau einmal nach seinem
	// Zustand.
	Mitgliederzahlen Mitgliederzahlen

	// RueckstandAnzahl ist die Zahl der Mitglieder im Rückstand — über den
	// gesamten Bestand und nicht nur über die aktiven: ein Austritt erlässt
	// keine Schulden (ADR-0006), und wer im Rückstand ausgetreten ist, bleibt
	// es auch als Ehemaliger.
	RueckstandAnzahl int

	// NeuEingetretenDiesenMonat und AusgetretenDiesenMonat zählen Bewegungen
	// und keine Zustände: eine Mitgliedschaft, deren Eintritt oder Austritt
	// schon erreicht und in den laufenden Monat gefallen ist, zählt hier
	// unabhängig davon, was aus ihr inzwischen geworden ist — auch ein noch
	// im selben Monat wieder eingetretenes Mitglied zählt bei seinem Austritt
	// mit. Ein Termin später in diesem Monat, der noch bevorsteht, zählt
	// dagegen nicht mit: "eingetreten" und "ausgetreten" sind hier vollendete
	// Vorgänge, nicht bloß ein Datum auf diesem Kalenderblatt.
	NeuEingetretenDiesenMonat int
	AusgetretenDiesenMonat    int
}

// Mitgliederzahlen zählt die Mitglieder nach ihrem Zustand, je einmal: die
// fünf Zahlen ergeben zusammen Gesamt.
//
// Ruhend steht hier — anders als überall sonst (CONTEXT.md → Ruhend) — nicht
// neben Aktiv/In Kündigungsfrist, sondern an ihrer Stelle: ein ruhendes
// Mitglied zählte sonst doppelt, einmal im Lebenszyklus und einmal im Ausweis
// daneben. Das deckt sich mit dem Monatssoll, das dieselben Mitgliedschaften
// genauso herausrechnet.
type Mitgliederzahlen struct {
	Neu                int
	Aktiv              int
	InKuendigungsfrist int
	Ruhend             int
	Ausgetreten        int
}

// Gesamt ist die Zahl aller Mitglieder — die Summe der fünf Zustände, die sich
// gegenseitig ausschließen und zusammen den ganzen Bestand ergeben.
func (z Mitgliederzahlen) Gesamt() int {
	return z.Neu + z.Aktiv + z.InKuendigungsfrist + z.Ruhend + z.Ausgetreten
}

// Monatsuebersicht liefert die Zahlen des Dashboards. Gelesen wird dieselbe
// maßgebliche Mitgliedschaft je Mitglied, die auch die Mitgliederliste zeigt
// (einschließlich der Ehemaligen, denn ohne sie fehlte die Ausgetreten-Zahl
// und der Rückstand Ehemaliger).
func (s *MemberService) Monatsuebersicht() (Monatsuebersicht, error) {
	zeilen, err := s.eintraegeLesen(true, "")
	if err != nil {
		return Monatsuebersicht{}, err
	}

	heute := heute()

	var u Monatsuebersicht
	for _, z := range zeilen {
		e := z.eintrag

		if e.Rueckstand.Offen {
			u.RueckstandAnzahl++
		}

		status := statusAus(e.Eintritt, e.Kuendigungsdatum, e.Austritt, heute)

		// Die Reihenfolge entscheidet über den einen Fall, in dem sich zwei
		// Merkmale überschneiden könnten: Ausgetreten und Neu gehen beide vor
		// Ruhend, aus demselben Grund, aus dem statusAus vom Ende des
		// Lebenszyklus her prüft — ein erreichtes Ende ist die aussagekräftigere
		// Auskunft, und ein noch nicht begonnener Zeitraum zieht ohnehin nichts
		// ein, ruhend gestellt oder nicht (siehe
		// TestMonatsuebersicht_RuhendVorDemEintrittZaehltAlsNeuUndNichtAlsRuhend).
		switch {
		case status.Ausgetreten():
			u.Mitgliederzahlen.Ausgetreten++
		case status.Neu():
			u.Mitgliederzahlen.Neu++
		case e.Ruhend:
			u.Mitgliederzahlen.Ruhend++
			u.RuhendCents += e.BeitragCents
		case status.InKuendigungsfrist():
			u.Mitgliederzahlen.InKuendigungsfrist++
		default:
			u.Mitgliederzahlen.Aktiv++
		}

		if zaehltInsMonatssoll(status, e.Ruhend) {
			u.MonatssollCents += e.BeitragCents
		}
	}

	if u.NeuEingetretenDiesenMonat, u.AusgetretenDiesenMonat, err = s.monatsbewegungen(heute); err != nil {
		return Monatsuebersicht{}, err
	}

	return u, nil
}

// zaehltInsMonatssoll sagt, ob eine Mitgliedschaft mit diesem Status und
// Ruhend-Merkmal diesen Monat Beitrag einzieht — dieselbe Regel für
// Monatsuebersicht hier und für den MoneyMoney-Export (service/moneymoney.go,
// ADR-0013), an einer Stelle, damit beide nicht auseinanderlaufen können.
// Aktiv und in Kündigungsfrist ziehen ein, Ausgetreten, Neu und Ruhend nicht.
func zaehltInsMonatssoll(status Status, ruhend bool) bool {
	return !status.Ausgetreten() && !status.Neu() && !ruhend
}

// monatsbewegungen zählt, wie viele Mitgliedschaften im Monat von heute bereits
// begonnen bzw. geendet haben. Gezählt wird über alle Mitgliedschaften und
// nicht nur über die eine maßgebliche je Mitglied: ein Austritt zählt auch
// dann, wenn im selben Monat schon wieder eingetreten wurde, und dann zeigt
// die maßgebliche Mitgliedschaft längst den neuen Zeitraum.
//
// Die Obergrenze ist heute selbst und nicht das Monatsende: ein Eintritt oder
// Austritt später in diesem Monat steht zwar schon fest, ist aber noch nicht
// eingetreten (das zeigen Mitgliederzahlen.Neu bzw. der Lebenszyklus-Status
// weiterhin richtig) — "eingetreten" und "ausgetreten" wären sonst eine
// Ankündigung, die das Dashboard als bereits geschehen ausgäbe.
func (s *MemberService) monatsbewegungen(heute string) (neu, ausgetreten int, err error) {
	anfang := monatsanfang(heute)

	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM mitgliedschaft WHERE eintritt >= ? AND eintritt <= ?`,
		anfang, heute).Scan(&neu); err != nil {
		return 0, 0, fmt.Errorf("neueintritte diesen monat lesen: %w", err)
	}

	// NULL-Austritte scheiden aus dem Vergleich von selbst aus: SQL vergleicht
	// NULL mit nichts als wahr, auch nicht mit sich selbst.
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM mitgliedschaft WHERE austritt >= ? AND austritt <= ?`,
		anfang, heute).Scan(&ausgetreten); err != nil {
		return 0, 0, fmt.Errorf("austritte diesen monat lesen: %w", err)
	}

	return neu, ausgetreten, nil
}

// monatsanfang liefert den ersten Tag des Monats zu einem ISO-Kalendertag
// ("JJJJ-MM-TT") als Text — reine Zeichenkettenarbeit wie bei laeuftNoch statt
// eines Umwegs über time.Parse: die ersten acht Zeichen sind bereits
// "JJJJ-MM-", der Tag wird auf 01 gesetzt.
func monatsanfang(heute string) string {
	return heute[:8] + "01"
}
