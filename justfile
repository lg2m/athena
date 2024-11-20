binary_name := "ath"

run arg:
    @go run cmd/athena/main.go {{arg}}

build:
    @go build -o {{binary_name}} cmd/athena/main.go

clean:
    @echo "Cleaning up.."
    rm {{binary_name}}
    
test:
    @echo "Running test cases"
    @go test -v ./...
