package bots

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/apex/log"
	handlers "github.com/osiloke/dobot/pkg/handlers/whatsapp"
	"github.com/osiloke/dostow-contrib/api"
)

// Answer stores an answer
type Answer struct {
	ID          string    `json:"id"`
	Action      string    `json:"action"`
	Answer      string    `json:"answer"`
	Created     time.Time `json:"created_at"`
	Modified    time.Time `json:"modified_at"`
	Question    *Question `json:"question"`
	Type        string    `json:"type"`
	MessageText string    `json:"messageText"`
	MessageID   string    `json:"messageID"`
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
			next := q.Next
			if next != nil {
				nextMessage := next.Answer
				if len(next.Question) > 0 {
					nextMessage = next.Question
					// data["question"] = next.ID
					// // data["questionText"] = q.Next.Question
					// // data["answerText"] = q.Next.Answer
					// d.api.Store.Create(d.answerStore, data)
				}
				err = d.MessageHandler.SendText(message, nextMessage)
			}
		}
		return err
	}
	d.api.Store.Create(d.answerStore, data)
	return d.MessageHandler.SendText(message, text)

}

// GetLastAnswer get the last answer a whatsapp user sent
func (d *DostowBot) GetLastAnswer(message *whatsapp.TextMessage) (*Answer, error) {
	log := log.WithField("phone", message.Info.RemoteJid).WithField("message", message.Text)
	q := answerQuery{Phone: message.Info.RemoteJid}
	log.WithField("q", q).Debug("Get last answer")
	raw, err := d.api.Store.Search(d.answerStore, api.QueryParams(&q, 2, 0))
	if err == nil {
		var qs answerResult
		if err = json.Unmarshal(*raw, &qs); err == nil {
			if len(qs.Data) > 0 {
				answer := qs.Data[0]
				log.WithField("answer", answer).Debug("Found answer")
				return answer, nil
			}
		} else {
			log.WithError(err).Error("unable to unmarshal result")
		}
	} else {
		log.WithError(err).Error("error while retrieving answers")
	}
	return nil, errors.New("no answer found")
}

// HandleTextMessage see MessageHandler.HandleTextMessage
func (d *DostowBot) HandleTextMessage(message whatsapp.TextMessage) {
	log := log.WithField("whatsappID", message.Info.RemoteJid).WithField("messageID", message.Info.Id)
	if message.Info.FromMe || message.Info.Timestamp < d.MessageHandler.StartTime() {
		return
	}
	lastAnswer, _ := d.GetLastAnswer(&message)
	if lastAnswer != nil && len(lastAnswer.MessageID) > 0 {
		if message.Info.Id == lastAnswer.MessageID {
			return
		}
		messageTime := time.Unix(int64(message.Info.Timestamp), 0)
		log.Debugf("last answer created at %s, message time %s ", lastAnswer.Created, messageTime)
		if lastAnswer.Created.After(messageTime) {
			log.Debugf("last answer is old and was created at %s, message time %s ", lastAnswer.Created, messageTime)
			return
		}
	}
	var err error
	if !strings.Contains(strings.ToLower(message.Text), "help") {
		if lastAnswer != nil && lastAnswer.Question != nil && lastAnswer.Question.Next != nil {
			tplName := fmt.Sprintf("%s", lastAnswer.Question.ID)
			var data map[string]interface{}
			if len(lastAnswer.Question.Next.Action) > 0 {
				tplName = fmt.Sprintf("%s", lastAnswer.Question.Next.Action)
				// TODO: lastAnswer.Question.validate(message.Text)
				var resp interface{}
				log.WithField("lastAnswer", lastAnswer).Debugf("last answer has a question action for %s", message.Text)
				resp, err = d.TriggerAction(lastAnswer.Question.Next.Action, &message, lastAnswer)
				if err == nil {
					if raw, ok := resp.(*json.RawMessage); ok {
						err = json.Unmarshal(*raw, &data)
					}
					log.WithField("data", data).Debug("action triggered")
				}
				if err != nil {
					log.WithError(err).Errorf("Cannot trigger action - %s", lastAnswer.Question.Action)
					d.SendAnswer(nil, &message, "😦 I dont understand, please send help")
					return
				}
			} else {
				data = map[string]interface{}{}
			}
			if len(lastAnswer.Question.Next.Answer) > 0 {
				var tmpl *template.Template
				log.Debugf("Create template - %s - %v", tplName, lastAnswer.Question.Next.Answer)
				tmpl, err = template.New(tplName).Parse(lastAnswer.Question.Next.Answer)
				if err == nil {
					var tmplBytes bytes.Buffer
					err = tmpl.Execute(&tmplBytes, map[string]interface{}{"data": data, "message": message.Text})
					if err == nil {
						tplstring := tmplBytes.String()
						lined := strings.Split(tplstring, "\\n")
						linedup := ``
						for _, v := range lined {
							linedup = fmt.Sprintf(`%s
%s`, linedup, v)
						}
						err = d.SendAnswer(lastAnswer.Question.Next, &message, linedup)
					} else {
						log.WithError(err).Error("error executing template")
					}
				} else {
					log.WithError(err).Error("error parsing template")
				}
			} else {
				err = d.SendAnswer(lastAnswer.Question.Next, &message, "Done!")
			}
			if err != nil {
				d.SendText(&message, "I can't answer your question right now. Please try again later.")
			}
			return
		}
		log.Debugf("search for a root question containing %s", message.Text)
		raw, err := d.api.Store.Search(d.questionStore, api.QueryParams(&questionQuery{Question: message.Text, Type: "root"}, 100, 0))
		if err == nil {
			var qs questionResult
			if err = json.Unmarshal(*raw, &qs); err == nil {
				if len(qs.Data) == 1 {
					log.Debugf("found a root")
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
	}
	raw, _, err := d.api.Store.List(d.questionStore, api.QueryParams(&questionQuery{Type: "root"}, 100, 0))
	if err == nil {
		var qs questionResult
		err = json.Unmarshal(*raw, &qs)
		if err == nil {
			// d.MessageHandler.SendText(&message, "😕 Hey! i don't understand")
			msg := `👨 Try asking me one of the questions below 👇

`
			for _, q := range qs.Data {
				msg = msg + q.Question + `
`
			}
			err = d.SendAnswer(nil, &message, msg)
		}
	}
	if err != nil {
		log.WithError(err).Error("error sending message")
		d.SendText(&message, "I can't answer your question right now. Please try again later.")
	}
	return
}
