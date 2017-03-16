run: build
	bash -c "source .env; ./hako"

build:
	go build -o hako .
