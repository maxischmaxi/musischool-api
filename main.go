package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"github.com/resend/resend-go/v2"
)

type Anmeldung struct {
	Instrument            string `json:"instrument"`
	Lehrer                string `json:"lehrer"`
	Plz                   string `json:"plz"`
	Email                 string `json:"email"`
	SchuelerName          string `json:"schueler_name"`
	Wohnort               string `json:"ort"`
	Strasse               string `json:"strasse"`
	Erziehungsberechtigte string `json:"erziehungsberechtigte"`
	Telefon               string `json:"telefon"`
	Geburtsdatum          string `json:"geburtsdatum"`
	Vertrag               string `json:"vertrag"`
	Einverstaendnis       bool   `json:"einverstaendnis"`
	Token                 string `json:"token"`
}

type Kontakt struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

func FirstCharUppercased(name string) string {
	firstChar := strings.ToUpper(string(name[0]))
	rest := name[1:]
	return fmt.Sprintf("%s%s", firstChar, rest)
}

func ValidateToken(token string, secret string) bool {
	uri, err := url.Parse(fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s", secret, token))

	if err != nil {
		return false
	}

	resp, err := http.Get(uri.String())

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return false
	}

	var response map[string]interface{}

	err = json.Unmarshal(body, &response)

	if err != nil {
		return false
	}

	return response["success"].(bool)
}

func GeneratePdf(anmeldung Anmeldung) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	variantenString := ""
	switch anmeldung.Vertrag {
	case "1":
		variantenString = "30 Minuten Einzelunterricht"
	case "2":
		variantenString = "45 Minuten Einzelunterricht"
	case "3":
		variantenString = "60 Minuten Gruppenunterricht"
	case "4":
		variantenString = "45 Minuten Einzelunterricht"
	default:
		break
	}

	title := tr(fmt.Sprintf("Anmeldung und Unterrichtsvertrag zum %s", variantenString))

	pdf.SetTitle(title, true)
	pdf.SetAuthor("Musicschool CML", true)
	pdf.SetHeaderFunc(func() {
		pdf.Image("logo.png", 10, 10, 0, 0, false, "", 0, "")
		pdf.SetFont("Arial", "B", 12)
		pdf.Text(180, 10, time.Now().Format("02.01.2006"))
	})
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	pdf.Text(20, 55, fmt.Sprintf("Instrument: %s", tr(FirstCharUppercased(anmeldung.Instrument))))
	pdf.Text(110, 55, fmt.Sprintf("Lehrer: %s", tr(FirstCharUppercased(anmeldung.Lehrer))))
	pdf.Text(20, 61, fmt.Sprintf("%s: %s", tr("Straße"), tr(anmeldung.SchuelerName)))
	pdf.Text(110, 61, fmt.Sprintf("Geburtsdatum: %s", anmeldung.Geburtsdatum))
	pdf.Text(20, 67, fmt.Sprintf("%s: %s", tr("Straße"), tr(anmeldung.Strasse)))
	pdf.Text(110, 67, fmt.Sprintf("PLZ: %s", anmeldung.Plz))
	pdf.Text(20, 73, fmt.Sprintf("Wohnort: %s", tr(anmeldung.Wohnort)))
	pdf.Text(110, 73, fmt.Sprintf("Erziehungsberechtigte: %s", tr(anmeldung.Erziehungsberechtigte)))
	pdf.Text(20, 79, fmt.Sprintf("Telefon: +49 %s", anmeldung.Telefon))
	pdf.Text(110, 79, fmt.Sprintf("E-Mail: %s", anmeldung.Email))

	var lines []string

	if anmeldung.Vertrag == "1" || anmeldung.Vertrag == "2" || anmeldung.Vertrag == "3" {
		var preis string
		var minuten string

		switch anmeldung.Vertrag {
		case "1":
			preis = "88"
			minuten = "30"
		case "2":
			preis = "111"
			minuten = "45"
		case "3":
			preis = "66"
			minuten = "60"
		default:
			preis = "111"
			minuten = "45"
		}

		lines = []string{
			"Die Musikschule übernimmt den regelmäßigen Unterricht des Schülers beginnend am ...........................",
			fmt.Sprintf("Als Unterrichtsjahr gilt das Kalenderjahr. Der Unterricht wird als Lektion zu wöchentlich einmal %s", minuten),
			fmt.Sprintf("Minuten erteilt, monatliche Gebühr = %s,- Euro.", preis),
			"Das Honorar wird als Jahreshonorar berechnet und ist in 12 gleichen Raten im Voraus bis zum 10.",
			"jeden Monats zu zahlen, einmalige Aufnahmegebühr: 20,- Euro. Der Unterricht kann nur an",
			"Schultagen erteilt werden. Bei Rücklastschriften berechnen wir 10,-€ pro nicht einlösbarer",
			"Lastschrift. Die erste Unterrichtsstunde ist ein Gratis-Probeunterricht, die vereinbarte Zeit gilt für alle folgenden",
			"Stunden. Will ein Schüler den Unterricht nach der kostenlosen Probestunde nicht fortsetzen, genügt",
			"eine entsprechende mündliche Mitteilung. Bei längerer Krankheit des Schülers entfällt das anteilige",
			"Honorar nach der vierten einander folgenden versäumten Stunde.",
			"Der Kurs kann von den Vertragspartnern mit sechswöchiger Frist zum 30.April/ 31.August/",
			"31.Dezember in schriftlicher Form gekündigt werden. Die Kündigung kann durch eine E-Mail",
			"erfolgen und muß vor Beginn der Kündigungsfrist bei o.g. Anschrift eingegangen sein. Eine",
			"Erhöhung des Honorars ist möglich und hat nach Grundsätzen der Billigkeit zu erfolgen. Sie muß",
			"mindestens 8 Wochen vorher dem Vertragspartner schriftlich mitgeteilt werden.",
			"Für vom Schüler versäumte oder abgesagte Stunden ist die Lehrkraft nicht nachleistungspflichtig,",
			"die anteilige Vergütung hierfür kann vom Honorar nicht abgezogen werden. Es besteht jedoch die",
			"Möglichkeit, in derselben Woche ersatzweise an einer anderen Unterrichtsstunde teilzunehmen,",
			"wenn die Lehrkraft im Falle ernsthafter Verhinderung mindestens 24 Stunden vorher davon",
			"Kenntnis erhalten hat. Aus anderen Gründen von der Lehrkraft abgesagte Stunden werden",
			"nachgegeben, ersatzweise wird das anteilige Honorar erstattet. Zahlungsweise: nur monatlich durch",
			"Einzugsverfahren. Änderungen und Ergänzungen des Vertrages sind nur wirksam, wenn sie",
			"schriftlich erfolgen. Werden einzelne Bestimmungen dieses Vertrages unwirksam, wird dadurch die",
			"Gültigkeit des Vertrages im Übrigen nicht berührt.",
		}
	} else {
		lines = []string{
			"Der Unterricht wird als 10-stündige Lektion zu jeweils 45 Minuten bei freier Vereinbarung des",
			"Zeitpunktes in Absprache mit der zuständigen Lehrkraft erteilt, einmahlige Gebühr = 450,- Euro.",
			"Aufnahmegebühr: 20,- Euro.",
			"Unterrichtsstunden können abgesagt werden, wenn die Lehrkraft im Falle ernsthafter Verhinderung",
			"mindestens 24 Stunden vorher davon Kenntnis erhalten hat, andernfalls gelten sie als gegeben.",
			"Zahlungsweise: per Überweisung oder durch Einzugsverfahren. Änderungen und Ergänzungen des",
			"Vertrages sind nur wirksam, wenn sie schriftlich erfolgen. Werden einzelne Bestimmungen dieses",
			"Vertrages unwirksam, wird dadurch die Gültigkeit des Vertrages im Übrigen nicht berührt.",
		}
	}

	y := 90.

	pdf.SetFont("Arial", "", 10)
	for _, line := range lines {
		pdf.Text(20, y, tr(line))
		y += 5
	}

	y += 10
	pdf.SetFont("Arial", "B", 11)
	pdf.Text(20, y, tr("Ermächtigung zum Einzug von Unterrichtsgebühren durch Lastschrift:"))

	y += 5
	pdf.SetFont("Arial", "", 11)
	pdf.Text(20, y, tr("Hiermit ermächtige ich Sie widerruflich, die von mir zu entrichtenden Unterrichtsgebühren"))
	y += 5
	pdf.Text(20, y, tr("beginnend ab ........................... bei Fälligkeit zu Lasten meines Kontos"))
	y += 7
	pdf.Text(20, y, tr("IBAN: ................................................................................"))
	y += 7
	pdf.Text(20, y, tr("durch Lastschrift einzuziehen. Wenn mein Konto die erforderliche Deckung nicht aufweist, besteht"))
	y += 5
	pdf.Text(20, y, tr("seitens des kontoführenden Kreditinstitutes keine Verpflichtung zur Einlösung."))
	y += 10
	pdf.Text(20, y, tr("......................................................"))
	pdf.Text(110, y, tr("........................................................................"))
	y += 5
	pdf.Text(20, y, tr("Ort, Datum"))
	pdf.Text(110, y, tr("Unterschrift Kontoinhaber"))
	y += 10
	pdf.Text(20, y, tr("........................................................................"))
	pdf.Text(110, y, tr("........................................................................"))
	y += 5
	pdf.Text(20, y, tr("Unterschrift Erziehungsberechtigte"))
	pdf.Text(110, y, tr("Unterschrift Musikschule CML"))

	var buffer bytes.Buffer
	err := pdf.Output(&buffer)

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func main() {
	resendApiKey := os.Getenv("RESEND_API_KEY")
	receiver := os.Getenv("RESEND_RECEIVER")
	recaptchaSecret := os.Getenv("RECAPTCHA_SECRET")
	client := resend.NewClient(resendApiKey)

	log.Println("Starting server...")
	log.Println("Receiver: ", receiver)
	log.Println("Resend API Key: ", resendApiKey)
	log.Println("Recaptcha Secret: ", recaptchaSecret)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://musicschool-cml.de", "https://www.musicschool-cml.de"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Type", "Content-Length"},
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	router.POST("/anmeldung", func(c *gin.Context) {
		var anmeldung Anmeldung
		err := c.BindJSON(&anmeldung)

		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if !ValidateToken(anmeldung.Token, recaptchaSecret) {
			log.Println("Recaptcha validation failed")
			c.JSON(400, gin.H{"error": "Recaptcha validation failed"})
			return
		}

		if !anmeldung.Einverstaendnis {
			log.Println("Einverständniserklärung nicht akzeptiert")
			c.JSON(400, gin.H{"error": "Einverständniserklärung nicht akzeptiert"})
			return
		}

		pdfBytes, err := GeneratePdf(anmeldung)

		if err != nil {
			log.Print(err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		var attachments = []*resend.Attachment{
			{
				Filename:    fmt.Sprintf("anmeldung-%s.pdf", anmeldung.SchuelerName),
				Content:     pdfBytes,
				ContentType: "application/pdf",
			},
		}

		params := &resend.SendEmailRequest{
			From:        "Musikschule CML <mail@mail.jeschek.dev>",
			To:          []string{anmeldung.Email},
			Subject:     "Anmeldung und Unterrichtsvertrag Musicschool CML",
			Attachments: attachments,
			Html: `
				<head></head>
				<section style="max-width: 42rem; padding: 2rem 1.5rem; margin: 0 auto; background-color: #ffffff;">
				<header>
				<a href="https://musicschool-cml.de">
				<img style="width: auto; height: 1.75rem;" src="https://musicschool-cml.de/favicon.ico" alt="" />
				</a>
				</header>

				<main style="margin-top: 2rem;">
				<h2 style="color: #4a5568;">Sehr geehrte Damen und Herren,</h2>

				<p style="margin-top: 0.5rem; line-height: 1.75; color: #718096;">
				Vielen Dank für das Ausfüllen des Anmeldeformulars auf musicschool-cml.de. <br />
				Hiermit erhalten Sie den Unterrichtsvertrag zum Download. <br />
				Bitte füllen Sie diesen aus und bringen Sie ihn zum ersten Unterricht (nach dem kostenlosen Probeunterricht) mit. <br /> <br />
				Danach wird der Vertrag von uns gegengezeichnet und Sie erhalten eine Kopie. <br />
				</p>

				<p style="margin-top: 2rem; color: #718096;">
				Mit freundlichen Grüßen, <br />
				Ihr Musikschule CML Team
				</p>

				</main>
				</section>
				`,
		}

		send, err := client.Emails.Send(params)

		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "OK", "id": send.Id})
	})

	router.POST("/contact", func(c *gin.Context) {
		var kontakt Kontakt
		err := c.BindJSON(&kontakt)

		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if !ValidateToken(kontakt.Token, recaptchaSecret) {
			log.Println("Recaptcha validation failed")
			c.JSON(400, gin.H{"error": "Recaptcha validation failed"})
			return
		}

		params := &resend.SendEmailRequest{
			From:    "Musikschule CML <mail@mail.jeschek.dev>",
			To:      []string{receiver},
			Subject: "Kontaktformular - musicschool-cml.de",
			Html: fmt.Sprintf(`
				<head></head>
				<section style="max-width: 42rem; padding: 2rem 1.5rem; margin: 0 auto; background-color: #ffffff;">
				<header>
				<a href="https://musicschool-cml.de">
				<img style="width: auto; height: 1.75rem;" src="https://musicschool-cml.de/favicon.ico" alt="" />
				</a>
				</header>

				<main style="margin-top: 2rem;">
				<h2 style="color: #4a5568;">Hi Jana,</h2>

				<p style="margin-top: 0.5rem; line-height: 1.75; color: #718096;">
				Das Kontaktformular auf musicschool-cml.de wurde von
				<span style="font-weight: 600;">%s</span> ausgefüllt.
				</p>

				<p style="margin-top: 2rem; color: #718096;">
				Die Nachricht:
				<br />
				%s
				</p>

				<p style="margin-top: 2rem; color: #718096;">
				Du kannst
				<a href="%s" style="color: #3182ce; text-decoration: underline;">
				%s
				</a>
				direkt antworten.
				</p>
				</main>
				</section>
				`, kontakt.Name, kontakt.Message, fmt.Sprintf("mailto:%s", kontakt.Email), kontakt.Email),
		}

		send, err := client.Emails.Send(params)

		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": send.Id})
	})

	log.Fatal(router.Run(":8080"))
}
