package service

import (
	"fmt"
	"strings"
)

// Der Trainingsslot ist ein wöchentlicher Termin als freier Text, die
// Trainingsfrequenz ist seine Anzahl (CONTEXT.md → Trainingsslot,
// Trainingsfrequenz). Beide stehen hier beieinander, weil die Frequenz nichts
// ist, was man außerhalb der Slots noch einmal festhalten könnte — genau das
// soll sie nicht sein.

// MaxTrainingsslots ist die Obergrenze: mehr als drei wöchentliche Termine
// bietet der Verein nicht an. Ein vierter Slot wäre eine Frequenz, die es nicht
// gibt, und wird deshalb abgewiesen statt abgeschnitten.
const MaxTrainingsslots = 3

// zuVieleSlots ist die Meldung zum einzigen Regelverstoß, den Slots kennen.
var zuVieleSlots = fmt.Sprintf("Es sind höchstens %d Trainingsslots möglich.", MaxTrainingsslots)

// Trainingsfrequenz ist, wie oft pro Woche ein Mitglied trainiert. Sie wird
// nirgends gespeichert, sondern ist die Anzahl der Trainingsslots seiner
// Mitgliedschaft — dadurch können Frequenz und Slots gar nicht erst
// auseinanderlaufen.
//
// Der Nullwert heißt "keine Frequenz" und nicht "1×": ein Mitglied ohne
// eingetragenen Termin trainiert nicht einmal die Woche, sondern es ist
// schlicht nichts vereinbart.
type Trainingsfrequenz int

// Bezeichnung ist der Text, den die Oberfläche zur Frequenz zeigt. Er steht wie
// bei Rueckstand.Bezeichnung im Service, damit Liste und spätere Ansichten
// dieselben Worte benutzen.
func (f Trainingsfrequenz) Bezeichnung() string {
	if f <= 0 {
		return "keine Frequenz"
	}

	return fmt.Sprintf("%d× pro Woche", int(f))
}

// trainingsfrequenzAus liest die Frequenz aus einer Slot-Liste ab. Die
// Ableitung steht hier ein einziges Mal: Mitgliedschaft und Listeneintrag
// zeigen beide dieselbe Frequenz, und zwei Zählungen könnten wieder
// auseinanderlaufen — genau das, was die Ableitung verhindern soll.
func trainingsfrequenzAus(slots []string) Trainingsfrequenz {
	return Trainingsfrequenz(len(slots))
}

// Frequenzfilter grenzt die Ergebnisliste nach der Trainingsfrequenz ein. Die
// Stufen tragen ihre Frequenz als Wert, deshalb ist der Vergleich unten ein
// Vergleich und keine Zuordnungstabelle.
//
// Eine Stufe für "kein Training" gibt es bewusst nicht: gefragt wird nach den
// Trainierenden einer Frequenz ("alle 2×-Trainierenden auf einmal"), und wer
// gar keinen Slot hat, ist keine solche Gruppe.
type Frequenzfilter int

const (
	// FrequenzfilterAlle ist der Nullwert und grenzt nicht ein.
	FrequenzfilterAlle Frequenzfilter = 0
	// FrequenzfilterEinmal: nur Mitglieder mit genau einem Trainingsslot.
	FrequenzfilterEinmal Frequenzfilter = 1
	// FrequenzfilterZweimal: nur Mitglieder mit genau zwei Trainingsslots.
	FrequenzfilterZweimal Frequenzfilter = 2
	// FrequenzfilterDreimal: nur Mitglieder mit genau drei Trainingsslots.
	FrequenzfilterDreimal Frequenzfilter = 3
)

// trifft entscheidet, ob eine Frequenz durch diesen Filter kommt.
func (f Frequenzfilter) trifft(frequenz Trainingsfrequenz) bool {
	if f == FrequenzfilterAlle {
		return true
	}

	return int(f) == int(frequenz)
}

// trainingsslotsNormalisieren schneidet Leerraum ab und lässt leere Angaben
// weg. Das Formular schickt immer alle Felder mit, auch die unausgefüllten —
// ein leeres Feld ist aber kein Trainingsslot und zählt weder für die Frequenz
// noch gegen die Obergrenze.
//
// Normalisiert wird hier, wo geprüft und gespeichert wird: sonst prüfte die
// Regel etwas anderes, als hinterher in der Datenbank steht.
func trainingsslotsNormalisieren(slots []string) []string {
	gesaeubert := make([]string, 0, len(slots))
	for _, slot := range slots {
		if slot = strings.TrimSpace(slot); slot != "" {
			gesaeubert = append(gesaeubert, slot)
		}
	}

	return gesaeubert
}

// trainingsslotsPruefen meldet den Verstoß gegen die Obergrenze — als Text und
// nicht als Fehler, damit der Aufrufer ihn zu den übrigen Meldungen seiner
// Eingabe legen kann.
func trainingsslotsPruefen(slots []string) []string {
	if len(trainingsslotsNormalisieren(slots)) > MaxTrainingsslots {
		return []string{zuVieleSlots}
	}

	return nil
}
