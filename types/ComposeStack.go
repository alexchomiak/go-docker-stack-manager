package types

import "gorm.io/gorm"

type ComposeStack struct {
	gorm.Model
	StackID             string `json:"stackId" gorm:"primaryKey"`
	ComposeFileContents string `json:"composeFileContents"`
	ComposeFileHash     string `json:"composeFileHash" gorm:"unique"`
}
