package main
import (
    "github.com/Syfaro/telegram-bot-api"
    "log"
    "fmt"
    "os"
    "math"
    "time"
    "strconv"
    "encoding/json"
    "io/ioutil"
    "bits/api"
)

type AuthJson struct {
    Key string `json:"key"`
    Sec string `json:"secret"`
    Uid string `json:"id"`
}

type Ticker struct {
    Last float64 `json:"last"`
    Timestamp float64 `json:"timestamp"`
}

type TickerData struct {
    Hour Ticker `json:"hour"`
    Day Ticker `json:"day"`
    Week Ticker `json:"week"`
}

type UserList struct {
    Users []int64
}


func Exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, err }
    return true, err
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func percentCalc(timer string) (TickerData, string) {
    var ticker *Ticker
    var tickerFile TickerData
    var config_data api.Config
    var msgNotify string
    var timingMsg string
    var percentLimit float64

    current_hour, _ := api.GetApiWrapper(config_data, "www.bitstamp.net/api/v2/ticker/btcusd/")

    exi, _ := Exists("ticker.json")
    if exi {
        tickerF, _ := ioutil.ReadFile("ticker.json")
        _ = json.Unmarshal([]byte(tickerF), &tickerFile)
        if timer == "h" {
            ticker = &tickerFile.Hour
            timingMsg = " hourly"
            percentLimit = 3
        } else if timer == "d" {
            ticker = &tickerFile.Day
            timingMsg = " daily"
            percentLimit = 5
        } else if timer == "w" {
            ticker = &tickerFile.Week
            timingMsg = " weekly"
            percentLimit = 10
        }


        currentLast, _ := strconv.ParseFloat(current_hour["last"], 32)

        var percent string
        log.Println("cur/ticker", currentLast, ticker.Last)
        percentDiff := (1 - currentLast/ticker.Last)*100
        if math.Abs(percentDiff) >= percentLimit {
            if percentDiff >= 0 {
                percent = strconv.FormatFloat((1 - currentLast/ticker.Last)*100, 'f', 6, 64)
                msgNotify = "Prise fell by " + percent + "%" + timingMsg
            } else if percentDiff <= 0 {
                percent = strconv.FormatFloat((1 - ticker.Last/currentLast)*100, 'f', 6, 64)
                msgNotify = "Prise rose by " + percent + "%" + timingMsg
            }

        }
    }

    ticker.Last, _ = strconv.ParseFloat(current_hour["last"], 32)
    ticker.Timestamp, _ = strconv.ParseFloat(current_hour["timestamp"], 32)

    log.Println("tickerfile", tickerFile)
    return tickerFile, msgNotify
}

func notifyWrapper(tickerType string, bot *tgbotapi.BotAPI) {
    newTicker, msgNotify := percentCalc(tickerType)

    fileW, _ := json.MarshalIndent(newTicker, "", " ")
    err := ioutil.WriteFile("ticker.json", fileW, 0644)
    check(err)

    var usersData UserList
    fileR, _ := ioutil.ReadFile("users.json")
    _ = json.Unmarshal([]byte(fileR), &usersData.Users)

    for _, chatid := range usersData.Users {
        msg := tgbotapi.NewMessage(chatid, msgNotify)
        bot.Send(msg)
    }

    log.Println(newTicker)
}

func main() {
    bot, err := tgbotapi.NewBotAPI("bot_token_here")
    if err != nil {
        log.Panic(err)
    }

    bot.Debug = true
    log.Printf("Authorized on account %s", bot.Self.UserName)

    var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
    ucfg.Offset = 0
    ucfg.Timeout = 1
    upd, _ := bot.GetUpdatesChan(ucfg)

    var data []string
    var registerMode bool
    var registerQuestion int
    regSequence := []string{"Send API key", "Send API secret", "Send UID"}

    tickerHour := time.NewTicker(60 * time.Minute)
    tickerDay := time.NewTicker(24 * 60 * time.Minute)
    tickerWeek := time.NewTicker(7 * 24 * 60 * time.Minute)

    for {
        select {

            case update := <-upd :
                if update.Message == nil {
                    break
                }

                UserName := update.Message.From.UserName
                ChatID := update.Message.Chat.ID
                Text := update.Message.Text

                log.Printf("[%s] %d %s", UserName, ChatID, Text)

                msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
                if update.Message.IsCommand() {
                    switch update.Message.Command() {
                    //TODO cases: /balance, /last (order), /open (order)
                    case "help":
                        msg.Text = "Commands are: \n/help - help message\n/start - recieve notifications about price change\n/register - auth with API keys to trade with this bot\n/whoami - show API config\n/suicide - remove both notifications and API access\n/current - hourly open and last prices\n/about - what does this bot do\n\n(registered users only:)\n/buy - buy BTC for all USD balance\n/sell - sell all available BTC\n/open - show open orders\n/cancel - cancel all orders"
                        bot.Send(msg)
                    case "start":
                        exi, _ := Exists("users.json")
                        var usersData UserList
                        var inList bool
                        if exi {
                            file, _ := ioutil.ReadFile("users.json")
                            usersData = UserList{}
                            _ = json.Unmarshal([]byte(file), &usersData.Users)
                            for _, item := range usersData.Users {
                                if item == ChatID {
                                    inList = true
                                }
                            }
                        }
                        if inList {
                            msg.Text = "Already monintoring for you. Trigger is >=3% increase/decrease per hour/day/week"
                            bot.Send(msg)
                        } else {
                            usersData.Users = append(usersData.Users, ChatID)

                            file, _ := json.MarshalIndent(usersData.Users, "", " ")
                            err := ioutil.WriteFile("users.json", file, 0644)
                            check(err)

                            msg.Text = "Started monintoring for you. Trigger is >=3% increase/decrease per hour/day/week"
                            bot.Send(msg)
                        }
                    case "register":
                        _ , err := Exists("users")
                        if err != nil {
                            os.MkdirAll("users/", os.ModePerm)
                        }

                        exi, _ := Exists("users/" + UserName + ".json")
                        if exi {
                            msg.Text = "Already registered"
                            bot.Send(msg)
                        } else {
                            registerMode = true
                            registerQuestion = 0
                            data = []string{}

                            msg.Text = regSequence[registerQuestion]
                            bot.Send(msg)
                        }
                    case "whoami":
                        exi, _ := Exists("users/" + UserName + ".json")
                        if exi {
                            file, _ := ioutil.ReadFile("users/" + UserName + ".json")

                            msg.Text = string(file)
                            bot.Send(msg)
                        } else {
                            msg.Text = "Your API is not registered yet"
                            bot.Send(msg)
                        }
                    case "suicide":
                        exi, _ := Exists("users/" + UserName + ".json")
                        if exi {
                            os.Remove("users/" + UserName + ".json")

                            msg.Text = "Successfully deleted your config"
                            bot.Send(msg)
                        } else {
                            msg.Text = "You API is not registered yet"
                            bot.Send(msg)
                        }
                        exi, _ = Exists("users.json")
                        if exi {
                            var usersData UserList
                            fileR, _ := ioutil.ReadFile("users.json")
                            _ = json.Unmarshal([]byte(fileR), &usersData.Users)
                            firstLen := len(usersData.Users)
                            for ind, item := range usersData.Users {
                                if item == ChatID {
                                    usersData.Users = append(usersData.Users[:ind], usersData.Users[ind+1:]...)
                                    fileW, _ := json.MarshalIndent(usersData.Users, "", " ")
                                    err := ioutil.WriteFile("users.json", fileW, 0644)
                                    check(err)
                                }
                            }
                            secLen := len(usersData.Users)
                            if firstLen != secLen {
                                msg.Text = "Successfully removed you from notification list"
                                bot.Send(msg)
                            } else {
                                msg.Text = "You are not in notification list"
                                bot.Send(msg)
                            }
                        } else {

                        }
                    case "about":
                        msg.Text = "This bot is designed to notify about BTC price change on Bitstamp exchange. You are added to notification list after /start and you can exit, doing /suicide . Notifications are sent in case of >= 3% price change up or down. Monitored periods are 1h., 1d., 7d. By doing /register you can also add your API keys to trade instantly, in this case you will recieve notifications only when recommended action is opposite to your last (you bought BTC, BTC price rose -> notification to sell /VS/ you sold BTC, BTC price rose -> NO notification). The /buy and /sell commands operate on ALL OF THE BALANCE, placed orders are MARKET.\n\n-disclaimer-\nPlease, be concerned, that you are the only one responsible for any profit gain/loss."
                        bot.Send(msg)
                    case "debug":
                        msg.Text = "wtf. u hacker? gtfo, nice try"
                        bot.Send(msg)
                    case "balance":
                        exi, _ := Exists("users/" + UserName + ".json")
                        if exi {
                            config_data := api.ReadJson("users/" + UserName + ".json")
                            postData := map[string]string{"": ""}
                            balance, _ := api.PostApiWrapper(config_data, "www.bitstamp.net/api/v2/balance/btcusd/", postData)

                            msg.Text = "BTC: " + balance["btc_available"] + "/ USD: " + balance["usd_available"]
                            bot.Send(msg)
                        } else {
                            msg.Text = "You API is not registered yet"
                            bot.Send(msg)
                        }
                    case "buy":
                        exi, _ := Exists("users/" + UserName + ".json")
                        if exi {
                            config_data := api.ReadJson("users/" + UserName + ".json")

                            postData := map[string]string{"": ""}
                            balance, _ := api.PostApiWrapper(config_data, "www.bitstamp.net/api/v2/balance/btcusd/", postData)
                            last_hour, _ := api.GetApiWrapper(config_data, "www.bitstamp.net/api/v2/ticker_hour/btcusd/")

                            last_price, _ := strconv.ParseFloat(last_hour["last"], 32)
                            usd_balance, _ := strconv.ParseFloat(balance["usd_available"], 32)
                            btc_amount := fmt.Sprintf("%f", usd_balance/last_price)

                            postBuyData := map[string]string{"amount": btc_amount}
                            ans, _ := api.PostApiWrapper(config_data, "www.bitstamp.net/api/v2/buy/market/btcusd/", postBuyData)
                            log.Println("buy ans: ", ans)
                        } else {
                            msg.Text = "You API is not registered yet"
                            bot.Send(msg)
                        }
                    case "sell":
                      exi, _ := Exists("users/" + UserName + ".json")
                      if exi {
                          config_data := api.ReadJson("users/" + UserName + ".json")

                          postSellData := map[string]string{"": ""}
                          balance, _ := api.PostApiWrapper(config_data, "www.bitstamp.net/api/v2/balance/btcusd/", postSellData)

                          postSellData = map[string]string{"amount": balance["btc_available"]}
                          ans, _ := api.PostApiWrapper(config_data, "www.bitstamp.net/api/v2/buy/market/btcusd/", postSellData)
                          log.Println("sell ans: ", ans)
                      } else {
                          msg.Text = "You API is not registered yet"
                          bot.Send(msg)
                      }
                    case "current":
                      var config_data api.Config
                      last_hour, _ := api.GetApiWrapper(config_data, "www.bitstamp.net/api/v2/ticker_hour/btcusd/")

                      msg.Text = "Open/last: " + last_hour["open"] + "/" + last_hour["last"]
                      bot.Send(msg)
                    case "open":
                        exi, _ := Exists("users/" + UserName + ".json")
                        if exi {
                            config_data := api.ReadJson("users/" + UserName + ".json")
                            postData := map[string]string{"": ""}
                            orders, _ := api.PostApiWrapper(config_data, "www.bitstamp.net/api/v2/open_orders/all/", postData)

                            log.Println("open orders: ", orders)
                        } else {
                            msg.Text = "You API is not registered yet"
                            bot.Send(msg)
                        }
                    case "cancel":
                        exi, _ := Exists("users/" + UserName + ".json")
                        if exi {
                            config_data := api.ReadJson("users/" + UserName + ".json")
                            postData := map[string]string{"": ""}
                            orders, _ := api.PostApiWrapper(config_data, "www.bitstamp.net/api/cancel_all_orders/", postData)

                            log.Println("order cancelation: ", orders)
                            msg.Text = "Your orders have been canceled."
                            bot.Send(msg)
                        } else {
                            msg.Text = "You API is not registered yet"
                            bot.Send(msg)
                        }
                    default:
                        msg.Text = "I don't know that command"
                        bot.Send(msg)
                    }
                } else if registerMode {
                    data = append(data, Text)
                    registerQuestion = registerQuestion + 1

                    if len(regSequence) != registerQuestion {
                        msg.Text = regSequence[registerQuestion]
                        bot.Send(msg)
                    } else if len(regSequence) == registerQuestion {
                        registerMode = false
                        dataJson := AuthJson{data[0], data[1], data[2]}
                        file, _ := json.MarshalIndent(dataJson, "", " ")
                        err := ioutil.WriteFile("users/" + UserName + ".json", file, 0644)
                        check(err)

                        msg.Text = "Registered successfully!"
                        bot.Send(msg)
                    }
                } else {
                    msg.Text = Text
                    bot.Send(msg)
                }


                if update.UpdateID >= ucfg.Offset {
                    ucfg.Offset = update.UpdateID + 1
                }

            default:
                select {
                    case <- tickerHour.C:
                        notifyWrapper("h", bot)
                    case <- tickerDay.C:
                        notifyWrapper("d", bot)
                    case <- tickerWeek.C:
                        notifyWrapper("w", bot)

                    default:
                    //DO NOTHING
                }
        }
    }
}
