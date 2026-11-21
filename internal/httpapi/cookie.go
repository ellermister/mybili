package httpapi

import (
	"io"
	"net/http"
	"strings"

	"mybili/internal/bilibili"
	"mybili/internal/dbtime"
	"mybili/internal/model"
	"mybili/internal/notify"
)

func (a *App) cookieContent() (string, error) {
	var row model.Setting
	err := a.DB.Where("name = ?", "cookies_content").First(&row).Error
	if err != nil {
		return "", nil
	}
	decoded := decodeSetting(row.Value)
	text, _ := decoded.(string)
	return text, nil
}

func (a *App) cookieExist(w http.ResponseWriter, r *http.Request) {
	content, err := a.cookieContent()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"exist": strings.TrimSpace(content) != ""})
}

func (a *App) cookieStatus(w http.ResponseWriter, r *http.Request) {
	content, err := a.cookieContent()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	if strings.TrimSpace(content) == "" {
		writeJSON(w, http.StatusOK, map[string]any{"logged": false})
		return
	}
	logged, err := bilibili.LoggedIn(content)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"logged": logged})
}

func (a *App) cookieUpload(w http.ResponseWriter, r *http.Request) {
	if a.Config.SettingReadOnly {
		writeJSON(w, http.StatusForbidden, map[string]any{"message": "read-only mode"})
		return
	}
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "cannot read uploaded file"})
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "missing cookie file"})
		return
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, 2<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": err.Error()})
		return
	}
	encoded, err := jsonMarshal(string(body))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	text := string(encoded)
	now := dbtime.DateTime{Time: dbtime.Now()}
	var existing model.Setting
	err = a.DB.Where("name = ?", "cookies_content").First(&existing).Error
	if err != nil {
		row := model.Setting{Name: "cookies_content", Value: &text, CreatedAt: &now, UpdatedAt: &now}
		if err := a.DB.Create(&row).Error; err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
			return
		}
	} else if err := a.DB.Model(&existing).Updates(map[string]any{"value": text, "updated_at": now}).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	logged := false
	if strings.TrimSpace(string(body)) != "" {
		logged, _ = bilibili.LoggedIn(string(body))
	}
	if q, ok := a.Tasks.(notify.Queue); ok {
		_ = notify.ClearCookieNotice(r.Context(), q)
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "valid": logged})
}
