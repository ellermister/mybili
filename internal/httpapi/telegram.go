package httpapi

import (
	"net/http"

	"mybili/internal/notify"
)

func (a *App) testTelegram(w http.ResponseWriter, r *http.Request) {
	if a.Config.SettingReadOnly {
		writeJSON(w, http.StatusForbidden, map[string]any{"message": "read-only mode"})
		return
	}
	var body struct {
		Token  string `json:"telegram_bot_token"`
		ChatID string `json:"telegram_chat_id"`
		APIURL string `json:"telegram_bot_api_url"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "Telegram connection test failed", "result": false})
		return
	}
	ok, _ := notify.Test(r.Context(), body.Token, body.ChatID, body.APIURL)
	message := "Telegram connection test successful"
	if !ok {
		message = "Telegram connection test failed"
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": message, "result": ok})
}
