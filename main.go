package main
import(
	"log"
	"fmt"
	"os"
	"time"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/cavaliercoder/grab"
	"context"
    "io/ioutil"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/drive/v3"
    "io"
    "net/http"
    "golang.org/x/oauth2"
    "encoding/json"

)
func main(){
	bot , err := tgbotapi.NewBotAPI("your awesome key comes here")
	if err != nil{
		log.Panic(err)
	}
	
	
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	service, err := getService()
	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "help":
				msg.Text = "type /sayhi or /status."
			case "sayhi":
				msg.Text = "or kabir kesaa hai"
			case "status":
				msg.Text = "ye command aroree ke naam"
			case "download": 
				runes := []rune(update.Message.Text)
				substr:=string(runes[10:])
				show := download(substr)
				msg.Text = show
				f, err := os.Open(show)
				 if err != nil {
      				panic(fmt.Sprintf("cannot open file: %v", err))
  				 }
  				defer f.Close()
  				file, err := createFile(service, show, "random", f, "root")
  				if err != nil {
     				panic(fmt.Sprintf("Could not create file: %v\n", err))
   				}
				fmt.Printf("File '%s' successfully uploaded.", file.ID)

			default:
				msg.Text = "I don't know that command"
			}
			bot.Send(msg)
		}
	}
}
func download(uri string)(string){
	fmt.Printf("%v",uri)
	client := grab.NewClient()
	req, _ := grab.NewRequest(".", uri)
	fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	//fmt.Printf("  %v\n", resp.HTTPResponse.Status)
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Download saved to ./%v \n", resp.Filename)
	return resp.Filename

}

func getClient(config *oauth2.Config) *http.Client {
   // The file token.json stores the user's access and refresh tokens, and is
   // created automatically when the authorization flow completes for the first
   // time.
   tokFile := "token.json"
   tok, err := tokenFromFile(tokFile)
   if err != nil {
      tok = getTokenFromWeb(config)
      saveToken(tokFile, tok)
   }
   return config.Client(context.Background(), tok)
}
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
   authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
   fmt.Printf("Go to the following link in your browser then type the "+
      "authorization code: \n%v\n", authURL)

   var authCode string
   if _, err := fmt.Scan(&authCode); err != nil {
      log.Fatalf("Unable to read authorization code %v", err)
   }

   tok, err := config.Exchange(context.TODO(), authCode)
   if err != nil {
      log.Fatalf("Unable to retrieve token from web %v", err)
   }
   return tok
}
// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
   f, err := os.Open(file)
   if err != nil {
      return nil, err
   }
   defer f.Close()
   tok := &oauth2.Token{}
   err = json.NewDecoder(f).Decode(tok)
   return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
   fmt.Printf("Saving credential file to: %s\n", path)
   f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
   if err != nil {
      log.Fatalf("Unable to cache oauth token: %v", err)
   }
   defer f.Close()
   json.NewEncoder(f).Encode(token)
}
func getService() (*drive.Service, error) {
   b, err := ioutil.ReadFile("credentials.json")
   if err != nil {
      fmt.Printf("Unable to read credentials.json file. Err: %v\n", err)
      return nil, err
   }

   // If modifying these scopes, delete your previously saved token.json.
   config, err := google.ConfigFromJSON(b, drive.DriveFileScope)

   if err != nil {
      return nil, err
   }

   client := getClient(config)

   service, err := drive.New(client)

   if err != nil {
      fmt.Printf("Cannot create the Google Drive service: %v\n", err)
      return nil, err
   }

   return service, err
}
func createFile(service *drive.Service, name string, mimeType string, content io.Reader, parentId string) (*drive.File, error) {

   f := &drive.File{
      MimeType: mimeType,
      Name:     name,
      Parents:  []string{parentId},
   }

   file, err := service.Files.Create(f).Media(content).Do()

   if err != nil {
      log.Println("Could not create file: " + err.Error())
      return nil, err
   }

   return file, nil
}




 