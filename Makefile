ENV_FILE ?= .env

-include $(ENV_FILE)
export

migrate-up:
	docker run --rm \
		--network host \
		-v $(PWD)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations \
		-database="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		up

migrate-down:
	docker run --rm \
		--network host \
		-v $(PWD)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations \
		-database="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		down 1

generate-p:
	oapi-codegen \
		-config ./services/product-service/api/oapi-codegen.yaml \
		./services/product-service/api/product-service.yaml

generate-o:
	oapi-codegen \
		-config ./services/order-service/api/oapi-codegen.yaml \
		./services/order-service/api/order-service.yaml

generate-cl:
	oapi-codegen \
		-config ./services/common/client/product/oapi-codegen/product-client.yaml \
		./services/product-service/api/product-service.yaml