package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/Asladck/WikiBot/api"
	key "github.com/Asladck/WikiBot/safeStuff"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

type User struct {
	User_ID     string
	Username    string
	Email       string
	Recent_Data string
	Saved_Data  string
	Language    string
}

type Bot struct {
	Session *discordgo.Session
	DB      *pgx.Conn
}

func NewBot(db *pgx.Conn) *Bot {
	session, err := discordgo.New("Bot " + key.TokenTaker())
	if err != nil {
		log.Panic("New Bot error: ", err)
	}
	return &Bot{
		Session: session,
		DB:      db,
	}
}

func (b *Bot) Start() {
	b.Session.AddHandler(b.BotSaver)
	b.Session.AddHandler(b.BotRecent)
	b.Session.AddHandler(b.BotShow)
	b.Session.AddHandler(b.BotLanguage)
	b.Session.AddHandler(b.BotSearch)
	if err := b.Session.Open(); err != nil {
		log.Panic("BotHandler error:", err)
	}
	log.Println("Бот запущен и слушает события.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	b.Session.Close()
}

var user User

func (b *Bot) BotSearch(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "!search" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Напишите что нужно искать или напишите !stop")
		if err != nil {
			log.Println("!search проблема:", err)
			return
		}
		b.Session.AddHandlerOnce(func(s *discordgo.Session, nextmsg *discordgo.MessageCreate) {
			if nextmsg.Author.ID == s.State.User.ID {
				return
			}
			if nextmsg.Content == "!stop" {
				s.ChannelMessageSend(m.ChannelID, "Поиск отменен")
				return
			}
			lang := b.GetUserLang(nextmsg.Author.ID)
			re := regexp.MustCompile(`\s+`)
			nextmsg.Content = re.ReplaceAllString(nextmsg.Content, "-")
			log.Println(nextmsg.Content)
			information := api.Search(key.SearchTaker(lang), nextmsg.Content)
			_, err := b.DB.Exec(context.Background(), "UPDATE users SET search_data = $1 WHERE discord_id = $2",
				information, nextmsg.Author.ID)
			if err != nil {
				log.Println("Ошибка в обновлении search_data")
			}
			s.ChannelMessageSend(m.ChannelID, "Твоя информация:\n")
			s.ChannelMessageSend(m.ChannelID, information)
		})
	}
}
func (b *Bot) BotShow(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "!show" {
		var savedData string
		err := b.DB.QueryRow(context.Background(), "SELECT saved_data FROM users WHERE discord_id = $1", m.Author.ID).Scan(&savedData)
		if err != nil {
			log.Panic("BotShow QueryRow error:", err)
		}
		s.ChannelMessageSend(m.ChannelID, "Ваша сохранённая информация\n")
		log.Println(savedData)
		s.ChannelMessageSend(m.ChannelID, savedData)

	}
}

func (b *Bot) BotSaver(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	var dataType string
	switch m.Content {
	case "!save":
		dataType = "recent_data"
	case "!search save":
		dataType = "search_data"
	default:
		return
	}
	var exists bool
	exists = false
	err := b.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE discord_id = $1)",
		m.Author.ID).Scan(&exists)
	if !exists {
		log.Println("Не существует")
		_, err = b.DB.Exec(
			context.Background(),
			"INSERT INTO users (discord_id, username, recent_data, saved_data, language,search_data) VALUES ($1, $2, $3, $4, $5, $6)",
			m.Author.ID, m.Author.Username, user.Recent_Data, user.Saved_Data, b.GetUserLang(m.Author.ID), "",
		)
		if err != nil {
			log.Println("Ошибка при сохранении пользователя:", err)
			s.ChannelMessageSend(m.ChannelID, "Ошибка при сохранении пользователя!")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Данные сохранены!")
	}
	var dataToSave string
	err = b.DB.QueryRow(context.Background(),
		fmt.Sprintf("SELECT %s FROM users WHERE discord_id = $1", dataType), m.Author.ID).Scan(&dataToSave)
	if err != nil {
		if err == sql.ErrNoRows {
			s.ChannelMessageSend(m.ChannelID, "Нет данных для сохранения.")
		} else {
			log.Println("Ошибка при получении данных:", err)
			s.ChannelMessageSend(m.ChannelID, "Ошибка при обработке данных!")
		}
		return
	}
	_, err = b.DB.Exec(context.Background(),
		"UPDATE users SET saved_data = $1 WHERE discord_id = $2",
		dataToSave, m.Author.ID)
	if err != nil {
		log.Println("Ошибка при обновлении saved_data:", err)
		s.ChannelMessageSend(m.ChannelID, "Ошибка при сохранении данных!")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Данные успешно сохранены!")
}

func (b *Bot) BotRecent(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content != "!recent" {
		return
	}

	lang := b.GetUserLang(m.Author.ID)
	url := key.UrlTaker(lang)
	data := api.TakeQuery(url)
	_, err := b.DB.Exec(context.Background(), "UPDATE users SET recent_data = $1 WHERE discord_id = $2",
		data, m.Author.ID)
	if err != nil {
		log.Println("Ошибка в обновлении recent_data")
	}
	msg := "Your current information on WIKI\n"
	if lang == "rus" {
		msg = "Ваша новая информация\n"
	}
	s.ChannelMessageSend(m.ChannelID, msg+data)
}

func (b *Bot) BotLanguage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	var newLang string

	switch m.Content {
	case "!eng":
		newLang = "eng"
	case "!rus":
		newLang = "rus"
	default:
		return
	}

	_, err := b.DB.Exec(
		context.Background(),
		`UPDATE users SET language = $1 WHERE discord_id = $2`,
		newLang, m.Author.ID,
	)
	if err != nil {
		log.Println("Ошибка при изменении языка:", err)
		s.ChannelMessageSend(m.ChannelID, "Ошибка при изменении языка!")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Язык сменён на "+newLang)
}

func (b *Bot) GetUserLang(userID string) string {
	var lang string
	err := b.DB.QueryRow(
		context.Background(),
		`SELECT language FROM users WHERE discord_id = $1`,
		userID,
	).Scan(&lang)

	if err != nil {
		log.Println("Ошибка при получении языка, устанавливаю English по умолчанию:", err)
		return "eng"
	}

	return lang
}
