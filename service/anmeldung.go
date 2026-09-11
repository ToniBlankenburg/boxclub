package service

import "time"

// Anmeldung ist, was beim Zustandekommen einer Mitgliedschaft einmalig
// festgehalten wird: der Tag, an dem die Anmeldung abgegeben wurde, und die
// Gebühr, die dafür fällig war (CONTEXT.md → Anmeldedatum).
//
// Beides reist zusammen, weil es zusammen gehört — es beschreibt denselben
// Vorgang, steht im Formular nebeneinander und wird zusammen geändert. Beide
// Angaben sind freiwillig: der Nullwert ist "nichts erfasst", und das ist bei
// den Altbeständen aus der Excel der Normalfall.
type Anmeldung struct {
	// Datum ist der Tag der Anmeldung. Er liegt in der Regel vor dem Eintritt
	// und ist ausdrücklich nicht derselbe Tag: der Eintritt ist der Beginn der
	// Mitgliedschaft, üblicherweise ein Monatserster. nil heißt "nicht erfasst".
	Datum *time.Time

	// GebuehrCents ist die einmalige Gebühr in Cent. Sie ist ein historischer
	// Wert — sie hält fest, was tatsächlich gezahlt wurde, und wird nie neu
	// berechnet. 0 heißt "keine erhoben"; einen Unterschied zwischen "keine" und
	// "null Euro" gibt es nicht, denn geflossen ist in beiden Fällen nichts.
	GebuehrCents int64
}
