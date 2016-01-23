package slack

import (
	"bytes"
	"log"
	"strings"

	"github.com/nlopes/slack"
)

type Slack struct {
	token  string // slack token
	rtm    *slack.RTM
	botUID string

	// singleton channel name
	channel string
}

func NewClient(token string) *Slack {
	if len(token) < 1 {
		panic("supbot: can't seem to start myself")
	}
	api := slack.New(token)

	s := &Slack{token: token, rtm: api.NewRTM()}
	go s.rtm.ManageConnection()

	return s
}

func (s *Slack) wasMentioned(msg string) bool {
	if len(msg) < 1 {
		return false
	}
	// NOTE: must be prefixed
	return strings.HasPrefix(msg, s.botUID)
}

// expect some byte and write to slack
func (s *Slack) Write(o []byte) (n int, err error) {
	outBuf := bytes.Buffer{}
	outBuf.Write(o)
	outBuf.WriteString("\n")

	s.rtm.SendMessage(
		s.rtm.NewOutgoingMessage(
			outBuf.String(),
			s.channel,
		),
	)
	return len(o), nil
}

func (s *Slack) InitializeRTM() {
Loop:
	for {
		select {
		case msg := <-s.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				log.Println("slackbot: hello dave.")
			case *slack.ConnectedEvent:
				log.Println("slackbot: I'm online dave.")
				for _, ch := range ev.Info.Channels {
					log.Printf("slackbot: joined channel %s\n", ch.Name)
				}
			case *slack.MessageEvent:
				s.channel = ev.Msg.Channel
				// must be mentioned
				if s.wasMentioned(ev.Text) {
					// for now, printlin
					log.Println(ev.Text)
				}
			case *slack.InvalidAuthEvent:
				log.Println("supbot: I seem to be disconnected, can't let you do that.")
				break Loop
			}
		}
	}
}
