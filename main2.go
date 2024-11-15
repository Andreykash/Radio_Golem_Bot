package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "math/rand"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
    TelegramToken       string `json:"telegram_token"`
    MusicLibraryPath    string `json:"music_library_path"`
    DirectoriesFilePath string `json:"directories_file_path"`
}

func loadConfig(filename string) (Config, error) {
    var config Config
    file, err := os.Open(filename)
    if err != nil {
        return config, err
    }
    defer file.Close()
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&config)
    return config, err
}

func loadDirectories(filename string) ([]string, error) {
    var directories []string
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("ошибка открытия файла %s: %v", filename, err)
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        directories = append(directories, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("ошибка чтения файла %s: %v", filename, err)
    }
    return directories, nil
}

func getRandomAlbum(directories []string) (string, []string, error) {
    if len(directories) == 0 {
        return "", nil, fmt.Errorf("не найдено ни одного альбома")
    }

    rand.Seed(time.Now().UnixNano())

    for {
        randomDir := directories[rand.Intn(len(directories))]
        log.Printf("Выбрана случайная директория: %s", randomDir)

        albumFiles, err := os.ReadDir(randomDir)
        if err != nil {
            return "", nil, err
        }

        var albumCover string
        var albumTracks []string

        for _, file := range albumFiles {
            if strings.HasSuffix(file.Name(), ".jpg") {
                albumCover = filepath.Join(randomDir, file.Name())
                log.Printf("Найдена обложка альбома: %s", albumCover)
            } else if strings.HasSuffix(file.Name(), ".mp3") || strings.HasSuffix(file.Name(), ".wav") {
                albumTracks = append(albumTracks, filepath.Join(randomDir, file.Name()))
                log.Printf("Найден трек: %s", filepath.Join(randomDir, file.Name()))
            }
        }

        if len(albumTracks) > 0 {
            return albumCover, albumTracks, nil
        }
        log.Println("В выбранной директории не найдено аудиофайлов, повторный выбор...")
    }
}

func sendMediaGroup(bot *tgbotapi.BotAPI, chatID int64, mediaGroup []interface{}, filePaths []string) error {
    const maxMediaGroupSize = 10
    const maxFileSize = 49 * 1024 * 1024 // 49 MB

    var groupToSend []interface{}
    var pathsToSend []string
    totalSize := 0

    for i, media := range mediaGroup {
        filePath := filePaths[i]

        info, err := os.Stat(filePath)
        if err != nil {
            return err
        }
        fileSize := int(info.Size())

        if totalSize+fileSize > maxFileSize {
            if err := sendBatch(bot, chatID, groupToSend); err != nil {
                return err
            }
            groupToSend = []interface{}{}
            pathsToSend = []string{}
            totalSize = 0
        }

        groupToSend = append(groupToSend, media)
        pathsToSend = append(pathsToSend, filePath)
        totalSize += fileSize
    }

    if len(groupToSend) > 0 {
        if err := sendBatch(bot, chatID, groupToSend); err != nil {
            return err
        }
    }
    return nil
}

func sendBatch(bot *tgbotapi.BotAPI, chatID int64, batch []interface{}) error {
    const maxMediaGroupSize = 10
    for i := 0; i < len(batch); i += maxMediaGroupSize {
        end := i + maxMediaGroupSize
        if end > len(batch) {
            end = len(batch)
        }
        mediaMessage := tgbotapi.NewMediaGroup(chatID, batch[i:end])
        _, err := bot.SendMediaGroup(mediaMessage)
        if err != nil {
            return err
        }
    }
    return nil
}

func main() {
    config, err := loadConfig("config.json")
    if err != nil {
        log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
    }

    directories, err := loadDirectories(config.DirectoriesFilePath)
    if err != nil {
        log.Fatalf("Не удалось загрузить директории: %v", err)
    }

    bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
    if err != nil {
        log.Fatalf("Не удалось создать бота: %v", err)
    }
    log.Printf("Запущен бот с именем: %s", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := bot.GetUpdatesChan(u)

    log.Println("Бот готов получать обновления")

    for update := range updates {
        if update.Message != nil {
            log.Printf("Получено сообщение от @%s: %s", update.Message.From.UserName, update.Message.Text)
            if update.Message.IsCommand() {
                switch update.Message.Command() {
                case "start":
                    // Создаем кнопку для запроса музыки
                    button := tgbotapi.NewKeyboardButton("Получить музыку")
                    keyboard := tgbotapi.NewReplyKeyboard(
                        tgbotapi.NewKeyboardButtonRow(button),
                    )

                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нажмите кнопку ниже, чтобы получить случайный музыкальный альбом:")
                    msg.ReplyMarkup = keyboard

                    bot.Send(msg)
                }
            } else if update.Message.Text == "Получить музыку" {
                albumCover, albumTracks, err := getRandomAlbum(directories)
                if err != nil {
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
                    bot.Send(msg)
                    log.Printf("Ошибка при получении альбома: %v", err)
                    continue
                }

                // Отправка обложки альбома с названием и исполнителем
                if albumCover != "" {
                    albumName := filepath.Base(filepath.Dir(albumCover))
                    caption := fmt.Sprintf("Альбом: %s", albumName)
                    photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(albumCover))
                    photo.Caption = caption

                    mediaGroup := []interface{}{photo}
                    if err := sendBatch(bot, update.Message.Chat.ID, mediaGroup); err != nil {
                        msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Не удалось отправить обложку альбома: %v", err))
                        bot.Send(msg)
                        log.Printf("Ошибка при отправке обложки альбома: %v", err)
                        continue
                    }
                } else {
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Альбом: %s (обложка отсутствует)", filepath.Base(filepath.Dir(albumCover))))
                    bot.Send(msg)
                }

                // Отправка аудио файлов
                var mediaGroup []interface{}
                var filePaths []string

                for _, track := range albumTracks {
                    audio := tgbotapi.NewInputMediaAudio(tgbotapi.FilePath(track))
                    mediaGroup = append(mediaGroup, audio)
                    filePaths = append(filePaths, track)
                }

                if len(mediaGroup) > 0 {
                    err = sendMediaGroup(bot, update.Message.Chat.ID, mediaGroup, filePaths)
                    if err != nil {
                        msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Не удалось отправить аудио: %v", err))
                        bot.Send(msg)
                        log.Printf("Ошибка при отправке аудио: %v", err)
                        continue
                    }
                    log.Printf("Аудио успешно отправлено")
                } else {
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "В выбранной директории не найдено аудиофайлов, попробуйте еще раз.")
                    bot.Send(msg)
                }
            }
        }
    }
}
