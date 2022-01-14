/*
Copyright © 2020 Osiloke Harold Emoekpere <me@osiloke.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ws "github.com/Rhymen/go-whatsapp"
	"github.com/apex/log"
	"github.com/osiloke/dobot/pkg/bots"
	handlers "github.com/osiloke/dobot/pkg/handlers/whatsapp"
	"github.com/osiloke/dostow-contrib/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// botCmd represents the bot command
var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Start a bot",
	Long: `Start a bot.
	This will launch a whatsapp bot by default`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.DebugLevel)

		filepath := viper.GetString("session.path")
		apiURL := viper.GetString("dostow.url")
		apiKey := viper.GetString("dostow.key")
		apiToken := viper.GetString("dostow.token")

		log.WithField("apiURL", apiURL).Debug("launching whatsapp bot with dostow")
		wac, err := ws.NewConn(20 * time.Second)
		if err != nil {
			panic(err)
		}
		// wac.SetClientVersion(0, 4, 2080)
		a := api.NewClientWithUser(apiURL, apiKey, apiToken, http.DefaultClient)
		actionStore := bots.NewDostowActionStore(a)
		actions := bots.Actions{}
		v, _ := ioutil.ReadFile("./actions.json")
		err = json.Unmarshal(v, &actions)
		if err != nil {
			panic(err)
		}
		handler := handlers.NewMessageHandler(wac)
		bot := bots.NewDostowBot(a, "question", "answer", actions, actionStore, handler)
		//Add handler
		wac.AddHandler(bot)
		// TODO: this should be a method in the handler
		err = handlers.Login(wac, filepath)
		if err != nil {
			log.WithError(err).Error("failed")
		}
		//verifies phone connectivity
		pong, err := wac.AdminTest()

		if !pong || err != nil {
			log.Fatalf("error pinging in: %v\n", err)
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		//Disconnect safe
		fmt.Println("Shutting down now.")
		session, err := wac.Disconnect()
		if err != nil {
			log.Fatalf("error disconnecting: %v\n", err)
		}
		// TODO: this should be a method in the handler
		if err := handlers.SaveSession(session, filepath); err != nil {
			log.Fatalf("error saving session: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(botCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// botCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	botCmd.Flags().StringP("api-url", "a", "https://api.dostow.com/v1/", "Dostow api URL")
	botCmd.Flags().StringP("api-key", "k", "", "Dostow api key")
	botCmd.Flags().StringP("api-token", "t", "", "Dostow api token")
	botCmd.Flags().StringP("session-path", "s", "./session.gob", "Path to session")
	viper.BindPFlag("dostow.url", botCmd.Flags().Lookup("api-url"))
	viper.BindPFlag("dostow.key", botCmd.Flags().Lookup("api-key"))
	viper.BindPFlag("dostow.token", botCmd.Flags().Lookup("api-token"))
	viper.BindPFlag("sessions.path", botCmd.Flags().Lookup("session-path"))
}
