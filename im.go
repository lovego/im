package im

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/lovego/errs"
	"github.com/lovego/logger"
)

type IM struct {
	redisUrl string
	channel  string

	poolForPublish *redis.Pool
	//               system     user
	pullRequests map[string]map[string][]pullRequest
	sync.RWMutex
}

func New(redisUrl, channel string, poolForPublish *redis.Pool, log *logger.Logger) *IM {
	im := &IM{
		redisUrl:       redisUrl,
		channel:        channel,
		poolForPublish: poolForPublish,
		pullRequests:   make(map[string]map[string][]pullRequest),
	}
	im.setupPoolForPublish()

	go im.loop(log)
	return im
}

func (im *IM) loop(log *logger.Logger) {
	subscribeConn, err := im.getSubscribeConn()
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		switch v := subscribeConn.Receive().(type) {
		case redis.Message:
			msg := message{}
			if err := json.Unmarshal(v.Data, &msg); err != nil {
				log.Error(errs.Trace(err))
			} else {
				im.feedPullRequests(msg)
			}
		case error:
			subscribeConn.Close()
			for {
				subscribeConn, err = im.getSubscribeConn()
				if err != nil {
					log.Error(err)
					time.Sleep(time.Minute)
					continue
				}
				break
			}
		}
	}
}
