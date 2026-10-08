module github.com/jochem11/inventory-manager/services/user-service

go 1.26.2

require (
	github.com/go-playground/locales v0.14.2
	github.com/go-playground/mold/v4 v4.5.1
	github.com/go-playground/universal-translator v0.18.2
	github.com/go-playground/validator/v10 v10.30.5
	github.com/segmentio/ksuid v1.0.4
	github.com/twmb/franz-go v1.22.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.71.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260928230214-8a89bd6388cc
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
	gorm.io/gorm v1.31.2
	gorm.io/plugin/opentelemetry v0.1.16
)

require (
	github.com/twmb/franz-go/pkg/kmsg v1.14.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk v1.47.0 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/ClickHouse/ch-go v0.61.5 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.30.0 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gosimple/slug v1.15.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jochem11/inventory-manager/shared v0.0.0
	github.com/klauspost/compress v1.20.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.30 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/go-camelcase v0.0.0-20160726192923-7085f1e3c734 // indirect
	github.com/segmentio/go-snakecase v1.2.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260928230214-8a89bd6388cc // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/clickhouse v0.7.0 // indirect
	gorm.io/driver/postgres v1.5.11 // indirect
)

replace github.com/jochem11/inventory-manager/shared => ../../shared
