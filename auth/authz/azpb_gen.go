//go:generate protoc --go_out=. --go-grpc_out=. --go-grpc_opt=paths=source_relative --go-grpc_opt=Mazpb.proto=./authz  --go_opt=paths=source_relative --go_opt=Mazpb.proto=./authz  azpb.proto

package authz
