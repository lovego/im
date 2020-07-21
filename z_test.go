package im

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lovego/logger"
)

func ExampleIm() {
	im := New("redis://@localhost/0", "im", nil, logger.New(os.Stderr))

	conn := im.redisPool.Get()
	defer conn.Close()
	if _, err := conn.Do("DEL", "system/user"); err != nil {
		log.Fatal(err)
	}

	go testPull(im)

	time.Sleep(time.Millisecond)
	if err := im.Push("system", []string{"user"}, "notify"); err != nil {
		fmt.Println(err)
	}

	time.Sleep(time.Millisecond)
	testPull(im)

	// Output:
	// map[notify:1]
	// map[notify:1]
}

func testPull(im *IM) {
	versions, err := im.Pull("system", "user", map[string]string{
		"notify": "", "chats": "",
	}, 10*time.Millisecond)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(versions)
	}
}
