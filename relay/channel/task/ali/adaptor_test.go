package ali

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/require"
)

func testRelayInfo() *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{},
	}
}

func TestConvertToAliRequestWan27I2VBuildsMediaFromImage(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:    "wan2.7-i2v",
		Prompt:   "animate the first frame",
		Image:    "https://example.com/first.png",
		Size:     "720p",
		Duration: 10,
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, "wan2.7-i2v", aliReq.Model)
	require.Equal(t, "720P", aliReq.Parameters.Resolution)
	require.Equal(t, 10, aliReq.Parameters.Duration)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/first.png"},
	}, aliReq.Input.Media)
	require.Empty(t, aliReq.Input.ImgURL)

	body, err := common.Marshal(aliReq)
	require.NoError(t, err)
	require.Contains(t, string(body), `"media"`)
	require.NotContains(t, string(body), `"img_url"`)
}

func TestConvertToAliRequestWan27I2VBuildsFirstAndLastFrameFromImages(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.7-i2v",
		Prompt: "interpolate between frames",
		Images: []string{
			"https://example.com/first.png",
			"https://example.com/last.png",
		},
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/first.png"},
		{Type: "last_frame", URL: "https://example.com/last.png"},
	}, aliReq.Input.Media)
}

func TestConvertToAliRequestWan27I2VPrefersImageBeforeImagesAndInputReference(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:          "wan2.7-i2v",
		Prompt:         "use the direct image",
		Image:          " https://example.com/direct.png ",
		Images:         []string{"https://example.com/images-first.png", " https://example.com/images-last.png "},
		InputReference: "https://example.com/input-reference.png",
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/direct.png"},
		{Type: "last_frame", URL: "https://example.com/images-last.png"},
	}, aliReq.Input.Media)
}

func TestConvertToAliRequestWan27I2VFallsBackToFirstNonEmptyImage(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.7-i2v",
		Prompt: "skip blank images",
		Image:  " ",
		Images: []string{
			" ",
			" https://example.com/first.png ",
			" https://example.com/last.png ",
		},
		InputReference: "https://example.com/input-reference.png",
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/first.png"},
		{Type: "last_frame", URL: "https://example.com/last.png"},
	}, aliReq.Input.Media)
}

func TestConvertToAliRequestWan27I2VKeepsExplicitMetadataMedia(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:          "wan2.7-i2v",
		Prompt:         "continue the clip",
		Image:          "https://example.com/direct.png",
		Images:         []string{"https://example.com/images-first.png", "https://example.com/images-last.png"},
		InputReference: "https://example.com/input-reference.png",
		Metadata: map[string]interface{}{
			"input": map[string]interface{}{
				"media": []interface{}{
					map[string]interface{}{
						"type": "first_clip",
						"url":  "https://example.com/input.mp4",
					},
				},
			},
		},
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_clip", URL: "https://example.com/input.mp4"},
	}, aliReq.Input.Media)
	require.Empty(t, aliReq.Input.ImgURL)

	body, err := common.Marshal(aliReq)
	require.NoError(t, err)
	require.Contains(t, string(body), `"media"`)
	require.NotContains(t, string(body), `"img_url"`)
}

func TestConvertToAliRequestWan27I2VRequiresMedia(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.7-i2v",
		Prompt: "animate without a frame",
	}

	_, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "requires image"))
}

func TestConvertToAliRequestWan25I2VKeepsLegacyImgURL(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.5-i2v-preview",
		Prompt: "animate the first frame",
		Image:  "https://example.com/first.png",
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, "https://example.com/first.png", aliReq.Input.ImgURL)
	require.Empty(t, aliReq.Input.Media)

	body, err := common.Marshal(aliReq)
	require.NoError(t, err)
	require.Contains(t, string(body), `"img_url"`)
	require.NotContains(t, string(body), `"media"`)
}

func TestProcessAliOtherRatiosWan27(t *testing.T) {
	tests := []struct {
		name       string
		model      string
		resolution string
		want       map[string]float64
	}{
		{
			name:       "wan2.7-i2v 720P",
			model:      "wan2.7-i2v",
			resolution: "720P",
			want:       map[string]float64{"resolution-720P": 1},
		},
		{
			name:       "wan2.7-i2v 1080P",
			model:      "wan2.7-i2v",
			resolution: "1080P",
			want:       map[string]float64{"resolution-1080P": 1 / 0.6},
		},
		{
			name:       "wan2.7-t2v 720P",
			model:      "wan2.7-t2v",
			resolution: "720P",
			want:       map[string]float64{"resolution-720P": 1},
		},
		{
			name:       "wan2.7-t2v 1080P",
			model:      "wan2.7-t2v",
			resolution: "1080P",
			want:       map[string]float64{"resolution-1080P": 1 / 0.6},
		},
		{
			name:       "wan2.7-videoedit 720P",
			model:      "wan2.7-videoedit",
			resolution: "720P",
			want:       map[string]float64{"resolution-720P": 1},
		},
		{
			name:       "wan2.7-videoedit 1080P",
			model:      "wan2.7-videoedit",
			resolution: "1080P",
			want:       map[string]float64{"resolution-1080P": 1 / 0.6},
		},
		{
			name:       "wan2.7-r2v dated prefix match 720P",
			model:      "wan2.7-r2v-2026-06-12",
			resolution: "720P",
			want:       map[string]float64{"resolution-720P": 1},
		},
		{
			name:       "wan2.7-r2v dated prefix match 1080P",
			model:      "wan2.7-r2v-2026-06-12",
			resolution: "1080P",
			want:       map[string]float64{"resolution-1080P": 1 / 0.6},
		},
		{
			name:       "wan2.7-i2v unknown resolution no match",
			model:      "wan2.7-i2v",
			resolution: "2k",
			want:       map[string]float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aliReq := &AliVideoRequest{
				Model: tt.model,
				Parameters: &AliVideoParameters{
					Resolution: tt.resolution,
				},
			}
			got, err := ProcessAliOtherRatios(aliReq)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLookupAliRatiosPrefixMatch(t *testing.T) {
	ratios := map[string]map[string]float64{
		"wan2.7-r2v": {"720P": 1, "1080P": 1 / 0.6},
	}

	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "exact match", model: "wan2.7-r2v", want: true},
		{name: "dated prefix match", model: "wan2.7-r2v-2026-06-12", want: true},
		{name: "no match", model: "wan2.7-i2v", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := lookupAliRatios(tt.model, ratios)
			require.Equal(t, tt.want, ok)
		})
	}
}
