/*
Implements a basic IRC bot using github.com/thoj/go-ircevent
*/
package main

import (
	"code.google.com/p/gosqlite/sqlite"
	"flag"
	"fmt"
	"github.com/thoj/go-ircevent"
	"strconv"
	"strings"
	"time"
)

type seenRecord struct {
	message string
	when    time.Time
}

var (
	con                                 *irc.Connection
	db                                  *sqlite.Conn
	host, nick, user, channel, database string
	verbose                             bool
	seenNicks                           = map[string]seenRecord{}
	startTime                           = time.Now()
)

func main() {

	flag.StringVar(&host, "host", "irc.freenode.net:6667", "The IRC host:port to connect to. Defaults to irc.freenode.net:6667")
	flag.StringVar(&nick, "nick", "minibot", "The IRC nick to use. Defaults to minibot")
	flag.StringVar(&user, "user", "minibot", "The IRC user to use. Defaults to minibot")
	flag.StringVar(&channel, "channel", "#minibot", "The IRC channel to join. Defaults to #minibot")
	flag.StringVar(&database, "database", "minibot.db", "The sqlite database file. Defaults to minibot.db")
	flag.BoolVar(&verbose, "verbose", false, "Be verbose")
	flag.Parse()

	debug("connecting to irc")
	con = irc.IRC(nick, user)

	debug("opening sqlite database")
	db, err := sqlite.Open(database)
	if err != nil {
		panic("An error occurred while opening the database: " + err.Error())
	}
	defer db.Close()

	debug("if needed, initializing sqlite database")
	err = db.Exec("create table if not exists messages (sender text, destination text, message text, primary key (sender, destination))")
	if err != nil {
		panic("Could not initialize database: " + err.Error())
	}

	debug("connecting to irc server")
	err = con.Connect(host)
	if err != nil {
		panic("An error occurred while connecting to irc: " + err.Error())
	}
	con.AddCallback("001", func(event *irc.Event) { con.Join(channel) })

	con.AddCallback("PRIVMSG", respond)

	debug("starting main loop")

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			con.Privmsg(channel, "An error ocurred and should have been logged")
		}
	}()
	con.Loop()

}

// helper function to print debug messages
func debug(message string) {
	if verbose {
		fmt.Println(message)
	}
}

// Function to respond to PRIVMSG events. 
func respond(event *irc.Event) {
	see(event.Nick, event.Message)
	op := strings.Split(event.Message, " ")[0]
	if op[0] == '!' {
		switch op {
		case "!ping":
			{
				reply(event, "pong")
			}
		case "!whoami":
			{
				reply(event, "You are "+event.Nick)
			}
		case "!help":
			{
				printHelp(event)
			}
		case "!countdown":
			{
				countdown(event)
			}
		case "!seen":
			{
				seen(event)
			}
		case "!uptime":
			{
				reply(event, "uptime: "+time.Since(startTime).String())
			}
		default:
			{
				reply(event, "Unknown command. Try !help")
			}
		}
	}
}

// sets/updates the last seen time and message for a nick
func see(nick string, message string) {
	seenNicks[nick] = seenRecord{message, time.Now()}
}

// checks if a nick has been seen
func seen(event *irc.Event) {
	args := strings.Split(event.Message, " ")
	if len(args) < 2 {
		reply(event, "You didn't specify a nick")
		return
	}
	nick := args[1]
	last, present := seenNicks[nick]
	if present == false {
		reply(event, "I have not seen "+nick+" since I've started ("+startTime.Format("2006-01-02 15:04 MST")+")")
	} else {
		reply(event, "I last saw "+nick+" on "+last.when.Format("2006-01-02 15:04 MST")+" and he said "+last.message)
	}
}

// helper function to send a reply to a user
func reply(event *irc.Event, message string) {
	con.Privmsg(event.Arguments[0], event.Nick+": "+message)
}

// sleeps for the specified amount of minutes, and then alerts the nick that invoked it, optionally printing a message
// TODO: I am not explicitly (AFAIK) creating a new goroutine here, yet this does not block the bot. I need to understand why that is the case. 
func countdown(event *irc.Event) {
	args := strings.Split(event.Message, " ")
	if len(args) < 2 {
		reply(event, "I need at least an amount in minutes to wait, and optionally a message to give you when the wait is over. ")
		return
	}
	dur, err := strconv.Atoi(args[1])
	if err != nil {
		reply(event, "My first argument has to be a number. I will be the number of minutes I will sleep before alerting you.")
		return
	}
	reply(event, "I will alert you in "+args[1]+" minutes")
	msg := "You asked me to alert you " + args[1] + " minutes ago"
	if len(args) > 2 {
		msg = ""
		for i := 2; i < len(args); i++ {
			msg += args[i] + " "
		}
		msg += " (" + args[1] + " minutes ago)"
	}
	time.Sleep(time.Duration(dur) * time.Minute)
	reply(event, msg)
}

// prints a line for each command, with a brief description.  
func printHelp(event *irc.Event) {
	helpItems := [...]string{
		"!ping : replies pong",
		"!whoami : replies 'you are <nick>'",
		"!countdown <i> [s]: sleeps i minutes and alerts you, optionally printing s",
		"!seen <nick>: tells you the last time <nick> was seen by the bot",
		"!uptime: prints the bot's uptime",
		"!help : prints basic help"}
	for _, v := range helpItems {
		reply(event, v)
	}

}
