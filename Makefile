# Useful commands. Run 'make help' to show available tasks.


default: help


.PHONY: intro
intro:
	@echo -e '\n\e[1;34m------ [simple-backup-go] $(shell date) ------\e[0m\\n'


.PHONY: build
build: intro ## build executable file
	go build


.PHONY: help
help: intro
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":[^:]*?## "}; {printf "\033[1;38;5;69m%-15s\033[0;38;5;38m %s\033[0m\n", $$1, $$2}'
