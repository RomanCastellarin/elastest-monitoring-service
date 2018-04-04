package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type message struct {
	Message string `yaml:"status,omitempty"`
	System  system `yaml:"system,omitempty"`
}

type system struct {
	CPU cpu `yaml:"cpu,omitempty"`
}

type cpu struct {
	User user `yaml:"user,omitempty"`
}

type user struct {
	Pct float64 `yaml:"pct,omitempty"`
}

func main() {
	ems := os.Getenv("ET_EMS_LSBEATS_HOST")

	// Get dashboard's rabbitMQ host and port
	// etmrq := os.Getenv("ET_ETM_RABBIT_HOST")
	// etmrqport := os.Getenv("ET_ETM_RABBIT_PORT")
	// Get EMS's API URL
	// emsAPI = fmt.Sprintf("%s:8888", ems)
	// Subscribe external ElasticSearch instance
	// emsSubscribeElasticSearch(emsAPI, "#any", "elastest.software.imdea.org", 9202, "elastic", "changeme")
	// Subscribe dashboard to tjob and sut logs
	// emsSubscribeRabbitMQ(emsAPI, "#tjobdisplay", etmrq, etmrqport, "elastic", "changeme", "tjobdisplay")
	// emsSubscribeRabbitMQ(emsAPI, "#sutdisplay", etmrq, etmrqport, "elastic", "changeme",  "sutdisplay")
	// Add routing rules for sut and tjob logs
	// emsAddMachine(emsAPI, "WHEN e.source='sutlogs' DO SENDTO '#sutdisplay'")
	// emsAddMachine(emsAPI, "WHEN e.source='tjoblogs' DO SENDTO '#tjobdisplay'")

	ems = fmt.Sprintf("%s:3232", ems)
	u := url.URL{Scheme: "ws", Host: ems, Path: "/"}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		time.Sleep(300 * time.Second)
	}
	defer c.Close()
	log.Printf("Done!")

	done := make(chan message)
	iterations := 0
	state := ""
	cpu := 0.0
	items := 0.0

	go func() {
		defer close(done)
		for {
			_, input, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			var m message
			json.Unmarshal(input, &m)
			if strings.Contains(m.Message, "STATUS_ON") {
				if state == "" {
					state = "on"
				} else {
					average := cpu / items
					if average > 0.5 {
						fmt.Println("High CPU in 'off' state")
						os.Exit(1)
					}
					log.Printf("OFF state ok with %f CPU\n", average)
					iterations++
					items = 0.0
					cpu = 0.0
				}
				log.Println("Starting ON state")
			}
			if strings.Contains(m.Message, "STATUS_OFF") {
				if state == "" {
					state = "off"
				} else {
					average := cpu / items
					if average < 0.5 {
						fmt.Println("Low CPU in 'on' state")
						os.Exit(1)
					}
					log.Printf("ON state ok with %f CPU\n", average)
					iterations++
					items = 0.0
					cpu = 0.0
				}
				log.Println("Starting OFF state")
			}
			if m.System.CPU.User.Pct != 0.0 && state != "" {
				cpu += m.System.CPU.User.Pct
				items++
			}
		}
	}()

	for iterations < 10 {
		time.Sleep(1 * time.Second)
	}
	log.Println("Test finished successfully!")

}

// func emsSubscribeElasticSearch(url string, channel string, esHost string, esPort int, user string, pass string) string {
//     json := fmt.Sprintf("{ 'channel': '%s', 'ip': '%s', 'port': %d, 'user': '%s', 'password': '%s' }", channel, esHost, esPort, user, pass)
//     if subid,err:=request.post(url+"subscriber/elasticsearch", json); err != nil {
//         log.Fatal("Could not subscribe endpoint")
//     }
//     return subid
// }

// func emsSubscribeRabbitMQ(url string, channel string, esHost string, esPort int, user string, pass string, exchange string) string {
//     json := fmt.Sprintf("{ 'channel': '%s', 'ip': '%s', 'port': %d, 'user': '%s', 'password': '%s', 'key': "", 'exchange':'%s', 'exchange_type':'fanout' }", channel, esHost, esPort, user, pass, exchange)
//     if subid,err:=request.post(url+"subscriber/rabbitmq", json); err != nil {
//         log.Fatal("Could not subscribe endpoint")
//     }
//     return subid
// }

// func emsAddMachine(url string, machine string) string {
//     json := fmt.Sprintf("{'definition': '%s', 'momType': 'filtering'}", machine)
//     if momid,err:=request.post(url/MonitoringMachine, json); err != nil {
//         log.Fatal("Could not subscribe machine")
//     }
//     return momid
// }
