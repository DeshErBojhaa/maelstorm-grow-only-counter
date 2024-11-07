build:
	go build -o ./bin/main main.go

run: build
	../ms/maelstrom test -w g-counter --bin ./bin/main --node-count 3 --rate 100 --time-limit 20 --nemesis partition