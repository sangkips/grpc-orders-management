run-orders:
	@go run services/order-service/*.go

run-hotel:
	@go run services/hotel/*.go

gen:
	@protoc \
		--proto_path=proto "orders/orders.proto" \
		--go_out=services/common/genproto/orders --go_opt=paths=source_relative \
  		--go-grpc_out=services/common/genproto/orders --go-grpc_opt=paths=source_relative

auth:
	@protoc \
		--proto_path=proto "auth/auth.proto" \
		--go_out=services/common/genproto/auth --go_opt=paths=source_relative \
  		--go-grpc_out=services/common/genproto/auth --go-grpc_opt=paths=source_relative