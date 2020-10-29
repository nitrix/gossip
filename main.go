package main

import "log"

func main() {
	config := Config{}
	err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}

	audio := Audio{}
	go audio.Play()

	twitch := NewTwitch(config, &audio)
	twitch.Run()
}