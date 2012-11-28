all: 
	go build ircbot.go

run:
	./ircbot -host localhost:6667 -channel '#bottest' -verbose

clean: 
	rm -f ircbot 
