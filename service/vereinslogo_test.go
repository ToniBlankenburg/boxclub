package service_test

import (
	"bytes"
	"errors"
	"image"
	_ "image/png"
	"strconv"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// pngKennung und jpegKennung sind Kurznamen für service.PNGKennung und
// service.JPEGKennung — dieselben Bytes, die auch LogoAusUpload prüft, statt
// sie hier ein zweites Mal abzuschreiben (dieselbe Art Fixture wie pdf() in
// dokument_test.go, nur exportiert statt neu erfunden).
var (
	pngKennung  = service.PNGKennung
	jpegKennung = service.JPEGKennung
)

// svg baut ein winziges, aber gültiges SVG mit der angegebenen ViewBox — genug,
// damit oksvg es liest, ohne dass es um seinen Inhalt geht.
func svg(breite, hoehe int) []byte {
	return []byte(strings.ReplaceAll(strings.ReplaceAll(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 W H"><rect width="W" height="H" fill="#ff0000"/></svg>`,
		"W", strconv.Itoa(breite)), "H", strconv.Itoa(hoehe)))
}

func TestLogoAusUpload_NimmtPNGUndJPEGUnveraendertAn(t *testing.T) {
	png := append(bytes.Clone(pngKennung), []byte("nutzdaten")...)

	bild, mime, err := service.LogoAusUpload("logo.png", png)
	if err != nil {
		t.Fatalf("LogoAusUpload(png): %v", err)
	}
	if mime != "image/png" {
		t.Errorf("mime = %q, erwartet image/png", mime)
	}
	if !bytes.Equal(bild, png) {
		t.Error("PNG-Inhalt wurde verändert, erwartet unverändert (ADR-0012)")
	}

	jpeg := append(bytes.Clone(jpegKennung), []byte("nutzdaten")...)

	bild, mime, err = service.LogoAusUpload("logo.jpg", jpeg)
	if err != nil {
		t.Fatalf("LogoAusUpload(jpeg): %v", err)
	}
	if mime != "image/jpeg" {
		t.Errorf("mime = %q, erwartet image/jpeg", mime)
	}
	if !bytes.Equal(bild, jpeg) {
		t.Error("JPEG-Inhalt wurde verändert, erwartet unverändert (ADR-0012)")
	}
}

// Ein SVG wird beim Upload zu PNG gerastert — die längste Kante der ViewBox
// wird zu 1000 px, die andere im selben Verhältnis kleiner (ADR-0012). Ab hier
// existiert kein Vektor mehr.
func TestLogoAusUpload_RastertSVGZuPNGMitErhaltenemSeitenverhaeltnis(t *testing.T) {
	bild, mime, err := service.LogoAusUpload("logo.svg", svg(200, 100))
	if err != nil {
		t.Fatalf("LogoAusUpload(svg): %v", err)
	}
	if mime != "image/png" {
		t.Errorf("mime = %q, erwartet image/png — ein SVG wird zu PNG gerastert", mime)
	}
	if !bytes.HasPrefix(bild, pngKennung) {
		t.Error("Ergebnis beginnt nicht mit der PNG-Kennung")
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(bild))
	if err != nil {
		t.Fatalf("gerastertes Bild dekodieren: %v", err)
	}
	if cfg.Width != 1000 || cfg.Height != 500 {
		t.Errorf("Maße = %dx%d, erwartet 1000x500 (ViewBox 200x100 auf Kante 1000 skaliert)",
			cfg.Width, cfg.Height)
	}
}

func TestLogoAusUpload_WeistZuGrosseDateiAb(t *testing.T) {
	// Ein Byte über der Grenze: gültiges PNG, nur zu groß.
	zuGross := append(bytes.Clone(pngKennung), bytes.Repeat([]byte("x"), service.MaxLogoBytes)...)

	_, _, err := service.LogoAusUpload("logo.png", zuGross)

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("LogoAusUpload = %v, erwartet einen ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) != 1 {
		t.Fatalf("Meldungen = %+v, erwartet genau eine", validierung.Meldungen)
	}
	m := validierung.Meldungen[0]
	if m.Schluessel != "validierung.logo.zu_gross" || len(m.Args) == 0 || m.Args[0] != "logo.png" {
		t.Errorf("Meldung = %+v, erwartet eine, die die Datei beim Namen nennt", m)
	}

	// Genau an der Grenze geht es durch: die Grenze schließt ihren Wert ein.
	gerade := append(bytes.Clone(pngKennung),
		bytes.Repeat([]byte("x"), service.MaxLogoBytes-len(pngKennung))...)
	if _, _, err := service.LogoAusUpload("logo.png", gerade); err != nil {
		t.Fatalf("LogoAusUpload an der Grenze: %v", err)
	}
}

func TestLogoAusUpload_WeistUnbekanntesFormatAb(t *testing.T) {
	// Weder PNG- noch JPEG-Signatur, und kein gültiges SVG.
	_, _, err := service.LogoAusUpload("vertrag.pdf", []byte("%PDF-1.7\nkein Bild"))

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("LogoAusUpload = %v, erwartet einen ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) != 1 {
		t.Fatalf("Meldungen = %+v, erwartet genau eine", validierung.Meldungen)
	}
	m := validierung.Meldungen[0]
	if m.Schluessel != "validierung.logo.kein_format" || len(m.Args) == 0 || m.Args[0] != "vertrag.pdf" {
		t.Errorf("Meldung = %+v, erwartet eine, die die Datei beim Namen nennt", m)
	}
}

func TestLogoAusUpload_MeldetBeideGruendeAufEinmal(t *testing.T) {
	video := bytes.Repeat([]byte("x"), service.MaxLogoBytes+1)

	_, _, err := service.LogoAusUpload("training.mov", video)

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("LogoAusUpload = %v, erwartet einen ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) != 2 {
		t.Errorf("Meldungen = %q, erwartet beide Gründe auf einmal", validierung.Meldungen)
	}
}

func TestLogoAktualisieren_OhneAenderungBleibtDasVorhandeneLogoStehen(t *testing.T) {
	aktuell := append(bytes.Clone(pngKennung), []byte("altes-logo")...)

	logo, mime, err := service.LogoAktualisieren(aktuell, "image/png", service.LogoAenderung{})
	if err != nil {
		t.Fatalf("LogoAktualisieren: %v", err)
	}
	if !bytes.Equal(logo, aktuell) || mime != "image/png" {
		t.Errorf("Logo/Mime = %v/%q, erwartet unverändert %v/image/png", logo, mime, aktuell)
	}
}

func TestLogoAktualisieren_EntfernenLoeschtDasLogo(t *testing.T) {
	aktuell := append(bytes.Clone(pngKennung), []byte("altes-logo")...)

	logo, mime, err := service.LogoAktualisieren(aktuell, "image/png", service.LogoAenderung{Entfernen: true})
	if err != nil {
		t.Fatalf("LogoAktualisieren: %v", err)
	}
	if logo != nil || mime != "" {
		t.Errorf("Logo/Mime = %v/%q, erwartet entfernt (nil/\"\")", logo, mime)
	}
}

func TestLogoAktualisieren_HochgeladeneDateiErsetztDasVorhandeneLogo(t *testing.T) {
	aktuell := append(bytes.Clone(pngKennung), []byte("altes-logo")...)
	neu := append(bytes.Clone(jpegKennung), []byte("neues-logo")...)

	logo, mime, err := service.LogoAktualisieren(aktuell, "image/png", service.LogoAenderung{
		Hochgeladen: true, Dateiname: "logo.jpg", Inhalt: neu,
	})
	if err != nil {
		t.Fatalf("LogoAktualisieren: %v", err)
	}
	if !bytes.Equal(logo, neu) || mime != "image/jpeg" {
		t.Errorf("Logo/Mime = %v/%q, erwartet das neue Logo %v/image/jpeg", logo, mime, neu)
	}
}

// Eine neu ausgewählte Datei geht vor einem angekreuzten Entfernen — wer
// beides zugleich tut, wollte ersetzen und hat die Checkbox nur nicht wieder
// ausgeknipst.
func TestLogoAktualisieren_HochladenGehtVorEntfernen(t *testing.T) {
	aktuell := append(bytes.Clone(pngKennung), []byte("altes-logo")...)
	neu := append(bytes.Clone(jpegKennung), []byte("neues-logo")...)

	logo, mime, err := service.LogoAktualisieren(aktuell, "image/png", service.LogoAenderung{
		Entfernen: true, Hochgeladen: true, Dateiname: "logo.jpg", Inhalt: neu,
	})
	if err != nil {
		t.Fatalf("LogoAktualisieren: %v", err)
	}
	if !bytes.Equal(logo, neu) || mime != "image/jpeg" {
		t.Errorf("Logo/Mime = %v/%q, erwartet das neue Logo trotz Entfernen-Haken", logo, mime)
	}
}

func TestLogoAktualisieren_UngueltigeHochgeladeneDateiWeistAb(t *testing.T) {
	aktuell := append(bytes.Clone(pngKennung), []byte("altes-logo")...)

	_, _, err := service.LogoAktualisieren(aktuell, "image/png", service.LogoAenderung{
		Hochgeladen: true, Dateiname: "logo.txt", Inhalt: []byte("kein bild"),
	})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("LogoAktualisieren = %v, erwartet einen ValidierungsFehler", err)
	}
}

// Logo und LogoMime hängen an den Vereinsdaten (ADR-0012) und laufen durch
// SetVereinsdaten/GetVereinsdaten wie jede andere Angabe — nur eben binär statt
// Text.
func TestSetVereinsdaten_SpeichertDasLogoUnveraendert(t *testing.T) {
	svc := neuerService(t)

	logo := append(bytes.Clone(pngKennung), []byte("das-ist-ein-logo")...)

	daten := vereinsdatenImTest()
	daten.Logo = logo
	daten.LogoMime = "image/png"

	if err := svc.SetVereinsdaten(daten); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	gelesen, err := svc.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}
	if !bytes.Equal(gelesen.Logo, logo) {
		t.Errorf("Logo = %v, erwartet unverändert %v", gelesen.Logo, logo)
	}
	if gelesen.LogoMime != "image/png" {
		t.Errorf("LogoMime = %q, erwartet image/png", gelesen.LogoMime)
	}
}
