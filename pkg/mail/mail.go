package mail

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSender(host string, port int, username, password, from string) *Sender {
	return &Sender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *Sender) Send(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)
	return d.DialAndSend(m)
}

func (s *Sender) SendVerification(to, token, frontendURL string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)
	body := fmt.Sprintf(`
<h2>Confirm your email</h2>
<p>Click the link below to verify your email address:</p>
<p><a href="%s">Verify Email</a></p>
<p>This link expires in 24 hours.</p>
<p>If you did not register on SHIBA, ignore this email.</p>
`, link)
	return s.Send(to, "Verify your SHIBA email", body)
}

func (s *Sender) SendPhotoRejected(to, reason string) error {
	body := fmt.Sprintf(`
<h2>Photo Moderation Update</h2>
<p>Unfortunately, one of your photos has been rejected.</p>
<p><strong>Reason:</strong> %s</p>
<p>Please upload a new photo that complies with our guidelines.</p>
`, reason)
	return s.Send(to, "Your photo was rejected", body)
}

func (s *Sender) SendApplicationStatusChanged(to, castingTitle, status string) error {
	body := fmt.Sprintf(`
<h2>Application Status Update</h2>
<p>Your application for <strong>%s</strong> has been updated.</p>
<p><strong>New status:</strong> %s</p>
`, castingTitle, status)
	return s.Send(to, "Application status updated", body)
}

func (s *Sender) SendNewInvitation(to, agencyName, message string) error {
	body := fmt.Sprintf(`
<h2>New Invitation from %s</h2>
<p>%s</p>
<p>Log in to SHIBA to view and respond to this invitation.</p>
`, agencyName, message)
	return s.Send(to, "You have a new invitation on SHIBA", body)
}

func (s *Sender) SendNewApplication(to, modelName, castingTitle string) error {
	body := fmt.Sprintf(`
<h2>New Application Received</h2>
<p><strong>%s</strong> has applied to your casting: <strong>%s</strong></p>
<p>Log in to SHIBA to review the application.</p>
`, modelName, castingTitle)
	return s.Send(to, "New application for your casting", body)
}
