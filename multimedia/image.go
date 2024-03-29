package multimedia

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

// SliceImageInput is the input for SliceImage function
type SliceImageInput struct {
	SourcePath string

	// OutputPath is the directory where the sliced images will be saved.
	OutputDir    string
	MaxHeight    int
	MinHeight    int
	OutputFormat imaging.Format
	AspectRatio  float64
	OutputFiles  []string
}

// OutputFileName returns the output file name for the given y position
func (sii *SliceImageInput) OutputFileName(y int) string {
	return fmt.Sprintf("%s/%d.%s", sii.OutputDir, y, strings.ToLower(sii.OutputFormat.String()))
}

// SliceImage slices the image into multiple images with the given height.
// If the image height is smaller than the max height, it will be assigned with the min height.
// This will help to prevent the resulting images are so small due to the aspect ratio & original width.
// If the func return error, you must delete the generated files manually (if any).
func SliceImage(input *SliceImageInput) error {
	ori, err := imaging.Open(input.SourcePath)
	if err != nil {
		return err
	}

	originalWidth := ori.Bounds().Dx()
	cropHeight := int(float64(originalWidth) / input.AspectRatio)
	if cropHeight < input.MaxHeight {
		cropHeight = input.MinHeight
	}

	for y := 0; y <= ori.Bounds().Dy()-cropHeight; y += cropHeight {
		croppedImage := imaging.Crop(ori, image.Rect(0, y, originalWidth, y+cropHeight))
		outputImagePath := filepath.Join(input.OutputFileName(y))
		err := imaging.Save(croppedImage, outputImagePath)
		if err != nil {
			return err
		}

		input.OutputFiles = append(input.OutputFiles, outputImagePath)
	}

	return nil
}

// ScaleDownImageByWidthInput is the input for ScaleDownImageByWidth function
type ScaleDownImageByWidthInput struct {
	SourcePath string
	OutputPath string
	Width      int
	Filter     imaging.ResampleFilter
}

// ScaleDownImageByWidth scales down the image by width while maintaining the aspect ratio.
// If the image width is smaller than the desired width, do nothing.
func ScaleDownImageByWidth(input *ScaleDownImageByWidthInput) error {
	img, err := imaging.Open(input.SourcePath)
	if err != nil {
		return err
	}

	// if image width is smaller than the desired width, do nothing
	if img.Bounds().Dx() <= input.Width {
		return nil
	}

	// Determine the desired width (1080) while maintaining the aspect ratio
	height := int(float64(img.Bounds().Dy()) * float64(input.Width) / float64(img.Bounds().Dx()))

	// Resize the image
	resizedImg := imaging.Resize(img, input.Width, height, input.Filter)

	// Save the resized image
	return imaging.Save(resizedImg, input.OutputPath)
}

// ConvertImageInput is the input for ConvertImage function
type ConvertImageInput struct {
	SourcePath string
	OutputPath string
}

// ConvertImage converts the image to the desired format.
// The supported input and output format is relying from the "github.com/disintegration/imaging" package.
func ConvertImage(input *ConvertImageInput) error {
	img, err := imaging.Open(input.SourcePath)
	if err != nil {
		return err
	}

	return imaging.Save(img, input.OutputPath)
}

// MergeImagesToVideosInput is the input for MergeImagesToVideos function
type MergeImagesToVideosInput struct {
	// List of image file paths and their corresponding durations.
	// For example:
	// ImageDurations := map[string]float64{
	// 	"image1.jpg": 3.0,
	// 	"image2.jpg": 5.0,
	// 	"image3.jpg": 2.0,
	// }
	ImageDurations map[string]float64

	// Output video filename
	OutputPath string
	ErrStream  io.Writer
	OutStream  io.Writer
}

// MergeImagesToVideos merges multiple images into a video.
func MergeImagesToVideos(_ context.Context, input *MergeImagesToVideosInput) error {
	var cmdBuffer bytes.Buffer

	// Iterate over images and generate FFmpeg commands
	for image, duration := range input.ImageDurations {
		cmdBuffer.WriteString(fmt.Sprintf("-loop 1 -t %.2f -i %s ", duration, image))
	}

	// Execute FFmpeg command to merge images into a video
	ffmpegCmd := fmt.Sprintf("ffmpeg %s -filter_complex 'concat=n=%d:v=1:a=0[v]' -map '[v]' -c:v libx264 -pix_fmt yuv420p %s -y",
		cmdBuffer.String(), len(input.ImageDurations), input.OutputPath)

	cmd := exec.Command("bash", "-c", ffmpegCmd)
	cmd.Stdout = input.OutStream
	cmd.Stderr = input.OutStream

	return cmd.Run()
}

// ScaleUpAndFillImageInput is the input for ScaleUpAndFillImage function
type ScaleUpAndFillImageInput struct {
	SourcePath string
	OutputPath string
	Color      color.Color
	Width      int
	Height     int
}

// ScaleUpAndFillImage scales up the image to the desired width and height while maintaining the aspect ratio.
// It works by first creating image with defined width and height, fill it with the defined color, and then draw the image in the center.
func ScaleUpAndFillImage(_ context.Context, input *ScaleUpAndFillImageInput) error {
	originalImage, err := gg.LoadImage(input.SourcePath)
	if err != nil {
		return err
	}

	newImage := gg.NewContext(input.Width, input.Height)
	newImage.SetColor(input.Color)
	newImage.Clear()

	x := (input.Width - originalImage.Bounds().Dx()) / 2
	y := (input.Height - originalImage.Bounds().Dy()) / 2

	newImage.DrawImage(originalImage, x, y)

	return newImage.SavePNG(input.OutputPath)
}

// ScaleUpImageByResolutionInput is the input for ScaleUpImageByResolution function
type ScaleUpImageByResolutionInput struct {
	SourcePath string
	OutputPath string
	MaxWidth   int
	MaxHeight  int
	Filter     imaging.ResampleFilter
}

// ScaleUpImageByResolution scales up the image to the desired width and height while maintaining the aspect ratio.
func ScaleUpImageByResolution(_ context.Context, input *ScaleUpImageByResolutionInput) error {
	inputImage, err := imaging.Open(input.SourcePath)
	if err != nil {
		return err
	}

	// Calculate the new dimensions while maintaining aspect ratio
	newWidth := input.MaxWidth
	newHeight := inputImage.Bounds().Dy() * input.MaxWidth / inputImage.Bounds().Dx()

	if newHeight > input.MaxHeight {
		newHeight = input.MaxHeight
		newWidth = inputImage.Bounds().Dx() * input.MaxHeight / inputImage.Bounds().Dy()
	}

	scaledImage := imaging.Resize(inputImage, newWidth, newHeight, input.Filter)

	return imaging.Save(scaledImage, input.OutputPath)
}
