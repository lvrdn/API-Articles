run:
	go run cmd/main.go

db:
	docker run --name mypostgr -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=1234 -e POSTGRES_DB=realworld -d postgres 

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