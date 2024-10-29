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
		AllowOrigins: []string{"http://localhost:5173", "https://musicschool-cml.de", "https://www.musicschool-cml.de"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
		MaxAge:       12 * time.Hour,
	}))

	router.GET("/health", func(c *gin.Context) {
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
				<section className="max-w-2xl px-6 py-8 mx-auto bg-white dark:bg-gray-900">
				<header>
				<a href="https://musicschool-cml.de">
				<img className="w-auto h-7 sm:h-8" {...logo} alt="" />
				</a>
				</header>

				<main className="mt-8">
				<h2 className="text-gray-700 dark:text-gray-200">Hi Jana,</h2>

				<p className="mt-2 leading-loose text-gray-600 dark:text-gray-300">
				Das Kontaktformular auf musicschool-cml.de wurde von
				<span className="font-semibold ">%s</span> ausgefüllt.
				</p>

				<p className="mt-8 text-gray-600 dark:text-gray-300">
				Die Nachricht:
				<br />
				%s
				</p>

				<p className="mt-8 text-gray-600 dark:text-gray-300">
				Du kannst
				<a
				href="%s"
				className="text-blue-600 hover:underline dark:text-blue-400"
				>
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
