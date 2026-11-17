package models

// Item is a single inventory record, mapped to the "items" table by GORM.
//
// TODO(2): Review the field types (UUID string ID, string Name, int Stock, float64 Price).
// TODO(2): Add GORM tags, e.g. `gorm:"primaryKey"` on ID and `gorm:"not null"` on Name.
// TODO(2): Add JSON tags (id, name, stock, price) so responses use lowercase keys.
// TODO(2): Add `binding:"required"` / min-value validation tags for Create/Update requests.
// TODO(2): Generate a new UUID for ID when an item is created (BeforeCreate hook or in the handler).
type Item struct {
	ID    string
	Name  string
	Stock int
	Price float64
}
