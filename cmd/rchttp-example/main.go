package main

import (
	"context"
	// "errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
)

type gettit struct {
	resp *http.Response
	err  error
}

func main() {
	rcClient := rchttp.DefaultClient
	rcClient.MaxRetries = 4
	clientCompleteChan := make(chan *gettit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// kick off a client GET
	go func() {
		resp, err := rcClient.Get(ctx, "http://www.ons.gov.uk:8081")
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
