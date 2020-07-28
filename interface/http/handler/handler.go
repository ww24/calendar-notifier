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
	mux.HandleFunc("/launch", svc.sync) // TODO: change to sync
	mux.HandleFunc("/notify", svc.notify)
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
	w.Write(append(d, '\n'))
}

func (s *syncService) sync(w http.ResponseWriter, r *http.Request) {
	// TODO: IAM 認証

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
	w.Write(append(d, '\n'))
}

func (s *syncService) notify(w http.ResponseWriter, r *http.Request) {
	// TODO: api key による認証

	if r.Header.Get("content-type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// FIXME: DEBUG出力
	m := make(map[string]interface{})
	json.NewDecoder(r.Body).Decode(&m)
	fmt.Printf("%+v\n", m)

	w.WriteHeader(http.StatusNoContent)
}

func sendError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"error":"%s"}`+"\n", err.Error())
	fmt.Fprintf(os.Stderr, `{"error":"%s"}`+"\n", err.Error())
}
