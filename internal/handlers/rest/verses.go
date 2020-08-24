package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/holmes89/lectionary/internal"
	"github.com/sirupsen/logrus"
)

type verseHandler struct {
	service internal.VerseService
}

func NewVerseHandler(mr *mux.Router, service internal.VerseService) http.Handler {
	r := mr.PathPrefix("/verse").Subrouter()

	h := &verseHandler{
		service: service,
	}

	r.HandleFunc("/", h.FindAll).Methods("GET")

	return r
}

func (h *verseHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	q, ok := params["q"]
	if !ok {
		EncodeError(w, http.StatusBadRequest, "verse", "missing query param", "findall")
		return
	}
	results, err := h.service.Find(q[0], internal.NIV) //multiple searches?
	if err != nil {
		logrus.WithError(err).Error("unable to find results")
		EncodeError(w, http.StatusInternalServerError, "verse", "failed to find verse", "findall")
		return
	}
	EncodeJSONResponse(r.Context(), w, results)
}
