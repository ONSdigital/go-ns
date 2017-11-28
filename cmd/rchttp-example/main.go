package main

import (
	"context"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
)

type gettit struct {
	resp *http.Response
	err  error
}

func main() {
	url := flag.String("url", "http://localhost:8081", "URL to access")
	method := flag.String("method", "GET", "decide on GET, POST, PUT")
	data := flag.String("data", "this is the payload", "data to send (if POST, PUT)")
	contentType := flag.String("mime", "text/plain", "content type (header value)")
	flag.Parse()
	*method = strings.ToUpper(*method)

	rcClient := rchttp.DefaultClient
	rcClient.MaxRetries = 4
	rcClient.HTTPClient.Timeout = 3 * time.Second

	clientCompleteChan := make(chan *gettit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	payloadReader := strings.NewReader(*data)

	// kick off a client
	go func() {
		var (
			resp *http.Response
			err  error
		)
		switch *method {
		case "GET":
			resp, err = rcClient.Get(ctx, *url)
		case "POST":
			resp, err = rcClient.Post(ctx, *url, *contentType, payloadReader)
		case "PUT":
			resp, err = rcClient.Put(ctx, *url, *contentType, payloadReader)
		default:
			panic("Bad method")
		}
		clientCompleteChan <- &gettit{resp, err}
	}()

	var respErr *gettit
	select {
	case respErr = <-clientCompleteChan:
		if respErr == nil {
			panic("got nil - chan closed?")
		}
		if respErr.err != nil {
			panic(respErr.err)
		}
	case <-time.After(6 * time.Second):
		log.Debug("Cancelling after timer", nil)
		cancel()
	}

	// if we have a response, show it
	if respErr != nil {
		defer respErr.resp.Body.Close()
		body, err := ioutil.ReadAll(respErr.resp.Body)
		if err != nil {
			panic(err)
			return
		}
		log.Debug("bod", log.Data{"body": string(body), "status": respErr.resp.Status})
	}

	log.Debug("that's all folks", nil)
}
