package im

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/lovego/errs"
)

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
	err = psc.Subscribe(im.channel)
	if err != nil {
		psc.Close()
		return nil, errs.Trace(err)
	}
	return &psc, nil
}

func (im *IM) setupPoolForPublish() {
	if im.poolForPublish != nil {
		return
	}
	im.poolForPublish = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(
				dbUrl,
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
