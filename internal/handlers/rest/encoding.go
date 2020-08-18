package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// EncodeJSONResponse will take a given interface and encode the value as JSON
func EncodeJSONResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return enc.Encode(response)
}

// EncodeError responds with a given error code with some additional logging information
func EncodeError(w http.ResponseWriter, code int, domain string, message string, method string) {
	logrus.WithFields(
		logrus.Fields{
			"type":   code,
			"domain": domain,
			"method": method,
		}).Error(strings.ToLower(message))
	http.Error(w, message, code)
}
