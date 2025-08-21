module github.com/gofromzero/mer-sys/backend/services/product-service

go 1.25.0

replace github.com/gofromzero/mer-sys/backend/shared => ../../shared

require (
	github.com/gofromzero/mer-sys/backend/shared v0.0.0-00010101000000-000000000000
	github.com/gogf/gf/contrib/drivers/mysql/v2 v2.9.0
	github.com/gogf/gf/v2 v2.9.0
	github.com/smartystreets/goconvey v1.8.1
)