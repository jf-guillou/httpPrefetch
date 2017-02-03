package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type response struct {
	Res   string `json:"res"`
	State int    `json:"state"`
}

const (
	stateErr             = -1
	stateWaitPreloader   = 1
	statePreloading      = 2
	statePreloadingQueue = 3
	stateOk              = 4
	stateNoContent       = 5
	stateHTTPFail        = 6
)

var errEmptyBody = errors.New("Empty body")

var logger = log.New(os.Stdout, "", log.LstdFlags)

func main() {
	localAddress := flag.String("address", "127.0.0.1", "Local listen address")
	localPort := flag.Int("port", 8089, "Local listen port")

	flag.Parse()

	listen := fmt.Sprintf("%s:%d", *localAddress, *localPort)

	logger.Println("Hello")

	logger.Printf("Listenning at %s\n", listen)
	http.HandleFunc("/pf", pf)
	log.Fatal(http.ListenAndServe(listen, nil))

	logger.Println("Bye")
}

func pf(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("res")
	if url == "" {
		resp(w, url, stateErr)
		return
	}

	logger.Printf("Request %s", url)

	err := fetch(url)
	if err != nil {
		if err == errEmptyBody {
			resp(w, url, stateNoContent)
		} else {
			resp(w, url, stateHTTPFail)
		}
		return
	}

	resp(w, url, stateOk)
}

func resp(w http.ResponseWriter, url string, state int) error {
	headers := w.Header()
	headers.Set("Content-Type", "application/json; charset=utf-8")
	headers.Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	return enc.Encode(response{
		Res:   url,
		State: state,
	})
}

func fetch(url string) error {
	logger.Printf("Fetching %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.ContentLength == 0 {
		return errEmptyBody
	}

	logger.Printf("Fetched %d from %s\n", resp.ContentLength, url)
	return nil
}
