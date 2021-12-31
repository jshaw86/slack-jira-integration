build:
	go install github.com/golang/mock/mockgen
	go generate cmd/slack-jira-integration/main.go
	go build -o main cmd/slack-jira-integration/main.go

run:
	go run cmd/slack-jira-integration/main.go 

clean:
	rm main 

.PHONY: clean run
