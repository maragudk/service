package email_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/maragudk/service/email"
)

func TestSender_SendGenericEmail(t *testing.T) {
	t.Run("returns error on status code 422 and errors from API", func(t *testing.T) {
		s, e := newSender(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, err := w.Write([]byte(`{"ErrorCode":100, "Message":"Datacenter burning."}`))
			require.NoError(t, err)
		})
		defer s.Close()

		err := e.SendGenericEmail(context.Background(), "You", "you@example.com", "Hi", "Hey there.", "Yo.")
		require.Error(t, err)
		require.Equal(t, "error sending email, got error code 100", err.Error())
	})

	t.Run("returns error on 300+ status code from API", func(t *testing.T) {
		s, e := newSender(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer s.Close()

		err := e.SendGenericEmail(context.Background(), "You", "you@example.com", "Hi", "Hey there.", "Yo.")
		require.Error(t, err)
		require.Equal(t, "error sending email, got status 500", err.Error())
	})

	t.Run("does not return error on inactive recipient", func(t *testing.T) {
		s, e := newSender(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, err := w.Write([]byte(`{"ErrorCode":406, "Message":"Blerp."}`))
			require.NoError(t, err)
		})
		defer s.Close()

		err := e.SendGenericEmail(context.Background(), "You", "you@example.com", "Hi", "Hey there.", "Yo.")
		require.NoError(t, err)
	})
}

func newSender(h http.HandlerFunc) (*httptest.Server, *email.Sender) {
	mux := chi.NewRouter()
	mux.Post("/email", h)
	s := httptest.NewServer(mux)
	e := email.NewSender(email.NewSenderOptions{
		BaseURL:                   "http://localhost:1234",
		EndpointURL:               s.URL + "/email",
		MarketingEmailAddress:     "marketing@example.com",
		MarketingEmailName:        "Marketer",
		Token:                     "123abc",
		TransactionalEmailAddress: "transactional@example.com",
		TransactionalEmailName:    "Transactionalizer",
	})
	return s, e
}
