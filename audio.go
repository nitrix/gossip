package main

import (
	"bytes"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"io"
	"net/http"
	"sync"
	"time"
)


type Audio struct{
	mixer beep.Mixer
	mutex sync.Mutex
}

func (a *Audio) Play() {
	sampleRate := beep.SampleRate(22050)

	err := speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	if err != nil {
		return
	}

	speaker.Play(&a.mixer)
}

func (a *Audio) Queue(data *bytes.Buffer) error {
	if data.Len() == 0 {
		return nil
	}

	streamer, _, err := wav.Decode(data)
	if err != nil {
		return err
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()

	waited := false
	for a.mixer.Len() > 0 {
		time.Sleep(100 * time.Millisecond)
		waited = true
	}

	if waited {
		time.Sleep(1 * time.Second)
	}

	a.mixer.Add(streamer)

	return nil
}

func synthesize(msg string, buffer *bytes.Buffer) error {
	reader := bytes.NewReader([]byte(msg))

	response, err := http.Post("https://justin.nitrix.me/synthesize", "plain/text", reader)
	if err != nil {
		return fmt.Errorf("unable to synthesize message: %w", err)
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("invalid response code: %d", response.StatusCode)
	}

	_, err = io.Copy(buffer, response.Body)
	if err != nil {
		return fmt.Errorf("unable to copy response body: %w", err)
	}

	return nil
}