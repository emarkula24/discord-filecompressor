package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func Compress() {
	target_video := 8.3 //megabytes
	target_audio := 1.2 // megabytes
	d := 218.150998
	vb, ab := CalculateBitrates(d, target_video, target_audio)
	input := "/mnt/F6ECB1FBECB1B669/Videot/dvar bemiy.mp4"
	output := "./output.mp4"

	vbStr := strconv.FormatFloat(vb, 'f', 0, 64)
	abStr := strconv.FormatFloat(ab, 'f', 0, 64)

	// PASS 1
	cmd1 := exec.Command(
		"ffmpeg",
		"-y",
		"-i", input,
		"-c:v", "libx264",
		"-preset", "medium",
		"-b:v", vbStr,
		"-pass", "1",
		"-c:a", "aac",
		"-b:a", abStr,
		"-f", "mp4", "/dev/null",
	)
	// cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	if err := cmd1.Run(); err != nil {
		fmt.Println("Error running ffmpeg pass 1:", err)
		return
	}

	// PASS 2
	cmd2 := exec.Command(
		"ffmpeg",
		"-i", input,
		"-c:v", "libx264",
		"-preset", "medium",
		"-b:v", vbStr,
		"-pass", "2",
		"-c:a", "aac",
		"-b:a", abStr,
		output,
	)
	// cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		fmt.Println("Error running ffmpeg pass 2:", err)
	}
}

func CalculateBitrates(duration float64, tv, ta float64) (float64, float64) {

	targetVideoBitrate := float64(tv) * float64(8388.608) / duration
	targetAudioBitrate := float64(ta) * float64(8388.608) / duration
	return targetVideoBitrate * 1000, targetAudioBitrate * 1000
}
