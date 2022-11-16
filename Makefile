.PHONY: build-css
build-css: tailwindcss
	./tailwindcss -i tailwind.css -o public/styles/app.css --minify

.PHONY: cover
cover:
	go tool cover -html=cover.out

.PHONY: lint
lint:
	golangci-lint run

tailwindcss:
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mv tailwindcss-macos-arm64 tailwindcss

.PHONY: test
test:
	go test -coverprofile=cover.out -shuffle on ./...

.PHONY: watch-css
watch-css: tailwindcss
	./tailwindcss -i tailwind.css -o public/styles/app.css --watch
