package main

import (
	"bufio"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	var dial = "127.0.0.1:25"

	fmt.Print("dial: ")
	dial1, _ := reader.ReadString('\n')
	dial1 = strings.Replace(dial1, "\n", "", -1)
	if dial1 != "" {
		dial = dial1
	}

	// Connect to the remote SMTP server.
	client, err := smtp.Dial(dial)
	if err != nil {
		log.Fatal(err)
	}

	var sender string
	fmt.Print("sender: ")
	sender, _ = reader.ReadString('\n')
	sender = strings.Replace(sender, "\n", "", -1)

	var recipient string
	fmt.Print("recipient: ")
	recipient, _ = reader.ReadString('\n')
	recipient = strings.Replace(recipient, "\n", "", -1)

	var attach string
	fmt.Print("attach: ")
	attach, _ = reader.ReadString('\n')
	attach = strings.Replace(attach, "\n", "", -1)

	var config configMail
	config.Subject = "Hello!"

	if sender != "" {
		config.From = mail.Address{
			Name:    "sender",
			Address: sender,
		}
	}

	if recipient != "" {
		config.To = append(config.To, mail.Address{
			Name:    "recipient",
			Address: recipient,
		})
	}

	config.Cc = config.To

	config.Body.ContentType = "text/html"
	config.Body.Body = "Hello <b>Bob</b> and <i>Cora</i>!"

	if attach != "" {
		config.Attach = append(config.Attach, attach)
	}
	config.Embed = config.Attach

	message := gomail.NewMessage()
	senderFalse, recipientFalse := convertMailForm(message, config, client)

	// Send the email body.
	wc, err := client.Data()
	if err != nil {
		log.Fatal(err)
	}

	_, err = message.WriteTo(wc)
	if err != nil {
		log.Fatal(err)
	}

	err = wc.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Send the QUIT command and close the connection.
	err = client.Quit()
	if err != nil {
		log.Fatal(err)
	}

	if senderFalse != "" {
		log.Fatal("senderFalse: ", senderFalse)
	}

	if len(recipientFalse) > 0 {
		log.Fatal("recipientFalse: ", recipientFalse)
	}
}

type configMail struct {
	Subject string         `json:"subject"`
	From    mail.Address   `json:"from"`
	To      []mail.Address `json:"to"`
	Cc      []mail.Address `json:"cc"`
	Bcc     []mail.Address `json:"bcc"`
	Body    struct {
		ContentType string `json:"content_type"`
		Body        string `json:"body"`
	} `json:"body"`
	Attach []string `json:"attach"`
	Embed  []string `json:"embed"`
}

func convertMailForm(message *gomail.Message, config configMail, client *smtp.Client) (senderFalse string, recipientFalse map[string]bool) {
	// make default
	recipientFalse = map[string]bool{}

	message.SetHeader("Subject", config.Subject)

	headers := make(map[string][]string)

	headers["From"] = []string{config.From.String()}
	err := client.Mail(config.From.Address)
	if err != nil {
		senderFalse = config.From.Address
	}

	var checkDuplicate = make(map[string]bool)

	for _, people := range config.To {
		headers["To"] = append(headers["To"], people.String())
		var address = people.Address
		if !checkDuplicate[address] {
			err = client.Rcpt(address)
			if err != nil {
				recipientFalse[address] = true
			}
			checkDuplicate[address] = true
		}
	}
	for _, people := range config.Cc {
		headers["Cc"] = append(headers["Cc"], people.String())
		var address = people.Address
		if !checkDuplicate[address] {
			err = client.Rcpt(address)
			if err != nil {
				recipientFalse[address] = true
			}
			checkDuplicate[address] = true
		}
	}
	for _, people := range config.Bcc {
		headers["Bcc"] = append(headers["Bcc"], people.String())
		var address = people.Address
		if !checkDuplicate[address] {
			err = client.Rcpt(address)
			if err != nil {
				recipientFalse[address] = true
			}
			checkDuplicate[address] = true
		}
	}

	message.SetHeaders(headers)

	message.SetBody(config.Body.ContentType, config.Body.Body)

	var checkDuplicateFile = make(map[string]bool)

	for _, embed := range config.Embed {
		if !checkDuplicateFile[embed] {
			message.Attach(embed)
			checkDuplicateFile[embed] = true
		}
	}

	for _, attach := range config.Attach {
		if !checkDuplicateFile[attach] {
			message.Attach(attach)
			checkDuplicateFile[attach] = true
		}
	}
	return
}
