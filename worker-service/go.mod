module alla/worker-service

go 1.25.7

require (
	github.com/gorilla/mux v1.8.1
	github.com/stretchr/testify v1.7.5
	github.com/subosito/gotenv v1.6.0
	google.golang.org/grpc v1.79.3
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace alla/shared => ../shared

require (
	alla/shared v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.33.0 // indirect
)
