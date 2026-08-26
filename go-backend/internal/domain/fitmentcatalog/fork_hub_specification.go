package fitmentcatalog

import "time"

type ForkHubSpecification struct {
	ForkEntryID        uint      `gorm:"primaryKey;autoIncrement:false" json:"fork_entry_id"`
	HubSpecificationID uint      `gorm:"primaryKey;autoIncrement:false" json:"hub_specification_id"`
	CreatedAt          time.Time `json:"created_at"`
}

func (ForkHubSpecification) TableName() string {
	return "fitment_fork_hub_specifications"
}
