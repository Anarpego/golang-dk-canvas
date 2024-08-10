package handlers_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/matryer/is"
	"go.uber.org/zap"

	"canvas/handlers"
	"canvas/model"
)

type signupperMock struct {
	email model.Email
}

func (s *signupperMock) SignupForNewsletter(ctx context.Context, email model.Email) (string, error) {
	s.email = email
	return "123", nil
}

type senderMock struct {
	m model.Message
}

func (s *senderMock) Send(ctx context.Context, m model.Message) error {
	s.m = m
	return nil
}

func TestNewsletterSignup(t *testing.T) {
	mux := chi.NewMux()
	s := &signupperMock{}
	q := &senderMock{}
	handlers.NewsletterSignup(mux, s, q, zap.NewNop())

	t.Run("signs up a valid email address and sends a message", func(t *testing.T) {
		is := is.New(t)
		code, _, _ := makePostRequest(mux, "/newsletter/signup", createFormHeader(),
			strings.NewReader("email=me%40example.com"))
		is.Equal(http.StatusFound, code)
		is.Equal(model.Email("me@example.com"), s.email)

		is.Equal(q.m, model.Message{
			"job":   "confirmation_email",
			"email": "me@example.com",
			"token": "123",
		})
	})

	t.Run("rejects an invalid email address", func(t *testing.T) {
		is := is.New(t)
		code, _, _ := makePostRequest(mux, "/newsletter/signup", createFormHeader(),
			strings.NewReader("email=notanemail"))
		is.Equal(http.StatusBadRequest, code)
	})
}

type confirmerMock struct {
	token string
}

func (c *confirmerMock) ConfirmNewsletterSignup(ctx context.Context, token string) (*model.Email, error) {
	c.token = token
	email := model.Email("me@example.com")
	return &email, nil
}

func TestNewsletterConfirm(t *testing.T) {
	t.Run("confirms the newsletter signup and sends a message", func(t *testing.T) {
		is := is.New(t)
		mux := chi.NewMux()
		c := &confirmerMock{}
		q := &senderMock{}
		handlers.NewsletterConfirm(mux, c, q, zap.NewNop())

		code, _, _ := makePostRequest(mux, "/newsletter/confirm", createFormHeader(),
			strings.NewReader("token=123"))
		is.Equal(http.StatusFound, code)
		is.Equal("123", c.token)

		is.Equal(q.m, model.Message{
			"job":   "welcome_email",
			"email": "me@example.com",
		})
	})
}

func createFormHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	return header
}
