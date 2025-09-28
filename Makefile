build:
	go build -o bin/ddns-updater ./cmd/main.go

init_config:
	mkdir -p ~/.config/ddns-updater/
	cp ./config-example.json ~/.config/ddns-updater/config.json
