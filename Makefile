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

generate:
	oapi-codegen \
		-config ./services/product-service/api/oapi-codegen.yaml \
		./services/product-service/api/product-service.yaml