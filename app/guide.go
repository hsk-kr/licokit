package app

import (
	"fmt"

	"github.com/hsk-kr/licokit/lib/styles"
)

func Guide() {
	bullets := []string{
		"Install Tools before setting up Dotfiles. Dotfiles setup runs post-scripts that depend on installed tools.",
		"Change the Homebrew click key to Shift + Command + F.",
		"These commands assume you're using zsh. After installing wezterm, use your default terminal to install other software, then verify the results in wezterm.",
		"When launching Karabiner Elements for the first time, it may reset the configuration, requiring you to reconfigure your dotfiles.",
		"Homebrew should be installed by manually running the provided shell command.",
		"You'll need to install Go and nvm to set up Language Server Protocol (LSP).",
		"Remember to run source commands when needed (e.g., `source ~/.zshrc` or `NVM_PATH`).",
	}

	for _, b := range bullets {
		fmt.Printf(" %s %s\n",
			styles.GuideBullet.Render("●"),
			styles.GuideText.Render(b),
		)
	}
}
