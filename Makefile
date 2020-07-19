all: | unit-test smoke-test run;

setup:
	docker-compose build

run: setup
	docker-compose up api

unit-test: ;

smoke-test: setup
	docker-compose run --rm test
