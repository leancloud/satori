package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/amir/raidman"
)

func Run(dir string, name string, arg ...string) string {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(out.String())
}

type LogEntry struct {
	Message    string `json:"message"`
	Level      string `json:"level"`
	Logger     string `json:"logger_name"`
	StackTrace string `json:"stack_trace"`
}

func GetRiemannConn() *raidman.Client {
	var err error
	for i := 0; i < 3; i++ {
		conn, err := raidman.DialWithTimeout(
			"tcp",
			"localhost:5555",
			time.Duration(5)*time.Second,
		)
		if err == nil {
			return conn
		}
		time.Sleep(time.Second * 3)
	}
	panic(err)
}

func main() {
	gitVer := "meh"

	for {
		newGitVer := Run("/satori-conf", "git", "rev-parse", "HEAD")
		if newGitVer == gitVer {
			time.Sleep(time.Second * 1)
			continue
		}

		gitVer = newGitVer

		file, _ := os.OpenFile("/var/log/riemann.log.json", os.O_CREATE|os.O_RDONLY, 0777)
		file.Seek(0, 2)
		f := bufio.NewReader(file)

		for {
			_, err := f.ReadString('\n')
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
		}

		log.Println("Begin reloading riemann...")
		Run("/satori-conf", "git", "reset", "--hard")
		Run("/satori-conf", "git", "clean", "-f", "-d")

		pid := Run("/satori-conf", "supervisorctl", "pid", "riemann")
		if pid == "0" {
			log.Println("Riemann not running, starting...")
			Run("/satori-conf", "supervisorctl", "start", "riemann")
			time.Sleep(time.Second * 20)
		} else {
			Run("/satori-conf", "kill", "-SIGHUP", pid)
		}

		/*
			Ttl         float32           `json:"ttl,omitempty"`
			Time        int64             `json:"time,omitempty"`
			Tags        []string          `json:"tags,omitempty"`
			Host        string            `json:"host,omitempty"` // Defaults to os.Hostname()
			State       string            `json:"state,omitempty"`
			Service     string            `json:"service,omitempty"`
			Metric      interface{}       `json:"metric,omitempty"` // Could be Int, Float32, Float64
			Description string            `json:"description,omitempty"`
			Attributes  map[string]string `json:"attributes,omitempty"`

			{
			  "@timestamp": "2016-09-07T09:19:25.603+00:00",
			  "@version": 1,
			  "message": "Couldn't reload:",
			  "logger_name": "riemann.bin",
			  "thread_name": "SIGHUP handler",
			  "level": "ERROR",
			  "level_value": 40000,
			  "stack_trace": "",
			  "HOSTNAME": ""
			}
		*/
		var entry LogEntry
		var errord bool = false
		conn := GetRiemannConn()
		for i := 0; i < 5; {
			s, err := f.ReadBytes('\n')
			if err == io.EOF {
				errord = true
				time.Sleep(time.Second * 1)
				i++
				continue
			}
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(s, &entry); err != nil {
				continue
			}
			errord = false
			if entry.Message == "Hyperspace core online" {
				break
			}
			if entry.Message == "Couldn't reload:" {
				errord = true
				conn.Send(&raidman.Event{
					Service:     ".satori.riemann.reload-err",
					Host:        "Satori",
					Metric:      1,
					Description: entry.StackTrace,
				})
			}
		}
		if !errord {
			time.Sleep(time.Second * 3)
			conn.Send(&raidman.Event{
				Service:     ".satori.riemann.newconf",
				Host:        "Satori",
				Metric:      1,
				Description: newGitVer,
			})
			log.Println("Success")
		} else {
			log.Println("Reload error!")
		}
		conn.Close()
		file.Close()
	}
}
