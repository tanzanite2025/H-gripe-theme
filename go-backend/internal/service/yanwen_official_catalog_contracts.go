package service

import "time"

type YanwenProductSyncSummary struct {
	Environment      string    `json:"environment"`
	Scanned          int       `json:"scanned"`
	Added            int       `json:"added"`
	Updated          int       `json:"updated"`
	ChannelsDisabled int64     `json:"channels_disabled"`
	SyncedAt         time.Time `json:"synced_at"`
}

type YanwenCountrySyncSummary struct {
	Environment string    `json:"environment"`
	Scanned     int       `json:"scanned"`
	Added       int       `json:"added"`
	Updated     int       `json:"updated"`
	SyncedAt    time.Time `json:"synced_at"`
}

type YanwenWarehouseSyncSummary struct {
	Environment string    `json:"environment"`
	Scanned     int       `json:"scanned"`
	Added       int       `json:"added"`
	Updated     int       `json:"updated"`
	SyncedAt    time.Time `json:"synced_at"`
}
