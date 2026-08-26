package fitmentcatalog

import "time"

type FrameHubSpecification struct {
	FrameEntryID       uint      `gorm:"primaryKey;autoIncrement:false" json:"frame_entry_id"`
	HubSpecificationID uint      `gorm:"primaryKey;autoIncrement:false" json:"hub_specification_id"`
	CreatedAt          time.Time `json:"created_at"`
}

func (FrameHubSpecification) TableName() string {
	return "fitment_frame_hub_specifications"
}
