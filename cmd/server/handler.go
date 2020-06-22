package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ww24/calendar-notifier/config"
)

type handler struct {
	conf    config.Config
	handler func() (map[string]interface{}, error)
}

func newHandler(conf config.Config, hf func() (map[string]interface{}, error)) *handler {
	return &handler{
		conf:    conf,
		handler: hf,
	}
}

func (h *handler) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.defaultHandler)
	mux.HandleFunc("/launch", h.launchHandler)
	return mux
}

func (h *handler) defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	res := map[string]interface{}{
		"status": "ok",
		"mode":   h.conf.Mode,
	}
	d, err := json.Marshal(res)
	if err != nil {
		sendError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(d))
}

func (h *handler) launchHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		return
	case http.MethodPost:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	res, err := h.handler()
	if err != nil {
		sendError(w, r, err)
		return
	}
	d, err := json.Marshal(res)
	if err != nil {
		sendError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(d))
}

func sendError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"error":"%s"}`+"\n", err.Error())
	fmt.Fprintf(os.Stderr, `{"error":"%s"}`+"\n", err.Error())
}
