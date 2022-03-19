package main

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"fmt"

	g "github.com/AllenDang/giu"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/ncruces/zenity"
)

var (
	files        = []string{}
	clicked      = map[string]bool{}
	replay       = map[string]chan bool{}
	quit         = map[string]chan bool{}
	aps          = map[string]*audioPanel{}
	initialized  = false
	stride       = 7
	str          = "7"
	isRemoveMode = false
	ks           = []g.Key{
		g.KeyQ, g.KeyW, g.KeyE, g.KeyR, g.KeyT, g.KeyY, g.KeyU, g.KeyI, g.KeyO, g.KeyP,
		g.KeyA, g.KeyS, g.KeyD, g.KeyF, g.KeyG, g.KeyH, g.KeyJ, g.KeyK, g.KeyL,
		g.KeyZ, g.KeyX, g.KeyC, g.KeyV, g.KeyB, g.KeyN, g.KeyM,
	}
)

type audioPanel struct {
	sampleRate beep.SampleRate
	streamer   beep.StreamSeeker
	ctrl       *beep.Ctrl
	resampler  *beep.Resampler
	volume     *effects.Volume
}

func newAudioPanel(song string, sampleRate beep.SampleRate, streamer beep.StreamSeeker) *audioPanel {
	ctrl := &beep.Ctrl{Streamer: beep.Seq(streamer, beep.Callback(func() {
		quit[song] <- true
	}))}
	resampler := beep.ResampleRatio(4, float64(sampleRate.N(time.Second))/float64(beep.SampleRate(44000).N(time.Second)), ctrl)
	volume := &effects.Volume{Streamer: resampler, Base: 2}
	return &audioPanel{sampleRate, streamer, ctrl, resampler, volume}
}

func (ap *audioPanel) play() {
	speaker.Play(ap.volume)
}

func playReplayQuit(song string) {
	if len(quit[song]) > 0 {
		for len(quit[song]) > 0 {
			<-quit[song]
		}
	}

	if aps[song] == nil {
		loadfile(song)
		fmt.Println(aps[song])
	}

	aps[song].play()

	for {
		select {
		case <-replay[song]:
			speaker.Lock()
			err := aps[song].streamer.Seek(0)
			if err != nil {
				fmt.Println(err)
			}
			aps[song].ctrl.Paused = false
			speaker.Unlock()
			continue
		case <-quit[song]:
			speaker.Lock()
			clicked[song] = false
			err := aps[song].streamer.Seek(aps[song].streamer.Len() - 1)
			if err != nil {
				fmt.Println(err)
			}
			speaker.Unlock()
			loadfile(song)
			return
		}
	}
}

func onSongClick(song string) func() {
	if replay[song] == nil {
		replay[song] = make(chan bool, 10)
	}
	if quit[song] == nil {
		quit[song] = make(chan bool, 10)
	}
	return func() {
		if isRemoveMode {
			for i, file := range files {
				if file == song {
					files = append(files[:i], files[i+1:]...)
					saveSetlist()
				}
			}
			return
		}
		if !clicked[song] {
			replay[song] <- true
			go playReplayQuit(song)
			clicked[song] = true
		} else {
			replay[song] <- true
		}
	}
}
func loadfile(song string) error {
	var (
		streamer beep.StreamSeekCloser
		format   beep.Format
	)
	f, err := os.Open(song)
	if err != nil {
		return err
	}
	streamer, format, err = mp3.Decode(f)
	if err != nil {
		return err
	}
	if !initialized {
		err = speaker.Init(beep.SampleRate(44000), 2048)
		if err != nil {
			return err
		}
		initialized = true
	}
	aps[song] = newAudioPanel(song, format.SampleRate, streamer)
	return nil
}

func selectFiles() ([]string, error) {
	return zenity.SelectFileMutiple()
}

func loop() {
	buttons := []g.Widget{}
	for i, file := range files {
		splt := strings.Split(file, "/")
		if len(splt) == 1 {
			splt = strings.Split(file, "\\")
		}
		addon := ""
		if i < len(ks) {
			addon = "[" + string(ks[i]) + "] "
		}
		buttons = append(buttons,
			g.Button(addon+splt[len(splt)-1]).OnClick(onSongClick(file)),
		)
	}
	newStride, err := strconv.Atoi(str)
	if err == nil {
		stride = newStride

	}
	removeStr := "Remove Off"
	if isRemoveMode {
		removeStr = "Remove On"
	}

	rows := []g.Widget{
		g.Row(
			g.InputText(&str).Label("Buttons per line"),
			g.Button("[Esc] Stop").OnClick(func() {
				for k, v := range clicked {
					if v {
						quit[k] <- true
					}
				}
			}),
			g.Button("Add .mp3").OnClick(func() {
				f, err := selectFiles()
				if err != nil {
					fmt.Println(err)
				}
				for i, file := range f {
					if err = loadfile(file); err != nil {
						f = append(f[:i], f[i+1:]...)
					}
				}
				files = append(files, f...)
				saveSetlist()
			}),
			g.Button(removeStr).OnClick(func() {
				isRemoveMode = !isRemoveMode
			}),
		),
	}
	for i := 0; i < len(buttons); i += stride {
		rows = append(rows, g.Row(buttons[i:int(math.Min(float64(i+stride), float64(len(buttons))))]...))
	}
	if len(buttons) == 0 {
		rows = append(rows, g.Row(g.Label("No files found, drag'em in here")))
	}
	g.SingleWindow().Layout(
		rows...,
	)
}

func main() {
	data, err := os.ReadFile("./setlist.json")
	if err == nil {
		err = json.Unmarshal(data, &files)
		if err != nil {
			log.Println(err)
		}
		for i, file := range files {
			err = loadfile(file)
			if err != nil {
				files = append(files[:i], files[i+1:]...)
			}
		}
		saveSetlist()
	}
	wnd := g.NewMasterWindow("Soundboard", 1920*2/3, 1080*2/3, g.MasterWindowFlags(g.WindowFlagsNone))
	wnd.SetDropCallback(func(f []string) {
		files = append(files, f...)
		for i, file := range f {
			err = loadfile(file)
			if err != nil {
				f = append(f[:i], f[i+1:]...)
			}
		}
		files = append(files, f...)
		saveSetlist()
	})
	scuts := []g.WindowShortcut{{
		Key: g.KeyEscape,
		Callback: func() {
			for k, v := range clicked {
				if v {
					quit[k] <- true
				}
			}
		},
	}}
	for i := 0; i < len(files) && i < len(ks); i++ {
		oc := onSongClick(files[i])
		scuts = append(scuts, g.WindowShortcut{
			Key:      ks[i],
			Callback: oc,
		})
	}
	wnd = wnd.RegisterKeyboardShortcuts(scuts...)
	wnd.Run(loop)
}

func saveSetlist() {
	data, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		log.Println(err)
	}
	err = os.WriteFile("./setlist.json", data, 0644)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("setlist saved")
}
