# 06: Hinweis bei Volljährigkeit

**What to build:** Wird ein Mitglied mit abweichendem Kontoinhaber 18, ändert sich nichts an den Abbuchungen: der Zahler bleibt wirksam, das Feld „Abweichender Kontoinhaber“ ist gesperrt, aber gefüllt. Das Formular zeigt einen Hinweis, dass das Mitglied jetzt volljährig ist und der Kontoinhaber geprüft werden soll. Erst wenn der Admin den Kontoinhaber leert, zahlt das Mitglied selbst. Reine Anzeige im Formular — keine Aufgabenliste, keine Benachrichtigung.

**Blocked by:** 02 (Kind-Status), 03 (Abweichender Kontoinhaber)

**Status:** done

- [x] Der Export nach dem 18. Geburtstag nennt weiter den Kontoinhaber
- [x] Der Service meldet für Mitglieder mit abweichendem Zahler und Alter ≥ 18 den Hinweis (Stichtagstest um den Geburtstag)
- [x] Bei Kindern und bei Erwachsenen ohne abweichenden Kontoinhaber erscheint kein Hinweis
- [x] Hinweistext in Deutsch und Englisch
- [x] Tests über die Service-API; `go test ./...` grün
