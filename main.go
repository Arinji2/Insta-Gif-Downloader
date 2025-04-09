package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Config struct {
	URLs     []string `json:"urls"`
	Duration string   `json:"duration"`
	FPS      string   `json:"fps"`
	Width    string   `json:"width"`
	Height   string   `json:"height"`
	Browser  string   `json:"browser"`
}

func main() {
	defer func() {
		fmt.Println("\nPress Enter to exit...")
		fmt.Scanln()
	}()
	data, err := os.ReadFile("data.json")
	if err != nil {
		fmt.Println("❌ Failed to read data.json:", err)
		return
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Println("❌ Failed to parse data.json:", err)
		return
	}

	if len(config.URLs) == 0 {
		fmt.Println("❌ No URLs found in data.json")
		return
	}

	resultsDir := "results"
	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(resultsDir, os.ModePerm); err != nil {
			fmt.Println("❌ Failed to create results folder:", err)
			return
		}
	} else {
		os.RemoveAll(resultsDir)
		if err := os.MkdirAll(resultsDir, os.ModePerm); err != nil {
			fmt.Println("❌ Failed to create results folder:", err)
			return
		}
	}

	width := config.Width
	height := config.Height
	if width == "auto" {
		width = "-1"
	}
	if height == "auto" {
		height = "-1"
	}
	scale := fmt.Sprintf("scale=%s:%s", width, height)

	ytDLP := "yt-dlp"
	ffmpeg := "ffmpeg"
	if runtime.GOOS == "windows" {
		ytDLP = "./yt-dlp.exe"
		ffmpeg = "./ffmpeg.exe"
	}

	for i, url := range config.URLs {
		fmt.Printf("\n📥 [%d/%d] Downloading: %s\n", i+1, len(config.URLs), url)

		videoName := fmt.Sprintf("video_%d.mp4", i)
		gifName := filepath.Join(resultsDir, fmt.Sprintf("video_%d.gif", i))

		cmd := exec.Command(ytDLP, "--quiet", "--no-warnings", "--cookies-from-browser", config.Browser, "-f", "mp4", "-o", videoName, url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Failed to download video: %v\n", err)
			continue
		}

		fmt.Println("🎞️  Converting to GIF...")

		args := []string{"-v", "warning", "-i", videoName, "-vf", scale}
		if config.Duration != "auto" {
			args = append(args, "-t", config.Duration)
		}
		if config.FPS != "auto" {
			args = append(args, "-r", config.FPS)
		}
		args = append(args, gifName)

		convert := exec.Command(ffmpeg, args...)
		convert.Stdout = os.Stdout
		convert.Stderr = os.Stderr
		if err := convert.Run(); err != nil {
			fmt.Printf("❌ Failed to convert video to GIF: %v\n", err)
			continue
		}

		if err := os.Remove(videoName); err != nil {
			fmt.Printf("⚠️  Failed to delete temp video: %v\n", err)
		}

		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("❌ Failed to get current working directory: %v\n", err)
			continue
		}

		gifPath := filepath.Join(wd, gifName)
		info, err := os.Stat(gifPath)
		if err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			fmt.Printf("✅ Saved %s (%.2f MB)\n\n", gifName, sizeMB)
		} else {
			fmt.Printf("❌ Failed to get file size for %s\n", gifName)
		}
	}

	fmt.Println("\n🏁 Done! All GIFs are inside the 'results' folder.")
}
