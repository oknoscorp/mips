package polling

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oknoscorp/mips/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

const postbackQueueDir = "postback_queue"
const maxRequestsPerSecond = 100
const failedRequestRepeatAfter = 120

type PostbackEndpoint struct {
	Polling *Polling `json:"-"`
	URL     string   `json:"url"`
	Data    PushItem `json:"data"`
}

type PostbackRequest struct {
	Endpoints []PostbackEndpoint
}

type PostbackQueue struct {
	File        string
	StartedAt   *time.Time
	CompletedAt *time.Time
	LatestTryAt *time.Time
	LatestError error
}

func (pq *PostbackQueue) deleteFile() {
	err := os.Remove(helpers.Path(postbackQueueDir + "/" + pq.File))
	if err != nil {
		log.Error(err)
	}
}

func (pq *PostbackQueue) request(buffer chan *PostbackQueue) {
	var err error
	var req *http.Request
	var resp *http.Response

	defer func() {
		pq.LatestError = err
		buffer <- pq
	}()

	pq.LatestTryAt = helpers.TimeNow()
	content, err := os.ReadFile(helpers.Path(postbackQueueDir + "/" + pq.File))

	// if postback file cannot be found
	// we will signal for deletion
	if err != nil {
		pq.CompletedAt = helpers.TimeNow()
		log.Error(err)
		return
	}

	decode := PostbackEndpoint{}
	err = json.Unmarshal(content, &decode)

	// if postback file cannot be decoded
	// we will signal for deletion
	if err != nil {
		pq.CompletedAt = helpers.TimeNow()
		log.Error(err)
		return
	}

	req, err = http.NewRequest(http.MethodPost, decode.URL, bytes.NewBuffer(content))
	if err != nil {
		log.Error(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Error(err)
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("postback response status code was %v", resp.StatusCode))
		log.Error(err)
		return
	}

	pq.CompletedAt = helpers.TimeNow()

	defer resp.Body.Close()
}

// PushQueue will push new item to postback queue.
func (polling *Polling) PushQueue(p *PostbackQueue) {
	polling.PostbackQueue.Store(p.File, p)
}

// DeleteQueue will remove item from postback queue.
func (polling *Polling) DeleteQueue(p *PostbackQueue) {
	polling.PostbackQueue.Delete(p.File)
}

// Push will create new queue file and push worker info to it. File will be
// deleted once postback request is sent.
func (pr *PostbackRequest) Push() {
	var wg sync.WaitGroup
	for _, ep := range pr.Endpoints {
		wg.Add(1)
		go func(wg *sync.WaitGroup, ep *PostbackEndpoint) {
			defer wg.Done()
			uuid := uuid.New()
			file := uuid.String() + ".json"
			f, err := os.OpenFile(helpers.Path(postbackQueueDir, "/", file), os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				log.Error("cannot open postback file", err)
				return
			}
			defer f.Close()

			m, err := json.Marshal(ep)
			if err != nil {
				log.Error("cannot decode postback file", err)
				return
			}

			f.WriteString(string(m))

			defer func() {
				ep.Polling.PushQueue(&PostbackQueue{
					File:        file,
					StartedAt:   nil,
					CompletedAt: nil,
					LatestTryAt: nil,
				})
			}()

		}(&wg, &ep)
	}
	wg.Wait()
}

// pbSave will push income worker info to files. Files are then
// red from queue list and posted back to defined endpoint URLs.
func (polling *Polling) pbSave() {
	for {
		(<-polling.ChannelPostbackRequest).Push()
	}
}

// pbProcess if infinite loop function
// it handles push requests to other servers.
func (polling *Polling) pbProcess() {
	for {
		items := []*PostbackQueue{}
		counter := 0
		polling.PostbackQueue.Range(func(k, v any) bool {
			if counter > maxRequestsPerSecond {
				return false
			}

			if v.(*PostbackQueue).LatestTryAt != nil &&
				time.Since(*v.(*PostbackQueue).LatestTryAt).Seconds() < failedRequestRepeatAfter {
				return true
			}

			items = append(items, v.(*PostbackQueue))
			counter++
			return true
		})

		lenItems := len(items)

		buffer := make(chan *PostbackQueue, lenItems)
		for _, item := range items {
			go item.request(buffer)
		}

		fmt.Println("total items in queue", lenItems)
		for i := 0; i < len(items); i++ {
			if r := <-buffer; r.CompletedAt != nil {
				r.deleteFile()
				polling.PostbackQueue.Delete(r.File)
			}
		}

		time.Sleep(3 * time.Second)
	}
}

// RecoverQueue will read unprocessed files from postback_queue dir and it
// will pushem them to PostbackQueue list.
func (polling *Polling) recoverQueue() {
	d, e := os.ReadDir(helpers.Path(postbackQueueDir))
	if e != nil {
		log.Error(e)
		return
	}

	for _, f := range d {
		polling.PushQueue(&PostbackQueue{
			File:        f.Name(),
			StartedAt:   nil,
			CompletedAt: nil,
			LatestTryAt: nil,
		})
	}
}
