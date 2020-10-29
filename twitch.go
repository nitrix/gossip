package main

import (
	"bytes"
	"github.com/jrm780/gotirc"
	"log"
	"strings"
	"sync"
	"time"
)

type Twitch struct{
	config Config
	client *gotirc.Client
	tasks chan func()
}

func NewTwitch(config Config, audio *Audio) *Twitch {
	t := Twitch{
		config: config,
		tasks: make(chan func()),
	}

	options := gotirc.Options{
		Host:     "irc.chat.twitch.tv",
		Port:     6667,
		Channels: []string{"#" + config.channelName},
	}

	client := gotirc.NewClient(options)

	client.OnChat(func(channel string, tags map[string]string, msg string) {
		log.Printf("[%s]: %s", tags["display-name"], msg)

		// Strip channel owner at signs.
		// msg = strings.Replace(msg, "@" + Channel + " ", "", 1)
		// msg = strings.Replace(msg, "@" + strings.ToLower(Channel) + " ", "", 1)

		// Some messages can be skipped.
		if t.isSkippable(msg) {
			return
		}

		// We have to use this waitgroup because we're synthesizing the messages concurrently,
		// but we want them to play sequentially and preserve the order they came in.
		// Essentially, this is just to speed up the synthesizing, which can be pretty slow.
		wg := sync.WaitGroup{}
		wg.Add(1)

		// Shared by the two goroutines below.
		data := &bytes.Buffer{}

		// Synthesize the message.
		go func() {
			defer wg.Done()

			err := synthesize(msg, data)
			if err != nil {
				log.Printf("unable to synthesize message: %s\n", msg)
				return
			}
		}()

		// Queue up playing the sound.
		t.tasks <- func() {
			wg.Wait()

			err := audio.Queue(data)
			if err != nil {
				log.Println("unable to gossip")
			}

			time.Sleep(1 * time.Second)
		}
	})

	t.client = client

	return &t
}

func (t *Twitch) sequentiallyProcessTasks() {
	for tasks := range t.tasks {
		tasks()
	}
}

func (t *Twitch) isSkippable(msg string) bool {
	if strings.Contains(msg, "@") {
		return true
	}

	return false
}

func (t *Twitch) Run() {
	go t.sequentiallyProcessTasks()

	err := t.client.Connect(t.config.botUser, t.config.botToken)
	if err != nil {
		log.Println("unable to connect:", err)
	}
}