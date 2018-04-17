package audit

import (
	"encoding/json"
	"github.com/ONSdigital/go-ns/log"
)

type DummyProducer struct {
	OutputChan chan []byte
	ExitChan   chan bool
}

func (d *DummyProducer) Output() chan []byte {
	return d.OutputChan
}

func (d *DummyProducer) Run() {
	go func() {
		done := false
		for !done {
			select {
			case auditMsg := <-d.OutputChan:
				log.Debug("dummy producer: audit avro message received", nil)
				var e Event
				err := EventSchema.Unmarshal(auditMsg, &e)
				if err != nil {
					log.ErrorC("dummy producer: avro marshal error", err, log.Data{"event": e})
				} else {
					b, _ := json.MarshalIndent(e, "", " ")
					log.Info("dummy producer:", log.Data{"event": string(b)})
				}
			case <-d.ExitChan:
				log.Debug("dummy producer: exiting", nil)
				done = true
				close(d.OutputChan)
				close(d.ExitChan)
			}
		}
	}()
}
