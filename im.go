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
	redisUrl     string
	redisChannel string
	redisPool    *redis.Pool

	//               system =>  user
	pullRequests map[string]map[string][]*pullRequest
	mtx          sync.RWMutex
}

func New(redisUrl, redisChannel string, redisPool *redis.Pool, log *logger.Logger) *IM {
	im := &IM{
		redisUrl:     redisUrl,
		redisChannel: redisChannel,
		redisPool:    redisPool,
		pullRequests: make(map[string]map[string][]*pullRequest),
	}
	im.setupRedisPool()

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

func (im *IM) feedPullRequests(msg message) {
	im.mtx.RLock()
	defer im.mtx.RUnlock()

	systemReqs := im.pullRequests[msg.System]
	if len(systemReqs) == 0 {
		return
	}
	for user, msgVersion := range msg.UsersVersion {
		for _, req := range systemReqs[user] {
			if reqVersion, ok := req.versions[msg.Business]; ok && reqVersion != msgVersion {
				select {
				case req.ch <- map[string]string{msg.Business: msgVersion}:
				default:
				}
			}
		}
	}
}

func (im *IM) getSubscribeConn() (*redis.PubSubConn, error) {
	c, err := redis.DialURL(
		im.redisUrl,
		redis.DialConnectTimeout(3*time.Second),
		redis.DialWriteTimeout(3*time.Second),
	)
	if err != nil {
		return nil, errs.Trace(err)
	}

	psc := redis.PubSubConn{Conn: c}
	err = psc.Subscribe(im.redisChannel)
	if err != nil {
		psc.Close()
		return nil, errs.Trace(err)
	}
	return &psc, nil
}

func (im *IM) setupRedisPool() {
	if im.redisPool != nil {
		return
	}
	im.redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(
				im.redisUrl,
				redis.DialConnectTimeout(3*time.Second),
				redis.DialReadTimeout(3*time.Second),
				redis.DialWriteTimeout(3*time.Second),
			)
		},
		MaxIdle:     32,
		MaxActive:   128,
		IdleTimeout: 600 * time.Second,
	}

}
