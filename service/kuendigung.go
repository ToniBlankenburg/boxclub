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

	// Austritt ist der Tag, zu dem die Mitgliedschaft endet. Gespeichert wird
	// ausschließlich, was der Aufrufer mitbringt: aus dem Kündigungsdatum
	// abgeleitet wird hier nichts, weil Fristen Sonderfälle haben (Kulanz,
	// Aufhebungsvertrag) und eine Kündigung auch ganz ohne Termin erfasst werden
	// können muss. Vorgeschlagen wird der reguläre Termin dagegen sehr wohl —
	// siehe RegulaererAustritt, das allein das Formular vorbelegt.
	//
	// Er darf in der Zukunft liegen — das ist der Normalfall bei laufender
	// Kündigungsfrist.
	Austritt *time.Time
}

// KuendigungsfristMonate ist die reguläre Frist der Vereinssatzung: drei Monate
// zum Monatsende.
const KuendigungsfristMonate = 3

// RegulaererAustritt liefert den Tag, zu dem eine an diesem Tag erklärte
// Kündigung nach der Satzung regulär wirksam wird: KuendigungsfristMonate
// weiter, aufgerundet auf das Monatsende. Eine am 12.09. erklärte Kündigung
// wirkt damit zum 31.12.
//
// Das ist ein **Vorschlag und keine Regel**: gespeichert wird, was jemand
// abschickt (siehe Kuendigung.Austritt), und SetKuendigung prüft den Austritt
// nur gegen den Eintritt, nicht gegen diese Frist. Eine kürzere Frist gibt es
// (Aufhebungsvertrag, Kulanz), eine abweichende auch — der vorbelegte Wert im
// Formular lässt sich überschreiben und löschen.
//
// Gerechnet wird über das Monatsende und nicht taggenau, weil taggenau die
// Frage aufwürfe, was der 30. November plus drei Monate ist: einen 30. Februar
// gibt es nicht, und time.AddDate schöbe ihn stillschweigend in den März.
func RegulaererAustritt(kuendigungsdatum time.Time) time.Time {
	// Der letzte Tag eines Monats ist der "nullte" des folgenden; für das Ende
	// des Monats KuendigungsfristMonate weiter also einer mehr. time.Date
	// normalisiert den Überlauf ins nächste Jahr von sich aus.
	return time.Date(
		kuendigungsdatum.Year(),
		kuendigungsdatum.Month()+KuendigungsfristMonate+1, 0,
		0, 0, 0, 0, kuendigungsdatum.Location())
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
