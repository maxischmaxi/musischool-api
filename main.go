package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dpapathanasiou/go-recaptcha"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

type Config struct {
	Receiver string
}

func main() {
	config := Config{
		Receiver: os.Getenv("RESEND_RECEIVER"),
	}

	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	recaptcha.Init(os.Getenv("RECAPTCHA_SECRET"))
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Type", "Content-Length"},
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.POST("/anmeldung", func(c *gin.Context) {
		var anmeldung Anmeldung
		err := c.BindJSON(&anmeldung)

		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		success, err := recaptcha.Confirm(c.ClientIP(), anmeldung.Token)

		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{"error": "Recaptcha validation failed"})
			return
		}

		if !success {
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
			Html:        anmeldung.GetAnmeldeHTML(),
		}

		_, err = client.Emails.Send(params)

		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		params = &resend.SendEmailRequest{
			From:    "Musikschule CML <mail@mail.jeschek.dev>",
			To:      []string{config.Receiver},
			Subject: "Anmeldung und Unterrichtsvertrag Musicschool CML",
			Html:    anmeldung.GetAnmeldeInfoHTML(),
		}

		_, err = client.Emails.Send(params)

		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "ok"})
	})

	router.POST("/contact", func(c *gin.Context) {
		var kontakt Kontakt
		err := c.BindJSON(&kontakt)

		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		success, err := recaptcha.Confirm(c.ClientIP(), kontakt.Token)

		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{"error": "Recaptcha validation failed"})
			return
		}

		if !success {
			log.Println("Recaptcha validation failed")
			c.JSON(400, gin.H{"error": "Recaptcha validation failed"})
			return
		}

		params := &resend.SendEmailRequest{
			From:    "Musikschule CML <mail@mail.jeschek.dev>",
			To:      []string{config.Receiver},
			Subject: "Kontaktformular - musicschool-cml.de",
			Html:    kontakt.GetKontaktHTML(),
		}

		_, err = client.Emails.Send(params)

		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "ok"})
	})

	log.Fatal(router.Run(":8080"))
}
