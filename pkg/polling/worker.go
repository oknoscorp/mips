package polling

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/xasmirx/mips/pkg/command"
	"github.com/xasmirx/mips/pkg/helpers"
)

type Response struct {
	Request  *Request
	Error    error
	Response string
}

type Request struct {
	Command      string     `json:"-"`
	RequestAt    *time.Time `json:"-"`
	ResponseAt   *time.Time `json:"-"`
	Response     string     `json:"-"`
	Pending      bool       `json:"-"`
	Push         bool       `json:"-"`
	Error        error      `json:"-"`
	TimeoutCount int        `json:"-"`
}

type Worker struct {
	IP    string `json:"IP"`
	Valid bool   `json:"-"`

	Terminate chan bool

	Requests []*Request
	Buffer   chan *Response

	LoadedAt *time.Time
}

func (worker *Worker) terminate() {
	worker.Terminate <- true
}

func (worker *Worker) terminated() bool {
	select {
	case <-worker.Terminate:
		return true
	default:
	}
	return false
}

// poller listens for messages inside worker instance.
// It creates buffered channel and waits for all messages to arrive.
// For example we wait for stats and pools for Bitmain machines, and
// then we merge everything to a single map and save as JSON.
func (worker *Worker) poller(polling *Polling) {

	//if worker is already loaded then return
	if worker.LoadedAt != nil {
		return
	}

	worker.LoadedAt = helpers.TimeNow()

	for {

		if worker.terminated() {
			return
		}

		worker.Buffer = make(chan *Response, len(worker.Requests))
		for _, req := range worker.Requests {
			go worker.request(req, helpers.TimeNow())
		}

		// read all buffers
		for i := 0; i < len(worker.Requests); i++ {
			r := <-worker.Buffer
			r.Request.Error = r.Error
			r.Request.Response = r.Response
		}

		polling.ChannelWorker <- worker
		time.Sleep(time.Duration(timeLimitSeconds) * time.Second)
	}
}

// request sends API TCP request to worker and pushed data
// to a channel.
func (worker *Worker) request(req *Request, t *time.Time) {

	cmd := command.Command{
		IP:      worker.IP,
		Command: req.Command,
	}

	response, err := command.New(cmd).Execute()
	// if no error is returned, we send buffer to channel immediately
	if err == nil {
		worker.Buffer <- &Response{
			Request:  req,
			Error:    nil,
			Response: response,
		}
		return
	}

	log.Error(err)
	if strings.Contains(err.Error(), "timeout") {
		if req.TimeoutCount >= maxTimeoutRepeats {
			req.TimeoutCount = 0
			worker.Buffer <- &Response{
				Request: req,
				Error:   err,
			}
		} else {
			req.TimeoutCount += 1
			time.Sleep(3 * time.Second)
			worker.request(req, helpers.TimeNow())
		}
	} else {
		worker.Buffer <- &Response{
			Request: req,
			Error:   err,
		}
	}
}
