all: simple systemd

simple:
	go build -o bin/propush ./example/simple

systemd:
	go build -o bin/propushd ./example/systemd