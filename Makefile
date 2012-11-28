all: 
	go build ircbot.go

run:
	./ircbot -nick minibottest -user minibottest -verbose

clean: 
	rm -f ircbot 
