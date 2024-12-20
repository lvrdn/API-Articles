run:
	go run cmd/main.go

db:
	docker run --name mypost1 --hostname localhost -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=1234 -e POSTGRES_DB=realworld -d postgres

connect:
	psql -hlocalhost -p5432 -Uroot -drealworld 
