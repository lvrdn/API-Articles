run:
	go run cmd/main.go

db_connect:
	psql -hlocalhost -p5432 -Uroot -drealworld 

build:
	go build \
		-o ./bin/app \
		./cmd/

containers:
	docker compose up

app_on:
	docker compose start

app_off:
	docker compose stop

app_logs:
	docker logs -f rwa-rwa-1