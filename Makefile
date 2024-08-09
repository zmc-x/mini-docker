build:
	go build -o ./bin/mini-docker ./main.go

clean:
	rm -rf ./bin ./logs