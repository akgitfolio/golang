migrate create -ext sql -dir db/migrations -seq init_schema
go run main.go -action=version