package app

import (
	"github.com/hsk-kr/licokit/lib/config"
	"github.com/hsk-kr/licokit/lib/display"
	"github.com/hsk-kr/licokit/lib/terminal"
	"github.com/hsk-kr/licokit/lib/tools"
)

func Home(cfg *config.Config) {
	items := []terminal.SelectItem{{
		Name: "Tools",
	}, {
		Name: "Dotfiles",
	}, {
		Name: "Guide",
	},
	}

	display.DisplayHeader(true)

	for {
		choice, err := terminal.Select(items)

		if err != nil {
			return
		}

		switch choice {
		case "Tools":
			Tools(cfg)
			display.DisplayHeader(true)
			continue
		case "Dotfiles":
			if err := tools.SetupDotfiles(cfg.Dotfiles); err != nil {
				tools.WarningMessage(err.Error())
			} else {
				tools.SuccessMessage("Dotfiles setup complete.\n\n• Dotfiles updated via git pull\n• Symlinks refreshed from ~/licokit/dotfiles\n• To apply zsh changes, run: source ~/.zshrc")
			}
		case "Guide":
			Guide()
		default:
			NotSupported(choice)
		}

		display.DisplayHeader(false)
	}
}
