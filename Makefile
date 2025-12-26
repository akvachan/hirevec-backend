.PHONY: run watch

run:
	go build -o app && ./app

watch:
	watchexec -e go -r -c -- make run
