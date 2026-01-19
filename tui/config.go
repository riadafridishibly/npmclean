package tui

import "time"

type Config struct {
	ReplaceHomeWithTilde bool          `json:"replace_home_with_tilde"`
	ProgressUpdateFreq   time.Duration `json:"progress_update_freq"`
}
