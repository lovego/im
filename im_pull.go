package im

import "time"

type pullRequest struct {
	ch chan map[string]string
	//          business version
	versions map[string]string
}

func (im *IM) Pull(
	system, user string, versions map[string]string, timeout time.Duration,
) (map[string]string, error) {

	return nil
}

func (im *IM) feedPullRequests(msg message) {
	im.RLock()
	defer im.RUnlock()

	usersRequests := im.pullRequests[msg.System]
	if len(usersRequests) == 0 {
		return
	}

	for _, user := range msg.Users {
		requests := requests[usersRequests]
		if len(requests) > 0 {
		}

		us := onlineUsers.Get(uid)
		for _, u := range us {
			loadVersionsFromDb(u, []string{notification.Service}, false)
		}
	}
}
