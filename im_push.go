package im

import (
	"encoding/json"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/lovego/errs"
)

type message struct {
	System       string
	Business     string
	UsersVersion map[string]string // user => version
}

func (im *IM) Push(system string, users []string, business string) error {
	if len(im.systems) > 0 {
		if _, ok := im.systems[system]; !ok {
			return errs.New("args-err", "invalid system: "+system)
		}
	}
	if len(im.businesses) > 0 {
		if _, ok := im.businesses[business]; !ok {
			return errs.New("args-err", "invalid business: "+business)
		}
	}
	return im.push(system, users, business)
}

func (im *IM) push(system string, users []string, business string) error {
	if len(users) == 0 {
		return nil
	}
	conn := im.redisPool.Get()
	defer conn.Close()

	usersVersion := make(map[string]string)
	for _, user := range users {
		version, err := redis.Int64(conn.Do("HINCRBY", system+"/"+user, business, 1))
		if err != nil {
			return errs.Trace(err)
		}
		usersVersion[user] = strconv.FormatInt(version, 10)
	}

	b, err := json.Marshal(message{System: system, Business: business, UsersVersion: usersVersion})
	if err != nil {
		return errs.Trace(err)
	}

	if _, err := conn.Do("PUBLISH", im.redisChannel, string(b)); err != nil {
		return errs.Trace(err)
	}

	return nil
}
