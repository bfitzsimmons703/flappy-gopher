play:
	go run main.go

build:
	rm -f flappy-gopher && go build -ldflags "-s -w"