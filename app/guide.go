package app

import (
	"fmt"

	"github.com/hsk-kr/licokit/lib/styles"
)

func Guide() {
	bullets := []string{
		"Before erasing, run: licokit doctor --reset. It never edits projects and blocks on dirty or unpushed Git work.",
		"After reinstalling macOS, run: licokit install --profile personal.",
		"Use licokit update --profile personal to update Homebrew, App Store apps, standalone CLIs, dotfiles, runtimes, and plugins.",
		"LicoKit preserves replaced config files under ~/.local/state/licokit/backups instead of deleting them.",
		"macOS-owned dialogs, app sign-ins, SSH credentials, and Privacy & Security permissions still require confirmation.",
		"When launching Karabiner Elements for the first time, approve its required macOS permissions before relying on mappings.",
	}

	for _, b := range bullets {
		fmt.Printf(" %s %s\n",
			styles.GuideBullet.Render("●"),
			styles.GuideText.Render(b),
		)
	}
}
