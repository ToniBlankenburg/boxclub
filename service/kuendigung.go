package service

import "time"

// Kuendigung ist, was beim Ende einer Mitgliedschaft erfasst wird: der Tag, an
// dem gekündigt wurde, und der Tag, zu dem der Austritt wirksam wird. Dazwischen
// liegt die Kündigungsfrist, in der das Mitglied weiter trainiert und weiter
// zahlt (CONTEXT.md → Kündigungsdatum, Kündigungsfrist).
//
// Wie NeuesMitglied und MitgliedPatch ist das ein Eingabewert: an der
// Mitgliedschaft stehen die beiden Daten einzeln, weil der Austritt dort den
// Zeitraum begrenzt und nicht nur einen Vorgang festhält. Zusammen reisen sie,
// weil sie zusammen erfasst werden — im selben Formular, in einem Zug.
//
// Beide Angaben sind Zeiger, weil beide fehlen dürfen: eine Kündigung kann
// vorliegen, ohne dass der Termin feststeht, und ein Austritt aus den
// Altbeständen kann erfasst sein, ohne dass jemand den Tag der Erklärung notiert
// hat. Nur beide zugleich zu leeren ist keine Aussage (siehe pruefen).
type Kuendigung struct {
	// Datum ist der Tag, an dem das Mitglied gekündigt hat. nil heißt "nicht
	// erfasst" — nicht "nicht gekündigt": ein gesetzter Austritt ohne dieses
	// Datum ist der Normalfall der Altbestände.
	Datum *time.Time

	// Austritt ist der Tag, zu dem die Mitgliedschaft endet. Er wird von Hand
	// eingetragen und ausdrücklich **nicht** aus dem Kündigungsdatum berechnet:
	// Fristen haben Sonderfälle (Kulanz, Aufhebungsvertrag, Quartalsende), und
	// ein errechnetes Datum, das man überschreiben muss, ist lästiger als ein
	// leeres Feld. Er darf in der Zukunft liegen — das ist der Normalfall bei
	// laufender Kündigungsfrist.
	Austritt *time.Time
}

// fehlendesDatum ist die Meldung zur einzigen Eingabe, die gar nichts festhält.
// Sie abzuweisen schützt eine bereits erfasste Kündigung vor einem versehentlich
// leer abgeschickten Formular; eine Kündigung zurückzunehmen ist ein eigener
// Vorgang und in v1 nicht vorgesehen.
const fehlendesDatum = "Bitte ein Kündigungsdatum oder ein Austrittsdatum angeben."

// pruefen sammelt die Regelverstöße der Eingabe gegen den Eintritt des
// Zeitraums, den sie beendet — als Texte und nicht als Fehler, damit der
// Aufrufer sie zu den übrigen Meldungen legen kann.
//
// Verglichen wird in ISO-Textform: so vergleicht der ganze Service Kalendertage
// (siehe laufendeMitgliedschaftLesen), und ein Zeitpunktvergleich läge je nach
// Zonenversatz um einen Tag daneben.
//
// Gegen den Eintritt geprüft wird nur der Austritt: ein Kündigungsdatum vor dem
// Eintritt ist kein Tippfehler, sondern der Fall, in dem jemand zurücktritt,
// bevor seine Mitgliedschaft überhaupt beginnt.
func (k Kuendigung) pruefen(eintritt string) []string {
	if k.Datum == nil && k.Austritt == nil {
		return []string{fehlendesDatum}
	}

	var fehler []string

	if k.Austritt != nil {
		austritt := k.Austritt.Format(isoDatum)

		if austritt < eintritt {
			fehler = append(fehler, "Das Austrittsdatum darf nicht vor dem Eintrittsdatum liegen.")
		}

		// Der Austrittstag selbst ist erlaubt: eine Kündigungsfrist von null
		// Tagen gibt es (Aufhebungsvertrag), eine negative nicht.
		if k.Datum != nil && k.Datum.Format(isoDatum) > austritt {
			fehler = append(fehler, "Das Kündigungsdatum darf nicht nach dem Austrittsdatum liegen.")
		}
	}

	return fehler
}
