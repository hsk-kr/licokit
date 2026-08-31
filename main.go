package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hsk-kr/licokit/app"
	"github.com/hsk-kr/licokit/lib/config"
	"github.com/hsk-kr/licokit/lib/terminal"
	"github.com/hsk-kr/licokit/lib/tools"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	tools.RefreshHomebrewEnvironment()

	if len(args) > 0 {
		switch args[0] {
		case "install":
			return runInstall(cfg, args[1:])
		case "update":
			return runUpdate(cfg, args[1:])
		case "doctor":
			return runDoctor(cfg, args[1:])
		case "dotfiles":
			if len(args) != 1 {
				return fmt.Errorf("dotfiles does not accept arguments")
			}
			return tools.SetupDotfiles(cfg.Dotfiles)
		case "version", "--version", "-v":
			fmt.Println("licokit " + version)
			return nil
		case "help", "--help", "-h":
			printUsage()
			return nil
		default:
			printUsage()
			return fmt.Errorf("unknown command %q", args[0])
		}
	}

	terminal.HideCursor()
	defer terminal.ShowCursor()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		terminal.ShowCursor()
		os.Exit(0)
	}()

	app.Home(cfg)
	return nil
}

func runInstall(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	profile := flags.String("profile", "personal", "software profile: core, personal, or all")
	dryRun := flags.Bool("dry-run", false, "show planned work without changing the machine")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("install does not accept positional arguments")
	}
	return tools.InstallProfile(cfg, *profile, *dryRun)
}

func runUpdate(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	profile := flags.String("profile", "personal", "software profile: core, personal, or all")
	dryRun := flags.Bool("dry-run", false, "show planned work without changing the machine")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("update does not accept positional arguments")
	}
	return tools.UpdateProfile(cfg, *profile, *dryRun)
}

func runDoctor(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	profile := flags.String("profile", "personal", "software profile: core, personal, or all")
	reset := flags.Bool("reset", false, "fail if project repositories contain local-only work")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("doctor does not accept positional arguments")
	}
	return tools.Doctor(cfg, *profile, *reset)
}

func printUsage() {
	fmt.Print(`LicoKit — reproducible macOS development setup

Usage:
  licokit                         Open the interactive menu
  licokit install [--profile personal|core|all] [--dry-run]
  licokit update  [--profile personal|core|all] [--dry-run]
  licokit doctor  [--profile personal|core|all] [--reset]
  licokit dotfiles
  licokit version
`)
}
