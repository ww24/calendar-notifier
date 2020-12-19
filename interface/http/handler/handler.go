package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ww24/calendar-notifier/usecase"
)

// New returns http handler.
func New(sync usecase.Synchronizer) http.Handler {
	mux := http.NewServeMux()
	svc := newService(sync)
	mux.HandleFunc("/", svc.defaultHandler)
	mux.HandleFunc("/launch", svc.sync)
	return mux
}

type syncService struct {
	syn usecase.Synchronizer
}

func newService(sync usecase.Synchronizer) *syncService {
	return &syncService{
		syn: sync,
	}
}

func (s *syncService) defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	res := map[string]interface{}{
		"status": "ok",
		"mode":   s.syn.RunningMode(),
	}
	d, err := json.Marshal(res)
	if err != nil {
		sendError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(append(d, '\n')); err != nil {
		sendError(w, r, err)
		return
	}
}

func (s *syncService) sync(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		return
	case http.MethodPost:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.syn.Sync(r.Context()); err != nil {
		sendError(w, r, err)
		return
	}

	res := map[string]interface{}{"status": "maybe ok"}
	d, err := json.Marshal(res)
	if err != nil {
		sendError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(append(d, '\n')); err != nil {
		sendError(w, r, err)
		return
	}
}

func sendError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"error":"%s"}`+"\n", err.Error())
	fmt.Fprintf(os.Stderr, `{"error":"%s"}`+"\n", err.Error())
}
