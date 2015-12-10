package receiver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	defaultHTTPReceiverAddress string = "127.0.0.1:8081"
)

type httpReceiver struct {
	Address string `json:"address"`

	name    string
	log     Log
	storage Storage
}

func NewHTTPReceiver(name string, config ReceiverConfig) (*httpReceiver, error) {
	h := &httpReceiver{
		Address: defaultHTTPReceiverAddress,
		name:    name,
		storage: config.GetCommonStorage().(Storage),
		log:     config.GetLog().(Log),
	}
	if err := json.Unmarshal(config.GetSettings(name), h); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *httpReceiver) Start() {
	h.log.Write("INFO", "[%s] Starting listenig at: '%s'\n", h.name, h.Address)
	go h.start()
}

func (h *httpReceiver) start() {
	http.HandleFunc("/", h.httpHander)
	if err := http.ListenAndServe(h.Address, nil); err != nil {
		h.log.Write("FATAL", "Failed to listen http receiver: %s\n", err.Error())
	}
}

func (h *httpReceiver) httpHander(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.storage.Add(body)

	body = nil

}
