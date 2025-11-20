package image

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"

	"github.com/nfnt/resize"
	"golang.org/x/image/webp"
)

type Processor struct {
    quality    int  // JPEG quality (1-100)
    maxWidth   uint // Maximum width
    maxHeight  uint // Maximum height
}

func NewProcessor(quality int, maxWidth, maxHeight uint) *Processor {
    return &Processor{
        quality:   quality,
        maxWidth:  maxWidth,
        maxHeight: maxHeight,
    }
}

// Process compresses and resizes image
func (p *Processor) Process(file multipart.File, header *multipart.FileHeader) (io.Reader, error) {
    // Reset file pointer
    if _, err := file.Seek(0, 0); err != nil {
        return nil, err
    }

    // Decode image based on content type
    img, format, err := p.decodeImage(file, header.Header.Get("Content-Type"))
    if err != nil {
        return nil, err
    }

    // Resize if needed
    img = p.resizeImage(img)

    // Encode to JPEG with compression
    return p.encodeImage(img, format)
}

func (p *Processor) decodeImage(file io.Reader, contentType string) (image.Image, string, error) {
    switch contentType {
    case "image/jpeg", "image/jpg":
        img, err := jpeg.Decode(file)
        return img, "jpeg", err
    case "image/png":
        img, err := png.Decode(file)
        return img, "png", err
    case "image/webp":
        img, err := webp.Decode(file)
        return img, "webp", err
    default:
        // Try to decode anyway
        img, format, err := image.Decode(file)
        return img, format, err
    }
}

func (p *Processor) resizeImage(img image.Image) image.Image {
    bounds := img.Bounds()
    width := uint(bounds.Dx())
    height := uint(bounds.Dy())

    // Check if resize needed
    if width <= p.maxWidth && height <= p.maxHeight {
        return img
    }

    // Calculate new dimensions maintaining aspect ratio
    if width > height {
        return resize.Resize(p.maxWidth, 0, img, resize.Lanczos3)
    }
    return resize.Resize(0, p.maxHeight, img, resize.Lanczos3)
}

func (p *Processor) encodeImage(img image.Image, format string) (io.Reader, error) {
    var buf bytes.Buffer
    
    // Always encode to JPEG for better compression
    opts := &jpeg.Options{Quality: p.quality}
    if err := jpeg.Encode(&buf, img, opts); err != nil {
        return nil, fmt.Errorf("failed to encode image: %w", err)
    }

    return &buf, nil
}

// Default processors
func DefaultAvatarProcessor() *Processor {
    // Avatar: 400x400, quality 85
    return NewProcessor(85, 400, 400)
}

func DefaultImageProcessor() *Processor {
    // Post image: 1200x1200, quality 90
    return NewProcessor(90, 1200, 1200)
}