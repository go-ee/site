package main

import (
	"github.com/go-ee/utils/email"
	"github.com/sirupsen/logrus"
	"net/http"
)

func EmailSupport(sender *email.Sender) http.Handler {
	return &emailHandler{}
}

type emailHandler struct {
}

func (f *emailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Warn("email support, handle: %v, %v", w, r)
}
