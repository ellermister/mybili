package download

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybili/internal/dbtime"
)

type rules struct {
	multi         bool
	nameType      string
	nameContains  string
	nameRegex     string
	videoLimitSec int
	partLimitSec  int
	favCutoff     time.Time
	excludeFav    bool
	excluded      map[int64]bool
}

func loadRules(db *gorm.DB, now time.Time) (rules, error) {
	var out rules
	out.excluded = map[int64]bool{}
	multi, err := readSettingString(db, "multi_partition_download_enabled")
	if err != nil {
		return out, err
	}
	out.multi = multi == "on"
	var name struct {
		Type     string `json:"type"`
		Contains string `json:"contains"`
		Regex    string `json:"regex"`
	}
	if err := readSettingJSON(db, "name_exclude", &name); err != nil {
		return out, err
	}
	out.nameType = name.Type
	out.nameContains = name.Contains
	out.nameRegex = name.Regex
	out.videoLimitSec = durationLimit(db, "duration_video_exclude")
	out.partLimitSec = durationLimit(db, "duration_video_part_exclude")
	out.favCutoff = favCutoff(db, now)
	var fav struct {
		Enabled  bool    `json:"enabled"`
		Selected []int64 `json:"selected"`
	}
	if err := readSettingJSON(db, "fav_exclude", &fav); err != nil {
		return out, err
	}
	out.excludeFav = fav.Enabled
	for _, id := range fav.Selected {
		out.excluded[id] = true
	}
	return out, nil
}

func readSettingJSON(db *gorm.DB, name string, dest any) error {
	raw, err := readSettingRaw(db, name)
	if err != nil || len(raw) == 0 {
		return err
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return nil
	}
	return nil
}

func durationLimit(db *gorm.DB, name string) int {
	var body struct {
		Type           string  `json:"type"`
		CustomDuration float64 `json:"custom_duration"`
	}
	_ = readSettingJSON(db, name, &body)
	switch body.Type {
	case "30min":
		return 30 * 60
	case "60min":
		return 60 * 60
	case "custom":
		if body.CustomDuration <= 0 {
			return 0
		}
		return int(body.CustomDuration * 60)
	default:
		return 0
	}
}

func favCutoff(db *gorm.DB, now time.Time) time.Time {
	var body struct {
		Type       string `json:"type"`
		CustomDate string `json:"custom_date"`
	}
	_ = readSettingJSON(db, nameFavTime, &body)
	now = now.In(dbtime.Location())
	switch body.Type {
	case "1m":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, -1, 0)
	case "3m":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, -3, 0)
	case "6m":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, -6, 0)
	case "custom":
		parsed, err := time.ParseInLocation("2006-01-02", body.CustomDate, now.Location())
		if err != nil {
			return time.Time{}
		}
		return parsed
	default:
		return time.Time{}
	}
}

const nameFavTime = "fav_time_exclude"

type Candidate struct {
	ID            int64
	VideoID       int64
	Page          int
	Cid           int64
	Duration      int
	Title         string
	VideoDuration int
	FavTime       *dbtime.DateTime
	Pubtime       dbtime.DateTime
}

func (r rules) allow(row Candidate, favIDs, subIDs []int64) bool {
	if !r.multi && row.Page > 1 {
		return false
	}
	if r.videoLimitSec > 0 && row.VideoDuration > r.videoLimitSec {
		return false
	}
	if r.partLimitSec > 0 && row.Duration > r.partLimitSec {
		return false
	}
	if r.excludedByName(row.Title) {
		return false
	}
	if !r.favCutoff.IsZero() {
		when := row.Pubtime.Time
		if row.FavTime != nil && !row.FavTime.IsZero() {
			when = row.FavTime.Time
		}
		if !when.IsZero() && when.Before(r.favCutoff) {
			return false
		}
	}
	if r.excludeFav && excludedByRelation(r.excluded, favIDs, subIDs) {
		return false
	}
	return true
}

func (r rules) excludedByName(title string) bool {
	if r.nameType == "" || r.nameType == "off" {
		return false
	}
	if r.nameContains != "" && strings.Contains(title, r.nameContains) {
		return true
	}
	if r.nameRegex == "" {
		return false
	}
	matched, err := regexp.MatchString(r.nameRegex, title)
	if err != nil {
		return false
	}
	return matched
}

func excludedByRelation(excluded map[int64]bool, favIDs, subIDs []int64) bool {
	if len(favIDs) == 0 && len(subIDs) == 0 {
		return false
	}
	for _, id := range favIDs {
		if !excluded[id] {
			return false
		}
	}
	for _, id := range subIDs {
		if !excluded[-id] {
			return false
		}
	}
	return true
}
