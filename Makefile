build:
	GOARCH=wasm GOOS=js go build -o web/app.wasm
	go build -o bin/task-timer
	pnpm exec tailwindcss -i styles/main.css -o web/styles.css

run:
	go tool air

test:
	go test ./...
