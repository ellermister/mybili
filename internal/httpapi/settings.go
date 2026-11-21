package httpapi

import (
	"encoding/json"
	"net/http"

	"mybili/internal/dbtime"
	"mybili/internal/model"
)

func settingPresets() map[string]any {
	return map[string]any{
		"danmaku_download_enabled":         "off",
		"video_download_enabled":           "off",
		"favorite_sync_enabled":            "off",
		"human_readable_name_enabled":      "off",
		"usage_analytics_enabled":          "on",
		"fav_exclude":                      map[string]any{"enabled": false, "selected": []any{}},
		"multi_partition_download_enabled": "off",
		"name_exclude":                     map[string]any{"contains": "", "regex": "", "type": "off"},
		"video_quality":                    "1",
		"source_quality":                   map[string]any{},
		"duration_video_exclude":           map[string]any{"custom_duration": 0, "type": "off"},
		"duration_video_part_exclude":      map[string]any{"custom_duration": 0, "type": "off"},
		"fav_time_exclude":                 map[string]any{"custom_date": "", "type": "off"},
		"telegram_bot_api_url":             "",
		"telegram_bot_enabled":             "off",
		"telegram_bot_token":               "",
		"telegram_chat_id":                 "",
	}
}

func (a *App) getSettings(w http.ResponseWriter, r *http.Request) {
	presets := settingPresets()
	var rows []model.Setting
	if err := a.DB.Find(&rows).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	stored := map[string]any{}
	for _, row := range rows {
		stored[row.Name] = decodeSetting(row.Value)
	}
	for key := range presets {
		if truthy(stored[key]) {
			presets[key] = stored[key]
		}
	}
	writeJSON(w, http.StatusOK, presets)
}

func (a *App) saveSettings(w http.ResponseWriter, r *http.Request) {
	if a.Config.SettingReadOnly {
		writeJSON(w, http.StatusForbidden, map[string]any{"message": "read-only mode"})
		return
	}
	var payload map[string]any
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "request format is incorrect"})
		return
	}
	presets := settingPresets()
	for key := range presets {
		value, ok := payload[key]
		if !ok {
			continue
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"message": err.Error()})
			return
		}
		text := string(encoded)
		now := dbtime.DateTime{Time: dbtime.Now()}
		var existing model.Setting
		err = a.DB.Where("name = ?", key).First(&existing).Error
		if err != nil {
			row := model.Setting{Name: key, Value: &text, CreatedAt: &now, UpdatedAt: &now}
			if err := a.DB.Create(&row).Error; err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
				return
			}
			continue
		}
		if err := a.DB.Model(&existing).Updates(map[string]any{"value": text, "updated_at": now}).Error; err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Settings saved successfully"})
}

func decodeSetting(raw *string) any {
	if raw == nil {
		return nil
	}
	var value any
	if err := json.Unmarshal([]byte(*raw), &value); err != nil {
		return *raw
	}
	return value
}

func truthy(value any) bool {
	switch item := value.(type) {
	case nil:
		return false
	case string:
		return item != "" && item != "0"
	case bool:
		return item
	case float64:
		return item != 0
	default:
		return true
	}
}
