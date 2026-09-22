// Package i18n übersetzt die sichtbaren Texte der Oberfläche zwischen
// Deutsch und Englisch. Das Domänenmodell selbst — Go-Bezeichner, DB-Spalten,
// das Vokabular aus CONTEXT.md — bleibt davon unberührt; siehe ADR-0017.
package i18n

import "fmt"

// Sprache ist eine der beiden unterstützten Anzeigesprachen.
type Sprache string

const (
	Deutsch  Sprache = "de"
	Englisch Sprache = "en"
)

// Gueltig meldet, ob s eine unterstützte Sprache ist.
func (s Sprache) Gueltig() bool {
	return s == Deutsch || s == Englisch
}

var kataloge = map[Sprache]map[string]string{
	Deutsch:  deutsch,
	Englisch: englisch,
}

// Text liefert den Text zu schluessel in sprache. Fehlt der Schlüssel dort
// oder ist sprache keine bekannte Sprache, fällt Text auf den deutschen
// Katalog zurück, und fehlt er auch dort, auf den Schlüssel selbst — eine
// fehlende Übersetzung darf nie zu einer leeren Stelle in der Oberfläche
// führen (ADR-0017).
//
// args wird wie bei fmt.Sprintf eingesetzt, wenn Argumente übergeben werden;
// ohne Argumente kommt der Text unverändert zurück, damit ein zufälliges
// "%" im Text (etwa in einer IBAN oder einem Dateinamen) nicht als
// Formatverb missverstanden wird.
func Text(sprache Sprache, schluessel string, args ...any) string {
	wert, ok := kataloge[sprache][schluessel]
	if !ok {
		wert, ok = deutsch[schluessel]
	}
	if !ok {
		wert = schluessel
	}

	if len(args) == 0 {
		return wert
	}
	return fmt.Sprintf(wert, args...)
}
