package app

import (
	"fmt"

	"github.com/hsk-kr/licokit/lib/display"
	"github.com/hsk-kr/licokit/lib/styles"
	"github.com/hsk-kr/licokit/lib/terminal"
	"github.com/hsk-kr/licokit/lib/tools"
)

// CPUKiller is the screen for the high-CPU watchdog: a LaunchAgent that kills
// any of the user's processes which stays above 90% CPU for ~10 minutes.
func CPUKiller() {
	display.DisplayHeader(true)
	fmt.Println(styles.SectionTitle.Render("High CPU Killer"))
	fmt.Println(styles.GuideText.Render("Kills any of your processes that stays above 90% CPU for ~10 minutes (checked every 30s, 20 strikes)."))

	items := []terminal.SelectItem{
		{Name: "Enable (start now + on startup)"},
		{Name: "Disable"},
		{Name: "Show paths & commands"},
	}

	for {
		if tools.CPUKillerEnabled() {
			fmt.Println(styles.StatusInstalled.Render("\n● status: enabled — running now and on every login"))
		} else {
			fmt.Println(styles.StatusNotInstalled.Render("\n○ status: disabled"))
		}

		choice, err := terminal.Select(items)
		if err != nil {
			return
		}

		switch choice {
		case "Enable (start now + on startup)":
			if err := tools.EnableCPUKiller(); err != nil {
				tools.WarningMessage(err.Error())
			} else {
				tools.SuccessMessage("High CPU killer enabled.\n\n" +
					"• Running now and on every login (LaunchAgent)\n" +
					"• Kills a process after >90% CPU for ~10 min, then notifies\n" +
					"• Log:    tail -f /tmp/cpu-killer.log\n" +
					"• Tune:   ~/.config/cpu-killer/config (applies within 30s)")
			}
		case "Disable":
			if err := tools.DisableCPUKiller(); err != nil {
				tools.WarningMessage(err.Error())
			} else {
				tools.SuccessMessage("High CPU killer disabled and removed from startup.")
			}
		case "Show paths & commands":
			fmt.Println(styles.GuideText.Render(
				" script:  ~/scripts/cpu-killer.sh (repo: ~/licokit/dotfiles/scripts/cpu-killer.sh)\n" +
					" agent:   ~/Library/LaunchAgents/com.lico.cpu-killer.plist\n" +
					" config:  ~/.config/cpu-killer/config\n" +
					" log:     /tmp/cpu-killer.log\n" +
					" watch:   tail -f /tmp/cpu-killer.log"))
		default:
			NotSupported(choice)
		}
	}
}
