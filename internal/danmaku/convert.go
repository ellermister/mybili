package danmaku

import (
	"encoding/json"
	"fmt"
	"math"
)

type Row struct {
	Content  string
	Progress int
	Mode     int
	Color    *int
}

type PlayerItem struct {
	Text   string         `json:"text"`
	Time   float64        `json:"time"`
	Mode   int            `json:"mode"`
	Color  string         `json:"color"`
	Border bool           `json:"border"`
	Style  map[string]any `json:"style"`
}

func Convert(rows []Row) []PlayerItem {
	out := make([]PlayerItem, 0, len(rows))
	for _, row := range rows {
		color := 16777215
		if row.Color != nil {
			color = *row.Color
		}
		out = append(out, PlayerItem{
			Text:   downgrade(row.Content),
			Time:   math.Round(float64(row.Progress)/1000*1000) / 1000,
			Mode:   playerMode(row.Mode),
			Color:  fmt.Sprintf("#%06X", color&0xFFFFFF),
			Border: false,
			Style:  map[string]any{},
		})
	}
	return out
}

func playerMode(mode int) int {
	switch mode {
	case 5:
		return 1
	case 4:
		return 2
	default:
		return 0
	}
}

func downgrade(content string) string {
	var decoded []any
	if err := json.Unmarshal([]byte(content), &decoded); err == nil && len(decoded) > 4 {
		return fmt.Sprint(decoded[4])
	}
	return content
}
