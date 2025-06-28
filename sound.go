package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	log "github.com/sirupsen/logrus"
)

// SoundType represents different types of sounds that can be played
type SoundType string

const (
	FocusStart  SoundType = "focus_start"
	FocusEnd    SoundType = "focus_end"
	BreakStart  SoundType = "break_start"
	BreakEnd    SoundType = "break_end"
	BreakCancel SoundType = "break_cancel"
	FocusCancel SoundType = "focus_cancel"
)

const soundsDir = "$HOME/.pomo/sounds"
const sampleRate = beep.SampleRate(48000)

// SoundConfig holds configuration for different sound events
type SoundConfig struct {
	FocusStart   string `mapstructure:"focus-start"`
	FocusEnd     string `mapstructure:"focus-end"`
	BreakStart   string `mapstructure:"break-start"`
	BreakEnd     string `mapstructure:"break-end"`
	BreakCancel  string `mapstructure:"break-cancel"`
	FocusCancel  string `mapstructure:"focus-cancel"`
	SoundEnabled bool   `mapstructure:"sound-enabled"`
}

type Sound struct {
	path   string
	buffer *beep.Buffer
}

// SoundManager manages audio playback
type SoundManager struct {
	config SoundConfig
	sounds map[SoundType]Sound // Cache of local paths for each sound type
}

// Cleanup releases audio resources
func (sm *SoundManager) Cleanup() {
	log.Debug("cleaning up audio resources")
	speaker.Close()
}

// IsEnabled returns whether sound playback is enabled
func (sm *SoundManager) IsEnabled() bool {
	return sm.config.SoundEnabled
}

func NewSoundManager(config SoundConfig) *SoundManager {
	sm := &SoundManager{
		config: config,
		sounds: make(map[SoundType]Sound),
	}

	return sm
}

func (sm *SoundManager) Init() error {
	if err := speaker.Init(sampleRate, sampleRate.N(time.Second/10)); err != nil {
		return fmt.Errorf("initialize speaker: %w", err)
	}

	// Download all configured sounds
	soundsToDownload := map[SoundType]string{
		FocusStart:  sm.config.FocusStart,
		FocusEnd:    sm.config.FocusEnd,
		BreakStart:  sm.config.BreakStart,
		BreakEnd:    sm.config.BreakEnd,
		BreakCancel: sm.config.BreakCancel,
		FocusCancel: sm.config.FocusCancel,
	}

	group := sync.WaitGroup{}
	for soundType, soundPath := range soundsToDownload {
		if soundPath != "" {
			group.Add(1)
			go func(soundType SoundType, soundPath string) {
				log.Debugf("loading sound %s: %s", soundType, soundPath)
				localPath, err := cacheSound(soundPath)

				if err == nil {
					sound, err := loadSound(localPath, sampleRate)
					if err == nil {
						sm.sounds[soundType] = *sound
					} else {
						log.Warnf("failed to load sound %s: %v", soundType, err)
					}
				} else {
					log.Warnf("failed to load sound %s: %v", soundType, err)
				}
				group.Done()
			}(soundType, soundPath)
		}
	}
	group.Wait()
	return nil
}

func playBell() {
	fmt.Print("\a\r")
}

func loadSound(path string, speakerSampleRate beep.SampleRate) (*Sound, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open sound file %s: %w", path, err)
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)

	if err != nil {
		return nil, fmt.Errorf("decode MP3 %s: %w", path, err)
	}
	defer streamer.Close()

	resampled := beep.Resample(4, speakerSampleRate, format.SampleRate, streamer)
	buffer := beep.NewBuffer(format)
	buffer.Append(resampled)
	log.Debugf("loaded sound %s", path)
	return &Sound{
		path:   path,
		buffer: buffer,
	}, nil
}

func (sm *SoundManager) PlaySound(soundType SoundType) chan bool {
	// If sound is disabled, return without playing anything
	if !sm.IsEnabled() {
		log.Debug("sound is disabled")
		return nil
	}

	speaker.Clear()

	if sound, ok := sm.sounds[soundType]; ok {
		log.Debugf("playing sound %s: %v", soundType, sound.path)
		done := make(chan bool)
		audio := sound.buffer.Streamer(0, sound.buffer.Len())
		speaker.Play(beep.Seq(audio, beep.Callback(func() {
			log.Debugf("sound playback %s complete: %v", soundType, sound.path)
			done <- true
		})))
		return done
	}

	// If sound type is not initialized, play terminal bell
	log.Debugf("sound type %s is not initialized, using terminal bell", soundType)
	playBell()
	return nil
}

func cacheSound(url string) (string, error) {
	if !strings.HasPrefix(url, "http") {
		log.Debugf("using local sound file: %s", url)
		return url, nil
	}

	// Create sounds directory if it doesn't exist
	soundsDirPath := os.ExpandEnv(soundsDir)
	if err := os.MkdirAll(soundsDirPath, 0755); err != nil {
		return "", fmt.Errorf("create sounds directory %s: %w", soundsDirPath, err)
	}

	// Extract filename from URL
	filename := filepath.Base(url)
	localPath := filepath.Join(soundsDirPath, filename)

	// Check if file already exists
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download sound %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download sound %s: HTTP status %d", url, resp.StatusCode)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("create local file %s: %w", localPath, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("write sound file %s: %w", localPath, err)
	} else {
		log.Infof("successfully downloaded sound file to: %s", localPath)
	}

	return localPath, nil
}
