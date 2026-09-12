package service

import "time"

// Status ist der Lebenszyklus-Zustand einer Mitgliedschaft. Er wird **nicht
// gepflegt, sondern abgelesen** (CONTEXT.md → Status): gespeichert sind allein
// die drei Datumsfelder Eintritt, Kündigungsdatum und Austritt, der Zustand
// ergibt sich bei jeder Abfrage neu aus ihnen und dem heutigen Tag.
//
// Das ist der Gewinn gegenüber der Excel, in der der Status eine Spalte war,
// die jemand von Hand nachziehen musste: hier kann er mit den Daten gar nicht
// erst in Widerspruch geraten, und die aktive Liste räumt sich am Stichtag von
// selbst auf.
//
// Vier Zustände, mehr gibt es nicht. Die Excel kannte daneben `Mitglied` (was
// dasselbe ist wie `Aktiv`) sowie `Stillgelegt` und `Inakiv` — beides keine
// Zustände, sondern das Merkmal *ruhend*, das neben dem Lebenszyklus steht und
// als Einziges in dieser Reihe gespeichert wird (CONTEXT.md → Ruhend).
type Status int

const (
	// StatusNeu: der Eintritt liegt in der Zukunft. Das Mitglied ist erfasst,
	// angefangen hat es noch nicht.
	StatusNeu Status = iota

	// StatusAktiv: der Eintritt ist erreicht, der Austritt nicht.
	StatusAktiv

	// StatusInKuendigungsfrist: eine Kündigung ist erfasst, der Zeitraum läuft
	// aber noch. Das Mitglied trainiert weiter und zahlt weiter (CONTEXT.md →
	// Kündigungsfrist) — es ist ein aktives Mitglied mit Enddatum.
	StatusInKuendigungsfrist

	// StatusAusgetreten: der Austritt ist erreicht. Der einzige Zustand, der ein
	// Mitglied aus der Standardansicht nimmt.
	StatusAusgetreten
)

// Bezeichnung ist der Text, den die Oberfläche zum Zustand zeigt. Er steht wie
// bei Rueckstand.Bezeichnung im Service, damit Liste und spätere Ansichten
// dieselben Worte benutzen; die Farben dazu stehen umgekehrt im Template.
func (s Status) Bezeichnung() string {
	switch s {
	case StatusNeu:
		return "Neu"
	case StatusAktiv:
		return "Aktiv"
	case StatusInKuendigungsfrist:
		return "In Kündigungsfrist"
	case StatusAusgetreten:
		return "Ausgetreten"
	}

	// Alle vier Konstanten stehen ausdrücklich oben, damit kein unbekannter Wert
	// still als "Aktiv" durchgeht — ausgerechnet die Auskunft, die man bei einem
	// Ausgetretenen am wenigsten gebrauchen kann. Hierher kommt nur, wer sich
	// einen Status selbst zusammenrechnet; statusAus liefert nichts anderes.
	return "unbekannt"
}

// Die Fragen an den Zustand. Sie stehen hier, weil das Template die Konstanten
// nicht kennt: html/template kann keinen Go-Bezeichner vergleichen, wohl aber
// eine Methode aufrufen. Beantwortet werden sie deshalb dort, wo die Zustände
// definiert sind, und nicht über Textvergleiche im Markup.
//
// Ein Gegenstück zu StatusAktiv fehlt mit Absicht: die Kennzeichen-Kaskade prüft
// vom Ende des Lebenszyklus her und lässt "aktiv" als letzten Zweig übrig, und
// über die Standardansicht entscheidet ohnehin Ausgetreten. Eine Methode, die
// niemand ruft, kommt in einer Zeile zurück, sobald jemand sie braucht.

// Neu sagt, ob der Eintritt noch bevorsteht.
func (s Status) Neu() bool { return s == StatusNeu }

// InKuendigungsfrist sagt, ob eine Kündigung erfasst, der Zeitraum aber noch
// nicht beendet ist.
func (s Status) InKuendigungsfrist() bool { return s == StatusInKuendigungsfrist }

// Ausgetreten sagt, ob der Austritt erreicht ist. Das ist zugleich die Frage,
// die über die Standardansicht entscheidet: alles andere gilt als aktiv.
func (s Status) Ausgetreten() bool { return s == StatusAusgetreten }

// heute liefert den heutigen Kalendertag in ISO-Form — die Form, in der der
// ganze Service Kalendertage vergleicht (siehe laufendeMitgliedschaftLesen).
//
// Die Uhr steht bewusst nicht als Feld am Service: ein Vereinsadmin liest den
// Status gegen den Tag, an dem er hinschaut, und sonst gegen nichts. Prüfbar
// bleiben die Ränder trotzdem, weil statusAus den Tag als Parameter nimmt und
// die Tests ihre Fixtures relativ zu heute legen.
func heute() string {
	return time.Now().Format(isoDatum)
}

// statusAus liest den Zustand aus den drei Datumsfeldern eines Zeitraums ab.
// Die Ableitung steht hier ein einziges Mal: Mitgliedschaft und Listeneintrag
// zeigen denselben Status, und zwei Ableitungen könnten auseinanderlaufen —
// genau das, was ein abgeleiteter Wert verhindern soll.
//
// Verglichen wird in ISO-Textform gegen heute: so vergleicht der Service alle
// Kalendertage, und ein Zeitpunktvergleich läge je nach Zonenversatz um einen
// Tag daneben — aus der Datenbank gelesene Daten liegen in UTC, "heute" kommt
// aus der lokalen Uhr.
//
// Beide Ränder sind einschließend: am Eintrittstag ist der Eintritt erreicht,
// am Austrittstag der Austritt. Ein Zeitraum, der heute endet, ist damit heute
// schon vorbei.
//
// Geprüft wird vom Ende des Lebenszyklus her, weil der spätere Zustand immer
// der aussagekräftigere ist. Das entscheidet den einzigen Fall, in dem sich
// zwei Regeln überschneiden: wer gekündigt hat, bevor sein Eintritt überhaupt
// erreicht war (Kuendigung.pruefen lässt das ausdrücklich zu), steht *In
// Kündigungsfrist* und nicht *Neu* — dass er geht, ist die Auskunft, auf die es
// ankommt.
//
// In der Kündigungsfrist steht, bei wem eine Kündigung erfasst ist, deren
// Wirkung noch aussteht — und erfasst ist sie, sobald eines der beiden Daten
// steht. Beide dürfen einzeln fehlen und sagen einzeln schon genug: ein
// Kündigungsdatum ohne Termin heißt "erklärt, wann sie wirkt ist offen", ein
// Austritt ohne Kündigungsdatum "wir gehen am soundsovielten auseinander, den
// Tag der Erklärung hat niemand notiert" (der Normalfall der Altbestände).
// Nur beide zugleich leer ist keine Aussage — Kuendigung.pruefen weist das
// deshalb ab, und genau daran hängt diese Regel.
//
// *Aktiv* bleibt damit der Zustand, über dessen Ende nichts erfasst ist.
func statusAus(eintritt time.Time, kuendigungsdatum, austritt *time.Time, heute string) Status {
	if austritt != nil && austritt.Format(isoDatum) <= heute {
		return StatusAusgetreten
	}

	if kuendigungsdatum != nil || austritt != nil {
		return StatusInKuendigungsfrist
	}

	if eintritt.Format(isoDatum) > heute {
		return StatusNeu
	}

	return StatusAktiv
}
