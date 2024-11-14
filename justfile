run arg:
    @go run cmd/athena/main.go {{arg}}
    
test:
    @echo "Running test cases"
    @go test -v ./...
