module github.com/jochem11/inventory-manager/services/item-service

go 1.26.2

require gorm.io/gorm v1.31.2

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jochem11/inventory-manager/shared v0.0.0
	github.com/segmentio/ksuid v1.0.4
	golang.org/x/text v0.42.0 // indirect
)

replace github.com/jochem11/inventory-manager/shared => ../../shared
