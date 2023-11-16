FILENAME = urls.txt

run:
	go run ./cmd/url-processor $(FILENAME)

test-coverage:
	go test ./test --cover --count=1 -v