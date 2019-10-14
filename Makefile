default: build

docker_image_name="local/nginxlab"

build:
	docker build -t $(docker_image_name) .

up:
	docker-compose -f docker-compose.yml up -d

clean:
	docker-compose down

re: clean build up
