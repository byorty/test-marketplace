migrate-up:
	docker run --rm \
		--network host \
		-v $(PWD)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations \
		-database="postgres://postgres:postgres@localhost:5432/marketplace?sslmode=disable" \
		up

migrate-down:
	docker run --rm \
		--network host \
		-v $(PWD)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations \
		-database="postgres://postgres:postgres@localhost:5432/marketplace?sslmode=disable" \
		down 1

generate-p:
	oapi-codegen \
		-config ./services/product-service/api/oapi-codegen.yaml \
		./services/product-service/api/product-service.yaml

generate-o:
	oapi-codegen \
		-config ./services/order-service/api/oapi-codegen.yaml \
		./services/order-service/api/order-service.yaml