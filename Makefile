.EXPORT_ALL_VARIABLES:
YANDEX_IAM_TOKEN := $(shell yc iam create-token)

run: 
	@go run .
test:
	@echo YANDEX_IAM_TOKEN=${YANDEX_IAM_TOKEN} YANDEX_ORG_ID=${YANDEX_ORG_ID}



check:
	@curl -X GET \
  	-H "Authorization: Bearer ${YANDEX_IAM_TOKEN}" \
  	https://resource-manager.api.cloud.yandex.net/resource-manager/v1/clouds

token-echo:
	@echo ${YANDEX_IAM_TOKEN}

worklog:
	GET /v2/worklog?createdBy=<имя_или_идентификатор_пользователя>&createdAt=from:<начало>&createdAt=to:<окончание>
	Host: api.tracker.yandex.net
	Authorization: OAuth <OAuth-токен>
	X-Org-ID или X-Cloud-Org-ID: <идентификатор_организации>

myself:
	curl --request GET "api.tracker.yandex.net/v2/myself" \
     --header "Authorization: OAuth ${YANDEX_IAM_TOKEN}" \
     --header "X-Cloud-Org-Id: bpfk3docj86i5k1qhp58"
