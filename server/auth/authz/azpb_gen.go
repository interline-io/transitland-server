//go:generate protoc --plugin=protoc-gen-go=../../../protoc-gen-go-wrapper.sh --go_out=. --go-grpc_out=. --go-grpc_opt=paths=source_relative --go-grpc_opt=Mazpb.proto=./authz  --go_opt=paths=source_relative --go_opt=Mazpb.proto=./authz  azpb.proto
//                      --plugin=protoc-gen-go=../../protoc-gen-go-wrapper.sh --go_out=. --go_opt=paths=source_relative --go_opt=Mgtfs-realtime.proto=rt/pb  gtfs-realtime.proto

package authz
