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