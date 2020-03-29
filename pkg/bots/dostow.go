package bots

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/apex/log"
	handlers "github.com/osiloke/dobot/pkg/handlers/whatsapp"
	"github.com/osiloke/dostow-contrib/api"
)

// Answer stores an answer
type Answer struct {
	ID       string    `json:"id"`
	Action   string    `json:"action"`
	Answer   string    `json:"answer"`
	Created  time.Time `json:"created_at"`
	Modified time.Time `json:"modified_at"`
	Question *Question `json:"question"`
	Type     string    `json:"type"`
}

// Question stores a question
type Question struct {
	ID       string    `json:"id"`
	Action   string    `json:"action"`
	Answer   string    `json:"answer"`
	Created  time.Time `json:"created_at"`
	Modified time.Time `json:"modified_at"`
	Next     *Question `json:"next"`
	Question string    `json:"question"`
	Type     string    `json:"type"`
}

type questionResult struct {
	Data       []*Question
	TotalCount int `json:"total_count"`
}

type questionQuery struct {
	Action   string `json:"action,omitempty"`
	Answer   string `json:"answer,omitempty"`
	Question string `json:"question,omitempty"`
	Type     string `json:"type,omitempty"`
}

type answerResult struct {
	Data       []*Answer
	TotalCount int `json:"total_count"`
}

type answerQuery struct {
	Action   string `json:"action,omitempty"`
	Answer   string `json:"answer,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Question string `json:"question,omitempty"`
	Type     string `json:"type,omitempty"`
}

// DostowBot a bot that uses dostow
type DostowBot struct {
	api           *api.Client
	questionStore string
	answerStore   string
	actions       Actions
	actionStore   ActionStore
	*handlers.MessageHandler
}

// NewDostowBot create a dostow bot with actions and a message handler
func NewDostowBot(a *api.Client, questionStore, answerStore string,
	actions Actions, actionStore ActionStore, handler *handlers.MessageHandler) *DostowBot {
	return &DostowBot{a, questionStore, answerStore, actions, actionStore, handler}
}

// MessageToMap convert message to map
func (d *DostowBot) MessageToMap(message *whatsapp.TextMessage) map[string]interface{} {
	phone := message.Info.RemoteJid
	senderJid := message.Info.SenderJid
	messageID := message.Info.Id
	participant := message.ContextInfo.Participant
	pushName := message.Info.PushName
	messageText := message.Text
	return map[string]interface{}{
		// "answer":      text,
		"messageText": messageText, "phone": phone,
		"senderJID": senderJid, "messageID": messageID,
		"participant": participant, "pushName": pushName,
	}
}

// TriggerAction trigger an action
func (d *DostowBot) TriggerAction(action string, message *whatsapp.TextMessage, answer *Answer) (interface{}, error) {
	return DoAction(action, map[string]interface{}{"message": d.MessageToMap(message), "answer": answer}, d.actions, d.actionStore)
}

// SendAnswer send a text message
func (d *DostowBot) SendAnswer(q *Question, message *whatsapp.TextMessage, text string) error {
	log.WithField("message", message).Debug("Send Answer")
	data := d.MessageToMap(message)
	if q != nil {
		data["question"] = q.ID
		// data["questionText"] = q.Question
		// data["answerText"] = q.Answer
		d.api.Store.Create(d.answerStore, data)
		err := d.MessageHandler.SendText(message, text)
		if q.Next != nil {
			nextMessage := q.Next.Answer
			if len(q.Next.Question) > 0 {
				nextMessage = q.Next.Question
				data["question"] = q.Next.ID
				// data["questionText"] = q.Next.Question
				// data["answerText"] = q.Next.Answer
				d.api.Store.Create(d.answerStore, data)
			}
			err = d.MessageHandler.SendText(message, nextMessage)
		}
		return err
	}
	d.api.Store.Create(d.answerStore, data)
	return d.MessageHandler.SendText(message, text)

}

// GetLastAnswer get the last answer a whatsapp user sent
func (d *DostowBot) GetLastAnswer(message *whatsapp.TextMessage) (*Answer, error) {
	raw, err := d.api.Store.Search(d.answerStore, api.QueryParams(&answerQuery{Phone: message.Info.RemoteJid}, 2, 0))
	if err == nil {
		var qs answerResult
		if err = json.Unmarshal(*raw, &qs); err == nil {
			if len(qs.Data) > 0 {
				answer := qs.Data[0]
				return answer, nil
			}
		}
	}
	return nil, errors.New("no answer found")
}

// HandleTextMessage see MessageHandler.HandleTextMessage
func (d *DostowBot) HandleTextMessage(message whatsapp.TextMessage) {
	if message.Info.FromMe || message.Info.Timestamp < d.MessageHandler.StartTime() {
		return
	}
	lastAnswer, _ := d.GetLastAnswer(&message)
	raw, err := d.api.Store.Search(d.questionStore, api.QueryParams(&questionQuery{Question: message.Text, Type: "root"}, 100, 0))
	if err == nil {
		var qs questionResult
		if err = json.Unmarshal(*raw, &qs); err == nil {
			if len(qs.Data) == 1 {
				msg := qs.Data[0].Answer
				err = d.SendAnswer(qs.Data[0], &message, msg)
				if err != nil {
					log.WithError(err).Error("error sending message")
					d.SendText(&message, "I can't answer your question right now. Please try again later.")
				}
				return
			}
		}
	}
	if lastAnswer != nil && lastAnswer.Question != nil {
		// TOdo: lastAnswer.Question.validate(message.Text)
		if len(lastAnswer.Question.Action) > 0 {
			_, err = d.TriggerAction(lastAnswer.Question.Action, &message, lastAnswer)
			if err == nil {
				// d.MessageHandler.SendText(message, lastAnswer.Question.Answer)
				err = d.SendAnswer(lastAnswer.Question, &message, lastAnswer.Question.Answer)
				if err != nil {
					log.WithError(err).Error("error sending message")
					d.SendText(&message, "I can't answer your question right now. Please try again later.")
				}
				return
			}
			log.WithError(err).Errorf("Cannot trigger action - %s")
			d.MessageHandler.SendText(&message, "😦 Sorry about this but i cannot do this")

		}
	} else {
		raw, _, err = d.api.Store.List(d.questionStore, api.QueryParams(&questionQuery{Type: "root"}, 100, 0))
		if err == nil {
			var qs questionResult
			err = json.Unmarshal(*raw, &qs)
			if err == nil {
				d.MessageHandler.SendText(&message, "😕 Hey! i don't understand")
				msg := `👨 Try asking me one of the questions below 👇

`
				for _, q := range qs.Data {
					msg = msg + q.Question + `
`
				}
				err = d.SendAnswer(nil, &message, msg)
			}
		}
	}
	if err != nil {
		log.WithError(err).Error("error sending message")
		d.SendText(&message, "I can't answer your question right now. Please try again later.")
	}
	return
}
