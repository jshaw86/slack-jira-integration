build:
	go install github.com/golang/mock/mockgen@v1.6.0
	go generate cmd/slack-jira-integration/main.go
	go build -o main cmd/slack-jira-integration/main.go

run:
	go run cmd/slack-jira-integration/main.go 

clean:
	rm main 

.PHONY: clean run
