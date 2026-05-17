package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

const version = "1.0.0"

// findMusicFiles searches the given directory for supported audio files.
func findMusicFiles(dir string) ([]string, error) {
	supportedExts := map[string]bool{
		".mp3":  true,
		".flac": true,
		".ogg":  true,
		".wav":  true,
		".m4a":  true,
		".aac":  true,
		".opus": true, // added opus support - I use this format a lot
	}

	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if supportedExts[ext] {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// playFile plays a single audio file using mpv.
func playFile(file string, shuffle bool, loop bool) error {
	args := []string{"--no-video"}
	if shuffle {
		args = append(args, "--shuffle")
	}
	if loop {
		args = append(args, "--loop-playlist=inf")
	}
	args = append(args, file)

	cmd := exec.Command("mpv", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// playDirectory plays all music files in a directory using mpv.
func playDirectory(dir string, shuffle bool, loop bool) error {
	files, err := findMusicFiles(dir)
	if err != nil {
		return fmt.Errorf("error scanning directory: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no supported audio files found in: %s", dir)
	}

	args := []string{"--no-video"}
	if shuffle {
		args = append(args, "--shuffle")
	}
	if loop {
		args = append(args, "--loop-playlist=inf")
	}
	args = append(args, files...)

	cmd := exec.Command("mpv", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func main() {
	app := &cli.App{
		Name:    "cliamp",
		Usage:   "A simple CLI music player powered by mpv",
		Version: version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "shuffle",
				Aliases: []string{"s"},
				Usage:   "Shuffle the playlist",
				Value:   true, // default to shuffle on - I always want this
			},
			&cli.BoolFlag{
				Name:    "loop",
				Aliases: []string{"l"},
				Usage:   "Loop the playlist indefinitely",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				// Default to current directory if no argument provided
				return playDirectory(".", c.Bool("shuffle"), c.Bool("loop"))
			}

			target := c.Args().First()
			info, err := os.Stat(target)
			if err != nil {
				return fmt.Errorf("cannot access '%s': %w", target, err)
			}

			if info.IsDir() {
				return playDirectory(target, c.Bool("shuffle"), c.Bool("loop"))
			}
			return playFile(target, c.Bool("shuffle"), c.Bool("loop"))
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
