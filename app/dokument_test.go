package app

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Der Export ist die dritte Ausnahme von "app/ wird nicht unit-getestet"
// (CLAUDE.md). Getestet wird keine Fachlogik, sondern der einzige Weg, auf dem
// ein PDF die Datenbank wieder verlässt: Dokumente liegen als Blob darin
// (ADR-0007), und fällt dieser Weg aus, kommt niemand mehr an seinen Vertrag.
//
// Von Hand ist er hier nicht zu prüfen — er öffnet einen Dialog des
// Betriebssystems. Genau dafür ist das Speicherziel ein Funktionstyp: der Test
// setzt ein Ziel ein und sieht nach, was dort ankommt.

// testApp baut Service und Handler für einen Handler-Test und legt ein Mitglied
// samt laufender Mitgliedschaft an.
func testApp(t *testing.T) (*App, *service.MemberService, int64) {
	t.Helper()

	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	mitgliedID, err := svc.Create(service.NeuesMitglied{
		Vorname: "Marek", Nachname: "Sobotka", BeitragCents: 6500,
		Eintritt: time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := svc.Get(mitgliedID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	a, err := New(svc)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return a, svc, m.Mitgliedschaften[0].ID
}

// hochladen schickt eine Datei an die Vertragsablage eines Zeitraums, so wie das
// Formular im Browser es tut.
func hochladen(t *testing.T, a *App, mitgliedschaftID int64, name string, inhalt []byte) *httptest.ResponseRecorder {
	t.Helper()

	var koerper bytes.Buffer
	schreiber := multipart.NewWriter(&koerper)

	feld, err := schreiber.CreateFormFile("datei", name)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := feld.Write(inhalt); err != nil {
		t.Fatalf("Datei schreiben: %v", err)
	}
	if err := schreiber.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	anfrage := httptest.NewRequest("POST",
		"/api/mitgliedschaft/"+strconv.FormatInt(mitgliedschaftID, 10)+"/vertrag", &koerper)
	anfrage.Header.Set("Content-Type", schreiber.FormDataContentType())

	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, anfrage)

	return w
}

func TestVertragHochladenUndExportieren(t *testing.T) {
	a, svc, mitgliedschaftID := testApp(t)

	pdf := []byte("%PDF-1.7\nunterschrieben")

	if w := hochladen(t, a, mitgliedschaftID, "vertrag-sobotka.pdf", pdf); w.Code != http.StatusOK {
		t.Fatalf("Hochladen = %d, erwartet 200: %s", w.Code, w.Body.String())
	}

	abgelegt, err := svc.VertragInhalt(mitgliedschaftID)
	if err != nil {
		t.Fatalf("VertragInhalt: %v", err)
	}
	if !bytes.Equal(abgelegt.Inhalt, pdf) {
		t.Errorf("abgelegter Inhalt = %q, erwartet die hochgeladene Datei", abgelegt.Inhalt)
	}

	// Der Dialog schlägt den gespeicherten Dateinamen vor; der Test nimmt ihn
	// entgegen und nennt ein Ziel im Temp-Verzeichnis.
	var vorgeschlagen string
	ziel := filepath.Join(t.TempDir(), "export.pdf")
	a.SpeicherzielSetzen(func(vorschlag string) (string, error) {
		vorgeschlagen = vorschlag

		return ziel, nil
	})

	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("POST",
		"/api/mitgliedschaft/"+strconv.FormatInt(mitgliedschaftID, 10)+"/vertrag/export", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("Export = %d, erwartet 200: %s", w.Code, w.Body.String())
	}
	if vorgeschlagen != "vertrag-sobotka.pdf" {
		t.Errorf("Vorschlag = %q, erwartet den abgelegten Dateinamen", vorgeschlagen)
	}

	geschrieben, err := os.ReadFile(ziel)
	if err != nil {
		t.Fatalf("exportierte Datei lesen: %v", err)
	}
	if !bytes.Equal(geschrieben, pdf) {
		t.Errorf("exportierte Datei = %q, erwartet das abgelegte PDF", geschrieben)
	}
}

// Ein abgebrochener Dialog ist kein Vorgang: der Block kommt unverändert zurück,
// ohne Erfolgsmeldung und ohne Fehler. Geschrieben werden kann dabei nichts —
// der Handler bekommt gar kein Ziel genannt —, gemeldet aber sehr wohl, und
// genau das soll ausbleiben.
func TestVertragExportieren_AbgebrochenerDialogMeldetNichts(t *testing.T) {
	a, _, mitgliedschaftID := testApp(t)

	if w := hochladen(t, a, mitgliedschaftID, "vertrag.pdf", []byte("%PDF-1.7\nx")); w.Code != http.StatusOK {
		t.Fatalf("Hochladen = %d, erwartet 200: %s", w.Code, w.Body.String())
	}

	gefragt := false
	a.SpeicherzielSetzen(func(string) (string, error) {
		gefragt = true

		return "", nil
	})

	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("POST",
		"/api/mitgliedschaft/"+strconv.FormatInt(mitgliedschaftID, 10)+"/vertrag/export", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("Export = %d, erwartet 200: %s", w.Code, w.Body.String())
	}
	if !gefragt {
		t.Error("der Dialog wurde nicht gefragt — der Test prüfte dann keinen Abbruch")
	}
	if strings.Contains(w.Body.String(), "gespeichert") {
		t.Errorf("Antwort meldet einen Export, obwohl abgebrochen wurde:\n%s", w.Body.String())
	}
	// Der Block steht unverändert da: der Vertrag ist noch dran.
	if !strings.Contains(w.Body.String(), "vertrag.pdf") {
		t.Errorf("Antwort zeigt den abgelegten Vertrag nicht mehr:\n%s", w.Body.String())
	}
}

// Eine abgewiesene Datei kommt als Meldung im Block zurück und nicht als
// Fehlerstatus: htmx tauscht Fehlerantworten nicht ein, und der Klick bliebe für
// den Nutzer wirkungslos.
func TestVertragHochladen_WeistNichtPDFImBlockAb(t *testing.T) {
	a, svc, mitgliedschaftID := testApp(t)

	w := hochladen(t, a, mitgliedschaftID, "vertrag.pdf", []byte("PK\x03\x04 kein PDF"))

	if w.Code != http.StatusOK {
		t.Fatalf("Hochladen = %d, erwartet 200: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "keine PDF-Datei") {
		t.Errorf("Antwort nennt den Grund nicht:\n%s", w.Body.String())
	}

	if _, err := svc.VertragInhalt(mitgliedschaftID); err == nil {
		t.Error("VertragInhalt = nil, erwartet ErrNichtGefunden — abgelegt wurde nichts")
	}
}

// Ohne hinterlegten Dialog ist niemand gefragt worden — das soll dastehen und
// nicht wie ein Abbruch aussehen.
func TestVertragExportieren_OhneDateidialogMeldetEsDas(t *testing.T) {
	a, _, mitgliedschaftID := testApp(t)

	if w := hochladen(t, a, mitgliedschaftID, "vertrag.pdf", []byte("%PDF-1.7\nx")); w.Code != http.StatusOK {
		t.Fatalf("Hochladen = %d, erwartet 200: %s", w.Code, w.Body.String())
	}

	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("POST",
		"/api/mitgliedschaft/"+strconv.FormatInt(mitgliedschaftID, 10)+"/vertrag/export", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("Export = %d, erwartet 200: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Datei-Dialog") {
		t.Errorf("Antwort sagt nicht, dass der Dialog fehlt:\n%s", w.Body.String())
	}
}
