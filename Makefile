.PHONY: start stop up down

start:
	docker-compose start

stop:
	docker-compose stop

up:
	docker-compose up -d

down:
	docker-compose down
