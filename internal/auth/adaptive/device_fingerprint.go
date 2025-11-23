package adaptive

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type DeviceFingerprint struct {
	// store FingerprintStore // TODO:
}

type FingerprintData struct {
	UserAgent         string            `json:"user_agent"`
	ScreenResolution  string            `json:"screen_resolution"`
	Timezone          int               `json:"timezone"`
	Language          string            `json:"language"`
	Platform          string            `json:"platform"`
	Plugins           []string          `json:"plugins"`
	CanvasHash        string            `json:"canvas_hash"`
	WebGLHash         string            `json:"webgl_hash"`
	AudioHash         string            `json:"audio_hash"`
	Fonts             []string          `json:"fonts"`
	ClientHintsMobile string            `json:"ch_mobile"`
	ClientHintsPlat   string            `json:"ch_platform"`
	ClientHintsArch   string            `json:"ch_arch"`
	ClientHintsModel  string            `json:"ch_model"`
	CustomData        map[string]string `json:"custom_data"`
}

func (df *DeviceFingerprint) Generate(data *FingerprintData) string {
	features := []string{
		data.UserAgent,
		data.ScreenResolution,
		fmt.Sprintf("%d", data.Timezone),
		data.Language,
		data.Platform,
		strings.Join(data.Plugins, ","),
		data.CanvasHash,
		data.WebGLHash,
		data.AudioHash,
		strings.Join(data.Fonts, ","),
		data.ClientHintsMobile,
		data.ClientHintsPlat,
		data.ClientHintsArch,
		data.ClientHintsModel,
	}

	combined := strings.Join(features, "|")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

func (df *DeviceFingerprint) CompareSimilarity(fp1, fp2 string) float64 {
	// TODO:
	return 0
}

func (df *DeviceFingerprint) TrustDevice(ctx context.Context, userID, fingerprint string, duration time.Duration) error {
	// TODO:
	return nil
}
