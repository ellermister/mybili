package download

import (
	"encoding/json"

	"gorm.io/gorm"

	"mybili/internal/model"
)

func readSettingString(db *gorm.DB, name string) (string, error) {
	raw, err := readSettingRaw(db, name)
	if err != nil || len(raw) == 0 {
		return "", err
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	return string(raw), nil
}

func readSettingRaw(db *gorm.DB, name string) (json.RawMessage, error) {
	var row model.Setting
	err := db.Where("name = ?", name).Limit(1).Find(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 || row.Value == nil {
		return nil, nil
	}
	return json.RawMessage(*row.Value), nil
}
