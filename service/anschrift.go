package service

import "strings"

// Die Anschrift steht in drei getrennten Angaben und nicht in einem Textblock
// (CONTEXT.md → Anschrift). Der Typ und die Schreibweisen, in denen die
// Oberfläche ihn zeigt, stehen hier beieinander, damit Liste, Formular und
// spätere Ansichten die Anschrift gleich schreiben.

// Anschrift ist die Postanschrift eines Mitglieds. Alle drei Felder sind
// freiwillig — der Nullwert ist die leere Anschrift, und ein Mitglied ohne
// Anschrift ist ein gültiges Mitglied.
type Anschrift struct {
	// Adresse ist die Straße samt Hausnummer ("Kanalstraße 12") und nicht die
	// ganze Anschrift.
	Adresse string

	Postleitzahl string
	Ort          string
}

// Leer sagt, ob gar keine Angabe vorliegt. Die Oberfläche zeigt dann einen
// Platzhalter statt einer leeren Zelle.
func (a Anschrift) Leer() bool {
	return a.Adresse == "" && a.Postleitzahl == "" && a.Ort == ""
}

// OrtZeile ist Postleitzahl und Ort als eine Zeile ("12043 Berlin"). Fehlt eine
// der beiden Angaben, bleibt die andere allein stehen — ein führendes oder
// hängendes Leerzeichen entsteht dabei nicht.
func (a Anschrift) OrtZeile() string {
	return strings.TrimSpace(a.Postleitzahl + " " + a.Ort)
}

// Einzeilig ist die ganze Anschrift in einer Zeile ("Kanalstraße 12, 12043
// Berlin") — für den Tooltip der Listenzeile, deren Zelle die beiden Zeilen
// abschneidet. Fehlende Teile lassen kein Komma und kein doppeltes Leerzeichen
// zurück.
func (a Anschrift) Einzeilig() string {
	teile := make([]string, 0, 2)
	for _, teil := range []string{a.Adresse, a.OrtZeile()} {
		if teil != "" {
			teile = append(teile, teil)
		}
	}

	return strings.Join(teile, ", ")
}
