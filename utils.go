package main

import (
	"fmt"
	"strings"
)

func FirstCharUppercased(name string) string {
	firstChar := strings.ToUpper(string(name[0]))
	rest := name[1:]
	return fmt.Sprintf("%s%s", firstChar, rest)
}

func (k *Kontakt) GetKontaktHTML() string {
	content := `
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
				`

	mail := fmt.Sprintf("mailto:%s", k.Email)
	return fmt.Sprintf(content, k.Name, k.Message, mail, k.Email)
}

func (a *Anmeldung) GetAnmeldeHTML() string {
	return `
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
				`
}

func (a *Anmeldung) GetAnmeldeInfoHTML() string {
	content := `
				<head></head>
				<section style="max-width: 42rem; padding: 2rem 1.5rem; margin: 0 auto; background-color: #ffffff;">
				<header>
				<a href="https://musicschool-cml.de">
				<img style="width: auto; height: 1.75rem;" src="https://musicschool-cml.de/favicon.ico" alt="" />
				</a>
				</header>

				<main style="margin-top: 2rem;">
				<h2 style="color: #4a5568;">Neues Formular:</h2>

				<p style="margin-top: 0.5rem; line-height: 1.75; color: #718096;">
				Instrument: %s <br />
				Lehrer: %s <br />
				Schüler: %s <br />
				Geburtsdatum: %s <br />
				Straße: %s <br />
				PLZ: %s <br />
				Wohnort: %s <br />
				Erziehungsberechtigte: %s <br />
				Telefon: +49 %s <br />
				E-Mail: %s <br />
				</p>

				<p style="margin-top: 2rem; color: #718096;">
				Mit freundlichen Grüßen, <br />
				Ihr Musikschule CML Team
				</p>

				</main>
				</section>
				`

	return fmt.Sprintf(content, a.Instrument, a.Lehrer, a.SchuelerName, a.Geburtsdatum, a.Strasse, a.Plz, a.Wohnort, a.Erziehungsberechtigte, a.Telefon, a.Email)
}
