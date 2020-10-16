package main

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"os"
	"sync"
	"time"
)

func playBeep() {

	file, err := os.Open("../censor-beep-01.wav")
	if err != nil {
		fmt.Println("playBeep-Open", err)
		return
	}

	streamer, format, err := wav.Decode(file)
	if err != nil {
		fmt.Println("playBeep-Decode", err)
		return
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		fmt.Println("playBeep-Init", err)
		return
	}

	done := sync.WaitGroup{}
	done.Add(1)

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done.Done()
	})))

	done.Wait()
}
