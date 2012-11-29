all: 
	go build ircbot.go

run:
	./ircbot -host localhost:6667 -channel '#bottest' -verbose -nick 'martinbot'

clean: 
	rm -f ircbot 
