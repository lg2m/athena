binary_name := "ath"

run arg:
    @go run cmd/athena/main.go {{arg}}

build:
    @go build -o {{binary_name}} cmd/athena/main.go
    
test:
    @echo "Running test cases"
    @go test -v ./...
