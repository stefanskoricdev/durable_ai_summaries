.PHONY: watch-worker
watch-worker:
	@export $$(cat .env | xargs) && \
    ./bin/air -c worker.air.toml


.PHONY: watch-app
watch-app:
	@export $$(cat .env | xargs) && \
    ./bin/air -c app.air.toml

.ONESHELL:
setup:
	@curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s