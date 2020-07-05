PROJECT_ID ?=
REGION ?= asia-northeast1
IMAGE := asia.gcr.io/${PROJECT_ID}/calendar-notifier:latest
CONFIG := $(shell base64 < config.yml)

.PHONY: build
build:
	docker build -t calendar-notifier .
	docker tag calendar-notifier ${IMAGE}

.PHONY: deploy
deploy: CLOUD_RUN_SERVICE_ACCOUNT_EMAIL ?=
deploy:
	docker push ${IMAGE}
	@gcloud run deploy calendar-notifier \
	--image="${IMAGE}" \
	--region=${REGION} \
	--platform=managed \
	--max-instances=1 \
	--memory=128Mi \
	--service-account=${CLOUD_RUN_SERVICE_ACCOUNT_EMAIL} \
	--no-allow-unauthenticated \
	--set-env-vars \
	CONFIG=${CONFIG}

.PHONY: deploy-scheduler
deploy-scheduler: SCHEDULER_SERVICE_ACCOUNT_EMAIL ?=
deploy-scheduler:
	@gcloud scheduler jobs update http calendar-notifier \
	--http-method=post \
	--schedule="*/10 * * * *" \
	--uri="${SCHEDULED_ENDPOINT}/launch" \
	--oidc-service-account-email="${SCHEDULER_SERVICE_ACCOUNT_EMAIL}" \
	--time-zone="Asia/Tokyo"

.PHONY: run
run:
	@CONFIG=$(CONFIG) \
	docker-compose up -d
	docker logs -f calendar-notifier

.PHONY: install-tools
install-tools:
	cat tools.go | awk -F'"' '/_/ {print $$2}' | xargs -tI {} go install {}
