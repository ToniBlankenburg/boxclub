package service

import "time"

// Volljaehrigkeitsalter ist das Alter, ab dem ein Mitglied kein Kind mehr ist.
const Volljaehrigkeitsalter = 18

// IstKind sagt, ob das Mitglied heute ein Kind ist (CONTEXT.md → Kind).
func (m Mitglied) IstKind() bool {
	return m.IstKindAm(time.Now())
}

// IstKindAm sagt, ob das Mitglied an diesem Tag ein Kind ist. Der Status wird
// bei jedem Lesen abgeleitet und nie gespeichert.
//
// Mit Geburtsdatum entscheidet allein das Datum: Kind ist, wer den 18.
// Geburtstag noch nicht erreicht hat. Wer an einem 29.02. geboren ist, wird in
// einem Jahr ohne Schalttag am 01.03. volljährig (AddDate normalisiert so).
// Ohne Geburtsdatum zählt das von Hand gesetzte Kennzeichen; ein gesetztes
// Kennzeichen neben einem Geburtsdatum wird ignoriert.
func (m Mitglied) IstKindAm(tag time.Time) bool {
	if m.Geburtsdatum == nil {
		return m.KindManuell
	}

	volljaehrig := m.Geburtsdatum.AddDate(Volljaehrigkeitsalter, 0, 0).Format(isoDatum)

	return tag.Format(isoDatum) < volljaehrig
}
