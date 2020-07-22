package im

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/lovego/errs"
)

type pullRequest struct {
	versions map[string]string // business => version
	ch       chan map[string]string
}

func (im *IM) Pull(
	system, user string, versions map[string]string, timeout time.Duration,
) (map[string]string, error) {
	if len(im.systems) > 0 {
		if _, ok := im.systems[system]; !ok {
			time.Sleep(timeout) // avoid high load when invalid loop calls.
			return nil, errs.New("args-err", "invalid system: "+system)
		}
	}
	if len(versions) == 0 {
		time.Sleep(timeout) // avoid high load when invalid loop calls.
		return nil, nil
	}
	if len(im.businesses) > 0 {
		for business := range versions {
			if _, ok := im.businesses[business]; !ok {
				time.Sleep(timeout) // avoid high load when invalid loop calls.
				return nil, errs.New("args-err", "invalid business: "+business)
			}
		}
	}
	return im.pull(system, user, versions, timeout)
}

func (im *IM) pull(
	system, user string, versions map[string]string, timeout time.Duration,
) (map[string]string, error) {
	req := &pullRequest{versions: versions, ch: make(chan map[string]string, 1)}

	// register the req before load versions from redis, to avoid published message loss.
	im.registerPullRequest(system, user, req)
	defer im.unregisterPullRequest(system, user, req)

	if newVersions, err := im.loadFromRedis(system, user, versions); err != nil {
		return nil, err
	} else if len(newVersions) > 0 {
		return newVersions, nil
	}

	select {
	case newVersions := <-req.ch:
		return newVersions, nil
	case <-time.After(timeout):
		return map[string]string{}, nil
	}
}

func (im *IM) registerPullRequest(system, user string, req *pullRequest) {
	im.mtx.Lock()
	defer im.mtx.Unlock()

	systemReqs, ok := im.pullRequests[system]
	if !ok {
		systemReqs = make(map[string][]*pullRequest)
		im.pullRequests[system] = systemReqs
	}
	systemReqs[user] = append(systemReqs[user], req)
}

func (im *IM) unregisterPullRequest(system, user string, req *pullRequest) {
	im.mtx.Lock()
	defer im.mtx.Unlock()

	systemReqs := im.pullRequests[system]
	if len(systemReqs) == 0 {
		return
	}

	userReqs := systemReqs[user]
	for i, thisReq := range userReqs {
		if thisReq == req {
			systemReqs[user] = append(userReqs[:i], userReqs[i+1:]...)
		}
	}
}

func (im *IM) loadFromRedis(
	system, user string, versions map[string]string,
) (map[string]string, error) {
	conn := im.redisPool.Get()
	defer conn.Close()

	args := []interface{}{system + "/" + user}
	for business := range versions {
		args = append(args, business)
	}

	versionSlice, err := redis.Strings(conn.Do("HMGET", args...))
	if err != nil {
		return nil, errs.Trace(err)
	}

	newVersions := make(map[string]string)
	for i, version := range versionSlice {
		business := args[i+1].(string)
		if version != versions[business] {
			newVersions[business] = version
		}
	}
	return newVersions, nil
}
