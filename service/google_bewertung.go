package service

import "database/sql"

// GoogleBewertung sagt, ob ein Mitglied den Verein bei Google bewertet hat
// (CONTEXT.md → Google-Bewertung). Zweiwertig: es gibt keinen Zustand
// "vielleicht" und keine Sterne — der Verein fragt nach, wer noch nicht
// bewertet hat, und dafür genügt ja oder nein.
//
// Nicht zu verwechseln mit einer Bewertung *des* Mitglieds; der Verein bewertet
// seine Mitglieder nicht.
type GoogleBewertung bool

// Bezeichnung ist der Text, den die Oberfläche zeigt. Er steht wie bei
// Rueckstand.Bezeichnung im Service, damit Liste und Formular dieselben Worte
// benutzen.
func (g GoogleBewertung) Bezeichnung() string {
	if g {
		return "hat bewertet"
	}

	return "hat nicht bewertet"
}

// Scan liest die Spalte als Wahrheitswert. database/sql füllt einen benannten
// bool-Typ nicht von sich aus — ohne diese Methode müsste jede Abfrage in eine
// lokale Variable lesen und danach umwandeln.
func (g *GoogleBewertung) Scan(wert any) error {
	var b sql.NullBool
	if err := b.Scan(wert); err != nil {
		return err
	}

	*g = GoogleBewertung(b.Bool)

	return nil
}
