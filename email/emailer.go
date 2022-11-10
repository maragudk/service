package messaging

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/maragudk/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/maragudk/service/model"
)

const (
	marketingMessageStream     = "broadcast"
	transactionalMessageStream = "outbound"
)

type emailType int

const (
	marketing emailType = iota
	transactional
)

// nameAndEmail combo, of the form "Name <email@example.com>"
type nameAndEmail = string

type keywords = map[string]string

// Emailer can send transactional and marketing emails through Postmark.
// See https://postmarkapp.com/developer
type Emailer struct {
	baseURL           string
	client            *http.Client
	emailCount        *prometheus.CounterVec
	endpointURL       string
	log               *log.Logger
	marketingFrom     nameAndEmail
	token             string
	transactionalFrom nameAndEmail
}

type NewEmailerOptions struct {
	BaseURL                   string
	EndpointURL               string
	Log                       *log.Logger
	MarketingEmailAddress     string
	MarketingEmailName        string
	Metrics                   *prometheus.Registry
	Token                     string
	TransactionalEmailAddress string
	TransactionalEmailName    string
}

func NewEmailer(opts NewEmailerOptions) *Emailer {
	if opts.Log == nil {
		opts.Log = log.New(io.Discard, "", 0)
	}

	if opts.Metrics == nil {
		opts.Metrics = prometheus.NewRegistry()
	}

	if opts.EndpointURL == "" {
		opts.EndpointURL = "https://api.postmarkapp.com/email"
	}

	emailCount := promauto.With(opts.Metrics).NewCounterVec(prometheus.CounterOpts{
		Name: "app_emails_total",
	}, []string{"name", "success"})

	return &Emailer{
		baseURL:           strings.TrimSuffix(opts.BaseURL, "/"),
		client:            &http.Client{Timeout: 3 * time.Second},
		emailCount:        emailCount,
		endpointURL:       opts.EndpointURL,
		log:               opts.Log,
		marketingFrom:     createNameAndEmail(opts.MarketingEmailName, opts.MarketingEmailAddress),
		token:             opts.Token,
		transactionalFrom: createNameAndEmail(opts.TransactionalEmailName, opts.TransactionalEmailAddress),
	}
}

func (e *Emailer) SendGenericEmail(ctx context.Context, name string, email model.Email, subject, preheader, content string) error {
	return e.send(ctx, transactional, createNameAndEmail(name, email.String()), subject, preheader, "generic", keywords{
		"baseURL": e.baseURL,
		"title":   subject,
		"content": content,
	})
}

// requestBody used in Emailer.send.
// See https://postmarkapp.com/developer/user-guide/send-email-with-api
type requestBody struct {
	MessageStream string
	From          nameAndEmail
	To            nameAndEmail
	Subject       string
	TextBody      string
	HtmlBody      string
}

func (e *Emailer) send(ctx context.Context, typ emailType, to nameAndEmail, subject, preheader, template string, keywords keywords) error {
	var messageStream string
	var from nameAndEmail
	switch typ {
	case marketing:
		from = e.marketingFrom
		messageStream = marketingMessageStream
	case transactional:
		from = e.transactionalFrom
		messageStream = transactionalMessageStream
	}

	err := e.sendRequest(ctx, requestBody{
		MessageStream: messageStream,
		From:          from,
		To:            to,
		Subject:       subject,
		TextBody:      getEmail(template+".txt", preheader, keywords),
		HtmlBody:      getEmail(template+".html", preheader, keywords),
	})

	e.emailCount.WithLabelValues(template, strconv.FormatBool(err == nil)).Inc()

	return err
}

type postmarkResponse struct {
	ErrorCode int
	Message   string
}

// send using the Postmark API.
func (e *Emailer) sendRequest(ctx context.Context, body requestBody) error {
	bodyAsBytes, err := json.Marshal(body)
	if err != nil {
		return errors.Wrap(err, "error marshalling request body to json")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpointURL, bytes.NewReader(bodyAsBytes))
	if err != nil {
		return errors.Wrap(err, "error creating request")
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Postmark-Server-Token", e.token)

	response, err := e.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "error making request")
	}
	defer func() {
		_ = response.Body.Close()
	}()
	bodyAsBytes, err = io.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	// https://postmarkapp.com/developer/api/overview#response-codes
	if response.StatusCode == http.StatusUnprocessableEntity {
		var r postmarkResponse
		if err := json.Unmarshal(bodyAsBytes, &r); err != nil {
			return errors.Wrap(err, "error unwrapping postmark error response body")
		}

		// https://postmarkapp.com/developer/api/overview#error-codes
		switch r.ErrorCode {
		case 406:
			e.log.Printf("Not sending email, recipient %v is inactive.", body.To)
			return nil
		default:
			e.log.Println("Error sending email:", r.Message, "; error code:", r.ErrorCode)
			return errors.Newf("error sending email, got error code %v", r.ErrorCode)
		}
	}

	if response.StatusCode > 299 {
		e.log.Println("Error sending email:", response.StatusCode, string(bodyAsBytes))
		return errors.Newf("error sending email, got status %v", response.StatusCode)
	}

	return nil
}

// createNameAndEmail returns a name and email string ready for inserting into From and To fields.
func createNameAndEmail(name, email string) nameAndEmail {
	return fmt.Sprintf("%v <%v>", name, email)
}

//go:embed emails
var emails embed.FS

// getEmail from the given path, panicking on errors.
// It also replaces keywords given in the map.
// Email preheader text should be between 40-130 characters long.
func getEmail(path, preheader string, keywords keywords) string {
	emailBody, err := emails.ReadFile("emails/" + path)
	if err != nil {
		panic(err)
	}

	layout, err := emails.ReadFile("emails/layout" + filepath.Ext(path))
	if err != nil {
		panic(err)
	}

	email := string(layout)
	email = strings.ReplaceAll(email, "{{preheader}}", preheader)
	email = strings.ReplaceAll(email, "{{body}}", string(emailBody))
	for keyword, replacement := range keywords {
		email = strings.ReplaceAll(email, "{{"+keyword+"}}", replacement)
	}

	if _, ok := keywords["unsubscribe"]; ok {
		email = strings.ReplaceAll(email, "{{unsubscribe}}", "{{{ pm:unsubscribe }}}")
	} else {
		email = strings.ReplaceAll(email, "{{unsubscribe}}", "")
	}

	return email
}
