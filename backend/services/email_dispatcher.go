package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
)

func SendInvoiceEmail(recipientEmail string, pdfBytes []byte) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")

	encodedFile := base64.StdEncoding.EncodeToString(pdfBytes)

	//Construct the Multipart MIME Email
	boundary := "my-custom-boundary"
	body := new(bytes.Buffer)

	//Email Headers
	body.WriteString(fmt.Sprintf("From: %s\r\n", fromEmail))
	body.WriteString(fmt.Sprintf("To: %s\r\n", recipientEmail))
	body.WriteString("Subject: Your Medieval Store Invoice\r\n")
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))

	//Email Text Body
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
	body.WriteString("Thank you for your purchase! Your invoice is attached.\r\n\r\n")

	//Email PDF Attachment
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Type: application/pdf; name=\"invoice.pdf\"\r\n")
	body.WriteString("Content-Transfer-Encoding: base64\r\n")
	body.WriteString("Content-Disposition: attachment; filename=\"invoice.pdf\"\r\n\r\n")

	for i := 0; i < len(encodedFile); i += 76 {
		end := i + 76
		if end > len(encodedFile) {
			end = len(encodedFile)
		}
		body.WriteString(encodedFile[i:end] + "\r\n")
	}
	body.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	//Authenticate and Send using the Mailtrap Username and Password
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, []string{recipientEmail}, body.Bytes())

	return err
}

func SendDiscountNotificationEmail(recipientEmail, recipientName, productName string, originalPrice, discount float64) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")

	salePrice := originalPrice * (1 - discount/100)

	body := new(bytes.Buffer)
	body.WriteString(fmt.Sprintf("From: %s\r\n", fromEmail))
	body.WriteString(fmt.Sprintf("To: %s\r\n", recipientEmail))
	body.WriteString("Subject: A wishlist item is on sale\r\n")
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
	body.WriteString(fmt.Sprintf("Hi %s,\r\n\r\n", recipientName))
	body.WriteString(fmt.Sprintf("Great news! \"%s\" from your wishlist is now %.0f%% off.\r\n", productName, discount))
	body.WriteString(fmt.Sprintf("Was: %.2f gold pieces\r\n", originalPrice))
	body.WriteString(fmt.Sprintf("Now: %.2f gold pieces\r\n\r\n", salePrice))
	body.WriteString("Visit the Medieval Store to claim yours before it's gone!\r\n")

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, []string{recipientEmail}, body.Bytes())
}
