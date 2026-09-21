package service

import (
	"net/url"
	"sort"
	"strings"
)

// Serienmail ist eine vorbereitete Mail an mehrere Empfänger: die mailto-
// Adresse, mit der die Oberfläche das Standard-Mailprogramm öffnet, und die
// Zahl der Empfänger, gegen die sie sich richtet. Ohne Empfänger ist Link leer
// — dann gibt es nichts zu öffnen (siehe SerienmailVorbereiten).
type Serienmail struct {
	Link       string
	Empfaenger int
}

// SerienmailVorbereiten baut die mailto-Adresse für eine Serienmail an mehrere
// Empfänger. Die Adressen stehen im Bcc statt im To: die Empfänger einer
// Serienmail sollen einander nicht sehen — anders als bei einer einzelnen
// Mail an ein Mitglied gibt es hier keinen Empfängerkreis, der sich kennt.
//
// Leere und doppelte Adressen fallen heraus, ohne Meldung: eine fehlende
// E-Mail-Adresse ist am Mitglied selbst sichtbar, nicht erst hier. Bleibt
// keine Adresse übrig, ist der Rückgabewert der Nullwert.
func SerienmailVorbereiten(empfaenger []string) Serienmail {
	bcc := empfaengerBereinigen(empfaenger)
	if len(bcc) == 0 {
		return Serienmail{}
	}

	return Serienmail{
		Link:       "mailto:?bcc=" + url.QueryEscape(strings.Join(bcc, ",")),
		Empfaenger: len(bcc),
	}
}

// empfaengerBereinigen entfernt Leerraum, leere und doppelte Adressen und
// bringt den Rest in eine feste Reihenfolge — die Serienmail soll bei
// gleicher Auswahl immer dieselbe Adresse ergeben.
func empfaengerBereinigen(roh []string) []string {
	gesehen := make(map[string]bool, len(roh))
	bereinigt := make([]string, 0, len(roh))
	for _, e := range roh {
		e = strings.TrimSpace(e)
		if e == "" || gesehen[e] {
			continue
		}
		gesehen[e] = true
		bereinigt = append(bereinigt, e)
	}
	sort.Strings(bereinigt)

	return bereinigt
}
