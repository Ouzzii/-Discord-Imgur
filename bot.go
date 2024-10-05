package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"math/rand"
	"net/http"
	"github.com/bwmarrin/discordgo"
	"github.com/iunary/fakeuseragent"
	"github.com/joho/godotenv"
)

var running = false
var components = []discordgo.MessageComponent{
	discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "⭐",
				Style:    discordgo.SuccessButton,
				CustomID: "favorite",
			},
			discordgo.Button{
				Label:    "🗑️",
				Style:    discordgo.DangerButton,
				CustomID: "delete",
			},
		},
	},
}

type Imgur struct {
	id           string
	apiUrl       string
	imageUrl     string
	imageContent []byte
}

func checkUrl(url string) (bool, []byte) {
	randomAgent := fakeuseragent.RandomUserAgent()
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{"User-Agent": {randomAgent}}

	res, _ := client.Do(req)
	body, _ := io.ReadAll(res.Body)

	if res.Request.URL.String() != "https://i.imgur.com/removed.png" {

		return true, body
	} else {
		return false, body
	}

}

func checkChannel(channelName string) {
	//var channel discordgo.Channel
	//channel.ID
	//discordgo.NewState().ChannelAdd(channelName)
}

func createImgur() Imgur {
	const letters string = "abcdefghijklmnoprstuvwxyzABCDEFGHIJKLMNOPRSTUVWXYZ1234567890"
	var imgur Imgur
	for true {
		counter := 0
		imgur.id = ""
		for counter < 5 {
			imgur.id = imgur.id + string(letters[rand.Intn(len(letters))])
			counter++
		}
		imgur.apiUrl = "https://api.imgur.com/3/image/" + imgur.id
		imgur.imageUrl = "https://i.imgur.com/" + imgur.id + ".png"
		checked, content := checkUrl(imgur.imageUrl)
		if checked {
			imgur.imageContent = content
			break
		}
	}
	return imgur
}

func main() {

	err := godotenv.Load(".env")
	if err != nil{
		panic(".env Dosyasi yuklenirken hata ile karsilasildi..")
	}
	


	token := os.Getenv("TOKEN")
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Bot oluşturulamadı:", err)
		return
	}

	// Botun olayları dinlemeye başlaması
	bot.AddHandler(messageCreate)
	bot.AddHandler(onButtonClicked)
	err = bot.Open()
	if err != nil {
		fmt.Println("Bot başlatılamadı:", err)
		return
	}

	fmt.Println("Bot çalışıyor. CTRL+C ile çıkabilirsiniz.")
	select {}
}

func onButtonClicked(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Buton etkileşimlerini kontrol et
	if i.Type == discordgo.InteractionMessageComponent {
		switch i.MessageComponentData().CustomID {
		case "favorite":
			guildID := i.GuildID
			var categoryID string
			var channelexists bool = false
			var channelName string = i.Member.User.ID + "-" + i.Member.User.Username
			var channelID string
			if guildID == "" {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Bu komut sadece sunucularda çalışır.",
					},
				})
				return
			}

			// Sunucudaki tüm kanalları al
			channels, _ := s.GuildChannels(guildID)
			for _, channel := range channels {

				if channel.Name == channelName {
					channelexists = true
					channelID = channel.ID
				}

			}
			if !channelexists {
				for _, channel := range channels {
					if channel.Type == discordgo.ChannelTypeGuildCategory && channel.Name == "favoriler" {
						categoryID = channel.ID
						break

					}
				}
				if categoryID != "" {
					newchannel, _ := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
						Name:     channelName,
						Type:     discordgo.ChannelTypeGuildText,
						ParentID: categoryID,
					})
					channelID = newchannel.ID
				}
			}

			if i.Message.Content != "" {
				s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{Content: i.Message.Content, Components: components})
			} else {
				s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{Content: i.Message.Attachments[0].URL, Components: components})
			}

		case "delete":
			s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Mesaj oluşturma olayları burada işlenecek

	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
	if m.Content == "start" {
		running = true
		for running {
			imgur := createImgur()
			s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Components: components,
				Files: []*discordgo.File{
					{
						Name:        imgur.id + ".png",
						ContentType: "image/png",
						Reader:      bytes.NewReader(imgur.imageContent),
					},
				},
			})
		}
	}
	if m.Content == "stop" {
		running = false
	}
}
