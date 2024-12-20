package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Response map[string]interface{}

func SendErrMessage(w http.ResponseWriter, r *http.Request, text string, code int) {
	response := Response{
		"error": Response{
			"timestamp": time.Now(),
			"message":   text,
			"path":      r.URL.Path,
			"method":    r.Method,
		},
	}

	dataResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("marshal response with error; error: [%s]; path: [%s]\n; method: [%s]; error message: [%s]\n", err.Error(), r.URL.Path, r.Method, text)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(dataResponse)
}

func SendResponse(w http.ResponseWriter, r *http.Request, response Response) {
	dataResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("marshal response error: [%s]; path: [%s]\n; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(dataResponse)
}

func ReadBody(w http.ResponseWriter, r *http.Request) []byte {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read body error: [%s]; path: [%s]; method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	return body
}
