//go:generate protoc --go_out=. --go-grpc_out=. --go-grpc_opt=paths=source_relative --go-grpc_opt=Mchecker.proto=./authz  --go_opt=paths=source_relative --go_opt=Mchecker.proto=./authz  checker.proto

package authz
