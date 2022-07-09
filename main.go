package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	ts3 "github.com/jkoenig134/go-ts3"
	"html/template"
	"io/ioutil"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	/* flags */
	host := flag.String("host", "localhost", "ts3 server host")
	port := flag.Int("port", 10080, "ts3 server port")
	key := flag.String("key", "", "http query api key")
	flag.Parse()
	// TODO: config files / env vars
	// TODO: option for reading token from logs in ts3 server path

	/* teamspeak */
	// TODO: client reconnect strategy needed?
	var client ts3.TeamspeakHttpClient
	client = ts3.NewClient(ts3.NewConfig("http://"+*host+":"+strconv.Itoa(*port), *key))
	err := client.GlobalMessage("[b]foo[/b] \n bar")
	if err != nil {
		panic(err)
	}
	// TODO: fire own events for teamspeak users actions / messages / server stuff / etc.

	/* gin */
	// https://gin-gonic.com/docs/examples/
	gin.DisableConsoleColor()
	router := gin.Default()

	templ, err := loadTemplate()
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(templ)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "/assets/index.tmpl", map[string]any{
			"host": *host,
		})
	})

	router.GET("/dump", func(c *gin.Context) {
		c.JSON(http.StatusOK, dump(&client))
	})

	router.GET("/snapshot", func(c *gin.Context) {
		c.JSON(http.StatusOK, snapshotCreate(&client))
	})

	// TODO: handle process signals
	// https://gin-gonic.com/docs/examples/graceful-restart-or-stop/

	err = router.Run()
	if err != nil {
		panic(err)
	}
}

// loadTemplate loads templates embedded by go-assets-builder
func loadTemplate() (*template.Template, error) {
	t := template.New("")
	for name, file := range Assets.Files {
		if file.IsDir() || !strings.HasSuffix(name, ".tmpl") {
			continue
		}
		h, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		t, err = t.New(name).Parse(string(h))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func dump(teamspeakClient *ts3.TeamspeakHttpClient) map[string]any {
	data := make(map[string]any)

	hostInfo, err := teamspeakClient.HostInfo()
	if err != nil {
		panic(err)
	}
	data["host-info"] = *hostInfo

	serverList, err := teamspeakClient.ServerList()
	if err != nil {
		panic(err)
	}
	data["server-list"] = *serverList

	instance, err := teamspeakClient.InstanceInfo()
	if err != nil {
		panic(err)
	}
	data["instance"] = *instance

	channels, err := teamspeakClient.ChannelList()
	if err != nil {
		panic(err)
	}
	data["channels"] = *channels

	clientList, err := teamspeakClient.ClientList()
	if err != nil {
		panic(err)
	}
	clients := make(map[string]any)
	for _, c := range *clientList {
		info, err := teamspeakClient.ClientDbInfo(c.ClientDatabaseId)
		if err != nil {
			if strings.HasPrefix(err.Error(), "Query returned non 0 exit code: '512'") {
				/* invalid clientID */
				log.Println(err, c.ClientDatabaseId, c.ClientNickname)
			} else {
				panic(err)
			}
		} else if info != nil {
			clients[info.ClientUniqueIdentifier] = info
		}
	}
	data["clients"] = clients

	serverGroups, err := teamspeakClient.ServerGroupList()
	if err != nil {
		panic(err)
	}
	data["server-groups"] = *serverGroups

	channelGroups, err := teamspeakClient.ChannelGroupList()
	if err != nil {
		panic(err)
	}
	data["channel-groups"] = *channelGroups

	bans, err := teamspeakClient.BanList(ts3.BanListRequest{})
	if err != nil {
		if strings.HasPrefix(err.Error(), "Query returned non 0 exit code: '1281'") {
			/* database empty result set */
			log.Println(err)
		} else {
			panic(err)
		}
	} else {
		data["bans"] = *bans
	}

	complains, err := teamspeakClient.ComplainList()
	if err != nil {
		if strings.HasPrefix(err.Error(), "Query returned non 0 exit code: '1281'") {
			/* database empty result set */
			log.Println(err)
		} else {
			panic(err)
		}
	} else {
		data["complains"] = *complains
	}

	return data
}

func snapshotCreate(client *ts3.TeamspeakHttpClient) ts3.ServerSnapshot {
	snapshot, err := client.ServerSnapshotCreate(ts3.ServerSnapshotCreateRequest{Password: "foobar"})
	if err != nil {
		panic(err)
	}

	return *snapshot
}
