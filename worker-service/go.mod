module alla/worker-service

go 1.25.7

require (
	github.com/gorilla/mux v1.8.1
	github.com/stretchr/testify v1.7.5
	github.com/subosito/gotenv v1.6.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace alla/shared => ../shared

require (
	alla/shared v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.33.0 // indirect
)
