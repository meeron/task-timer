build:
	GOARCH=wasm GOOS=js go build -o web/app.wasm
	go build -o bin/task-timer

run: build
	./bin/task-timer
