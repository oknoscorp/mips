package polling

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/oknoscorp/mips/pkg/helpers"
	"github.com/oknoscorp/mips/pkg/server"
	log "github.com/sirupsen/logrus"
)

// maxTimeoutRepeats defines how many Timeout repeats can be executed
// consequently before poller pushes error for the miner.
var maxTimeoutRepeats int = 10

// timeLimitSeconds defines amount of seconds that must pass before
// new API request can be made.
// Recommended: 180.
var timeLimitSeconds float64 = 180

// loadWorkerInterval defines after how many seconds worker lists will
// be checked.
// Recommended: 300.
var loadWorkersInterval float64 = 300

// Polling struct holds current poller service status.
type Polling struct {
	Server *server.Server

	Workers sync.Map

	ChannelWorker          chan *Worker
	ChannelPostbackRequest chan *PostbackRequest

	PostbackQueue sync.Map
}

// PushItem holds the data that will be pushed to remote postback URL
type PushItem struct {
	IP        string                 `json:"ip"`
	Responses map[string]interface{} `json:"responses"`
	Errors    map[string]interface{} `json:"errors"`
	Time      time.Time              `json:"time"`
}

// New will initialize polling component.
// It also receives TCP connection from server passed
// by reference.
func New(server *server.Server) *Polling {
	new := &Polling{
		ChannelWorker:          make(chan *Worker),
		ChannelPostbackRequest: make(chan *PostbackRequest),
		Server:                 server,
		Workers:                sync.Map{},
	}

	go new.Poll()

	return new
}

// loadWorkers will load all workers from various endpoints.
// All workers have valid initial state, later this state
// can be updated if particular conditions are met.
func (polling *Polling) loadWorkers() {
	for {
		polling.loadWorkerLists()
		polling.Workers.Range(func(key, value any) bool {
			if value.(*Worker).LoadedAt == nil {
				go value.(*Worker).poller(polling)
			}
			return true
		})
		time.Sleep(time.Duration(loadWorkersInterval) * time.Second)
	}
}

// invalidTime checks if time since is less than timeLimitSeconds.
func (polling *Polling) invalidTime(t *time.Time) bool {
	if t == nil {
		return false
	}
	return time.Since(*t).Seconds() < timeLimitSeconds
}

// Reader will receive new Worker data.
func (polling *Polling) reader() {
	for {
		read := <-polling.ChannelWorker
		pollList := polling.Server.ConfigPostbacks()
		endpoints := []PostbackEndpoint{}

		responses := make(map[string]interface{})
		errors := make(map[string]interface{})

		for _, req := range read.Requests {
			responses[req.Command] = helpers.CreateMapFromPipeChar(req.Response)
			if req.Error != nil {
				errors[req.Command] = map[string]interface{}{
					"error":        req.Error,
					"error_string": req.Error.Error(),
				}
			}
		}

		for _, url := range pollList {
			endpoints = append(endpoints, PostbackEndpoint{
				Polling: polling,
				URL:     url,
				Data: PushItem{
					IP:        read.IP,
					Responses: responses,
					Errors:    errors,
					Time:      time.Now(),
				},
			})
		}

		polling.ChannelPostbackRequest <- &PostbackRequest{
			Endpoints: endpoints,
		}
	}
}

// loadWorkerLists will iterate over links defined in configuration.
// Each line represents a single IP address. IP addresses are validated,
// if not valid IP address will not be pushed to polling service.
func (polling *Polling) loadWorkerLists() {
	wg := sync.WaitGroup{}
	list := sync.Map{}
	for _, url := range polling.Server.ConfigLists() {
		wg.Add(1)
		go func(wg *sync.WaitGroup, url string) {
			defer wg.Done()
			request, err := http.Get(url)
			if err != nil {
				log.Error(err)
				return
			}

			if request.StatusCode != 200 {
				log.Error("url %s not reachable", url)
				return
			}

			body, berr := ioutil.ReadAll(request.Body)
			if berr != nil {
				log.Error(berr)
				return
			}

			split := strings.Split(string(body), "\n")
			for k, v := range split {
				v = strings.TrimSpace(v)
				parse := net.ParseIP(v)
				if parse == nil {
					log.Errorf("ip %s is not valid - loaded from list: %s line: %d", string([]byte(v)[0:30]), url, k+1)
					continue
				}

				list.Store(v, v)
			}
		}(&wg, url)
	}

	wg.Wait()

	// check if worker is in poll, if not push it
	list.Range(func(key, value any) bool {
		if _, ok := polling.Workers.Load(key.(string)); !ok {
			polling.Workers.Store(key.(string), &Worker{
				IP:        value.(string),
				Terminate: make(chan bool),
				Valid:     true,
				Requests: []*Request{
					{
						Command: "stats",
						Push:    true,
					},
					{
						Command: "pools",
						Push:    true,
					},
				},
			})
		}
		return true
	})

	// check if worker is in the list, if not remove it
	// also signal by chanel for termination
	polling.Workers.Range(func(key, value any) bool {
		if _, ok := list.Load(key.(string)); !ok {
			value.(*Worker).terminate()
			polling.Workers.Delete(key.(string))
		}
		return true
	})

	polling.Workers.Range(func(key, value any) bool {
		return true
	})
}

// Poll is main function that will execute polling. It simply executes
// infinite iterator and pushes various messages related to poll status.
func (polling *Polling) Poll() {
	polling.recoverQueue()
	go polling.loadWorkers()
	go polling.reader()
	go polling.pbSave()
	go polling.pbProcess()
}
