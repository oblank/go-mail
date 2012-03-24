package email

import (
	"bytes"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"time"
)

func NewBriefMessage(subject, content string, to ...string) *Message {
	message := &Message{Subject: subject, Content: content, To: make([]mail.Address, len(to))}
	for i := range to {
		message.To[i].Address = to[i]
	}
	return message
}

func NewBriefMessageFrom(subject, content, from string, to ...string) *Message {
	message := NewBriefMessage(subject, content, to...)
	message.From.Address = from
	return message
}

const crlf = "\r\n"

type Message struct {
	From    mail.Address // if From.Address is empty, Config.DefaultFrom will be used
	To      []mail.Address
	Cc      []mail.Address
	Bcc     []mail.Address
	Subject string
	Content string
}

// http://tools.ietf.org/html/rfc822
// http://tools.ietf.org/html/rfc2821
func (self *Message) String() string {
	var buf bytes.Buffer

	write := func(what string, recipients []mail.Address) {
		if len(recipients) == 0 {
			return
		}
		for i := range recipients {
			if i == 0 {
				buf.WriteString(what)
			} else {
				buf.WriteString(", ")
			}
			buf.WriteString(recipients[i].String())
		}
		buf.WriteString(crlf)
	}

	from := &self.From
	if from.Address == "" {
		from = &Config.DefaultFrom
	}
	fmt.Fprintf(&buf, "From: %s%s", from.String(), crlf)
	write("To: ", self.To)
	write("Cc: ", self.Cc)
	write("Bcc: ", self.Bcc)
	fmt.Fprintf(&buf, "Date: %s%s", time.Now().UTC().Format(time.RFC822), crlf)
	fmt.Fprintf(&buf, "Subject: %s%s%s", self.Subject, crlf, self.Content)
	return buf.String()
}

// Returns the first error
func (self *Message) Validate() error {
	if len(self.To) == 0 {
		return errors.New("Missing email recipient (email.Message.To)")
	}
	return nil
}

func (self *Message) Send() error {
	if err := self.Validate(); err != nil {
		return err
	}
	to := make([]string, len(self.To))
	for i := range self.To {
		to[i] = self.To[i].Address
	}
	from := self.From.Address
	if from == "" {
		from = Config.DefaultFrom.Address
	}
	addr := fmt.Sprintf("%s:%d", Config.Host, Config.Port)
	auth := smtp.PlainAuth("", Config.Username, Config.Password, Config.Host)
	return smtp.SendMail(addr, auth, from, to, []byte(self.String()))
}
