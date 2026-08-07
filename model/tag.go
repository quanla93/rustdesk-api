package model

type Tag struct {
	IdModel
	Name         string                 `json:"name" gorm:"default:'';not null;"`
	UserId       uint                   `json:"user_id" gorm:"default:0;not null;index"`
	Color        uint                   `json:"color" gorm:"default:0;not null;"` // Color is a Flutter color value from 0x00000000 to 0xFFFFFFFF; the first two digits are alpha and the last six are color, convertible to RGBA
	CollectionId uint                   `json:"collection_id" gorm:"default:0;not null;index"`
	Collection   *AddressBookCollection `json:"collection,omitempty"`
	TimeModel
}

type TagList struct {
	Tags []*Tag `json:"list"`
	Pagination
}
