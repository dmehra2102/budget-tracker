.PHONY: build run logs docker-up docker-down clean

build:
		go build -o bin/api cmd/api/main.go
		go build -o bin/worker cmd/worker/main.go

run:
		go run cmd/api/main.go

run-worker:
		go run cmd/worker/main.go

docker-up:
		docker-compose up -d

docker-down:
		docker-compose down

clean:
	rm -rf bin/
	docker-compose down -v

logs:
	docker-compose logs -f