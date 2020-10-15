package utility

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"os"
	"sync"
	"time"
)

func Play() {

	file, err := os.Open("../censor-beep-01.wav")
	if err != nil {
		fmt.Println("Play-Open", err)
		return
	}

	streamer, format, err := wav.Decode(file)
	if err != nil {
		fmt.Println("Play-Decode", err)
		return
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		fmt.Println("Play-Init", err)
		return
	}

	done := sync.WaitGroup{}
	done.Add(1)

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done.Done()
	})))

	done.Wait()
}
