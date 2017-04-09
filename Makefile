run: build
	bash -c "source .env; ./hako"

build:
	go build -i -v -o hako .
