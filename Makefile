.PHONY: start stop up down itest

start:
	docker-compose start

stop:
	docker-compose stop

up:
	docker-compose up -d

down:
	docker-compose down

itest:
	go run test/test.go
