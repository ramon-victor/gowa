package utils

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os/exec"
	"strings"

	_ "golang.org/x/image/webp"
)

const (
	// WebP RIFF container constants
	riffHeaderSize  = 12 // "RIFF" + size (4) + "WEBP"
	chunkHeaderSize = 8  // tag (4) + size (4)
	riffSizeOffset  = 4  // Offset to RIFF size field

	// VP8X extended header chunk layout (10-byte payload)
	vp8xChunkSize    = chunkHeaderSize + 10
	vp8xPayloadSize  = 10
	vp8xFlagsOffset  = chunkHeaderSize     // Byte 0 of payload: feature flags
	vp8xWidthOffset  = chunkHeaderSize + 4 // Bytes 4-6: canvas width - 1 (24-bit LE)
	vp8xHeightOffset = chunkHeaderSize + 7 // Bytes 7-9: canvas height - 1 (24-bit LE)

	// VP8X feature flags
	vp8xFlagEXIF byte = 0x08
)

func ConvertToWebP(ctx context.Context, inputPath string, outputPath string, mimeType string) error {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is not installed for WebP conversion")
	}

	var ffmpegArgs []string
	if strings.HasPrefix(mimeType, "video/") || mimeType == "image/gif" {
		// Animated sticker conversion
		ffmpegArgs = []string{
			"-y",
			"-t", "10",
			"-i", inputPath,
			"-vf", "fps=15,scale=512:512",
			"-loop", "0",
			"-an",
			"-vsync", "0",
			"-fs", "1000000",
			"-c:v", "libwebp",
			"-qscale:v", "10",
			outputPath,
		}
	} else {
		// Image sticker conversion
		ffmpegArgs = []string{
			"-y",
			"-i", inputPath,
			"-vf", "scale=512:512",
			"-c:v", "libwebp",
			"-lossless", "1",
			outputPath,
		}
	}

	convertCmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)
	var stderr bytes.Buffer
	convertCmd.Stderr = &stderr

	if err := convertCmd.Run(); err != nil {
		return fmt.Errorf("failed to convert sticker to WebP: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

func EmbedStickerEXIF(inputWebP []byte, packID, packName, packPublisher string, emojis []string) ([]byte, error) {
	meta := buildStickerMetadata(packID, packName, packPublisher, emojis)
	if meta == nil {
		return inputWebP, nil
	}

	exifData := buildWhatsAppEXIF(meta)
	out, err := injectWebPEXIF(inputWebP, exifData)
	if err != nil {
		return inputWebP, err
	}
	return out, nil
}

func buildStickerMetadata(packID, packName, packPublisher string, emojis []string) map[string]interface{} {
	if packID == "" && packName == "" && packPublisher == "" && len(emojis) == 0 {
		return nil
	}

	meta := make(map[string]interface{})
	if packID != "" {
		meta["sticker-pack-id"] = packID
	}
	if packName != "" {
		meta["sticker-pack-name"] = packName
	}
	if packPublisher != "" {
		meta["sticker-pack-publisher"] = packPublisher
	}
	if len(emojis) > 0 {
		meta["emojis"] = emojis
	}
	return meta
}

func buildWhatsAppEXIF(meta map[string]interface{}) []byte {
	jsonBytes, err := json.Marshal(meta)
	if err != nil {
		return nil
	}

	// WhatsApp sticker EXIF header structure
	header := []byte{
		0x49, 0x49, 0x2A, 0x00, // TIFF little-endian marker
		0x08, 0x00, 0x00, 0x00, // IFD offset
		0x01, 0x00, // Number of directory entries
		0x41, 0x57, // Tag ID (WhatsApp custom)
		0x07, 0x00, // Data type (undefined)
	}
	footer := []byte{0x16, 0x00, 0x00, 0x00} // Next IFD offset

	var buf bytes.Buffer
	buf.Write(header)
	binary.Write(&buf, binary.LittleEndian, uint32(len(jsonBytes)))
	buf.Write(footer)
	buf.Write(jsonBytes)

	return buf.Bytes()
}

func injectWebPEXIF(in []byte, exif []byte) ([]byte, error) {
	if !isValidWebP(in) {
		return nil, fmt.Errorf("not a RIFF WEBP file")
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(in))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	chunks, vp8xIndex, err := parseWebPChunks(in)
	if err != nil {
		return nil, err
	}

	chunks = ensureVP8XWithEXIF(chunks, vp8xIndex, cfg.Width, cfg.Height)

	return assembleWebP(chunks, exif), nil
}

func isValidWebP(data []byte) bool {
	return len(data) >= riffHeaderSize &&
		string(data[0:4]) == "RIFF" &&
		string(data[8:12]) == "WEBP"
}

func parseWebPChunks(in []byte) (chunks [][]byte, vp8xIndex int, err error) {
	vp8xIndex = -1
	pos := riffHeaderSize

	for pos+chunkHeaderSize <= len(in) {
		tag := string(in[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(in[pos+4 : pos+8]))
		dataEnd := pos + chunkHeaderSize + size

		if dataEnd > len(in) {
			return nil, -1, fmt.Errorf("truncated webp chunk: %s", tag)
		}

		pad := size & 1
		if tag == "VP8X" && size >= vp8xPayloadSize {
			vp8xIndex = len(chunks)
		}
		if tag != "EXIF" {
			chunk := make([]byte, chunkHeaderSize+size+pad)
			copy(chunk, in[pos:dataEnd])
			if pad == 1 {
				chunk[chunkHeaderSize+size] = 0
			}
			chunks = append(chunks, chunk)
		}
		pos = dataEnd + pad
	}
	return chunks, vp8xIndex, nil
}

func ensureVP8XWithEXIF(chunks [][]byte, vp8xIndex, width, height int) [][]byte {
	if vp8xIndex >= 0 {
		chunks[vp8xIndex][vp8xFlagsOffset] |= vp8xFlagEXIF
		return chunks
	}
	return append([][]byte{createVP8XChunk(width, height)}, chunks...)
}

func createVP8XChunk(width, height int) []byte {
	chunk := make([]byte, vp8xChunkSize)
	copy(chunk[0:4], "VP8X")
	binary.LittleEndian.PutUint32(chunk[4:8], vp8xPayloadSize)
	chunk[vp8xFlagsOffset] = vp8xFlagEXIF
	putUint24LE(chunk[vp8xWidthOffset:], width-1)
	putUint24LE(chunk[vp8xHeightOffset:], height-1)
	return chunk
}

func putUint24LE(b []byte, v int) {
	b[0] = uint8(v)
	b[1] = uint8(v >> 8)
	b[2] = uint8(v >> 16)
}

func assembleWebP(chunks [][]byte, exif []byte) []byte {
	var out bytes.Buffer
	out.WriteString("RIFF")
	out.Write([]byte{0, 0, 0, 0})
	out.WriteString("WEBP")

	for _, c := range chunks {
		out.Write(c)
	}

	writeChunk(&out, "EXIF", exif)

	b := out.Bytes()
	binary.LittleEndian.PutUint32(b[riffSizeOffset:], uint32(len(b)-8))
	return b
}

func writeChunk(buf *bytes.Buffer, tag string, data []byte) {
	buf.WriteString(tag)
	sz := make([]byte, 4)
	binary.LittleEndian.PutUint32(sz, uint32(len(data)))
	buf.Write(sz)
	buf.Write(data)
	if len(data)%2 == 1 {
		buf.WriteByte(0)
	}
}
