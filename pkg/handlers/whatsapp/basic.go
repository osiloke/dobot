package whatsapp

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
	"github.com/apex/log"
)

// Login to whatsapp
func Login(wac *whatsapp.Conn, filepath string) error {
	session, err := readSession(filepath)
	if err == nil {
		session, err = wac.RestoreWithSession(session)

		if err != nil {
			fmt.Fprintf(os.Stderr, "restoring failed: %v\n", err)
			//no saved session -> regular login
			qr := make(chan string)
			go func() {
				terminal := qrcodeTerminal.New()
				terminal.Get(<-qr).Print()
			}()
			session, err = wac.Login(qr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error during login: %v\n", err)
				return err
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "unable to read session: %v\n", err)
		qr := make(chan string)

		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()

		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error login after failed read: %v", err)
		}
	}

	if err = writeSession(session, filepath); err != nil {
		return fmt.Errorf("error saving session: %v", err)
	}

	return nil
}

// ReadSession loads a session from a file
func readSession(filepath string) (whatsapp.Session, error) {
	session := whatsapp.Session{}

	file, err := os.Open(filepath)
	if err != nil {
		return session, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err = decoder.Decode(&session); err != nil {
		return session, err
	}

	return session, nil
}

// WriteSession writes a session to a filepath
func writeSession(session whatsapp.Session, filepath string) error {
	log.Infof("writing session to %s", filepath)
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err = encoder.Encode(session); err != nil {
		return err
	}

	return nil
}

// SaveSession writes a session to a filepath
func SaveSession(session whatsapp.Session, filepath string) error {
	return writeSession(session, filepath)
}

// NewMessageHandler creates a new Message Handler
func NewMessageHandler(wac *whatsapp.Conn) *MessageHandler {
	return &MessageHandler{wac, uint64(time.Now().Unix())}
}

// MessageHandler manages messages sent from a whatsapp connection
type MessageHandler struct {
	wac       *whatsapp.Conn
	startTime uint64
}

// HandleError log error
func (mh *MessageHandler) HandleError(err error) {
	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Debugf("Connection failed, underlying error: %v", e.Err)
		log.Debug("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Debug("Reconnecting...")
		err := mh.wac.Restore()
		if err != nil {
			// log.Fatalf("Restore failed: %v", err)
			panic(err)
		}
	} else {
		log.Debugf("error occoured: %v\n", err)
	}
}

// SendText send a text message
func (mh *MessageHandler) SendText(message *whatsapp.TextMessage, text string) error {
	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: message.Info.RemoteJid,
		},
		Text: text,
	}
	_, err := mh.wac.Send(msg)
	return err
}

// HandleTextMessage receives whatsapp text messages and checks if the message was send by the current
// user, if it does not contain the keyword '@echo' or if it is from before the program start and then returns.
// Otherwise the message is echoed back to the original author.
func (mh *MessageHandler) HandleTextMessage(message whatsapp.TextMessage) {
	if message.Info.FromMe || message.Info.Timestamp < mh.startTime {
		return
	}

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: message.Info.RemoteJid,
		},
		Text: message.Text,
	}

	// send message back
	if _, err := mh.wac.Send(msg); err != nil {
		log.WithError(err).Error("error sending message")
	}

	log.Debugf("echoed message '%v' to user %v\n", message.Text, message.Info.RemoteJid)
}

// StartTime get start time
func (mh *MessageHandler) StartTime() uint64 {
	return mh.startTime
}
