package httpapi

import (
	"mybili/internal/danmaku"
	"mybili/internal/model"
)

func (a *App) playerDanmaku(rows []model.Danmaku) []danmaku.PlayerItem {
	input := make([]danmaku.Row, len(rows))
	for i, row := range rows {
		mode := row.Mode
		if mode == 0 {
			mode = 1
		}
		input[i] = danmaku.Row{
			Content:  row.Content,
			Progress: row.Progress,
			Mode:     mode,
			Color:    row.Color,
		}
	}
	return danmaku.Convert(input)
}
