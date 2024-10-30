package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v2"
)

type Kontakt struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

func main() {
	resendApiKey := os.Getenv("RESEND_API_KEY")
	receiver := os.Getenv("RESEND_RECEIVER")
	client := resend.NewClient(resendApiKey)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:    []string{"localhost:5173", "musicschool-cml.de", "*"},
		AllowMethods:    []string{"GET", "POST", "OPTIONS", "PUT"},
		AllowHeaders:    []string{"Origin", "Content-Type"},
		AllowWebSockets: false,
		MaxAge:          12 * time.Hour,
	}))

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	router.POST("/contact", func(c *gin.Context) {
		var kontakt Kontakt
		err := c.BindJSON(&kontakt)

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
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
			c.JSON(500, gin.H{"error": err.Error()})
		} else {
			c.JSON(200, gin.H{"message": send.Id})
		}
	})

	log.Fatal(router.Run(":8080"))
}
