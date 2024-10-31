package main

import (
	"bytes"
	"fmt"
	"github.com/VillanCh/wechat2md/format"
	"github.com/VillanCh/wechat2md/parse"
	"github.com/VillanCh/wechat2md/server"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

var (
	sigExitOnce = new(sync.Once)
)

func init() {
	go sigExitOnce.Do(func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		defer signal.Stop(c)

		for {
			select {
			case <-c:
				fmt.Printf("exit by signal [SIGTERM/SIGINT/SIGKILL]")
				os.Exit(1)
				return
			}
		}
	})
}

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name: "server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "port",
					Value: "8964",
				},
			},
			Action: func(c *cli.Context) error {
				port := c.String("port")
				if port == "" {
					return errors.Errorf("port is required")
				}
				server.Start(":" + port)
				return nil
			},
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "o,output",
		},
		cli.StringFlag{
			Name:  "image",
			Value: "save",
			Usage: "image format: base64, url, save",
		},
		cli.StringFlag{
			Name:  "dir",
			Value: "docs",
		},
		cli.StringFlag{
			Name:  "static",
			Value: filepath.Join("static", "articles"),
		},
		cli.BoolFlag{
			Name:  "path-abs,abs",
			Usage: "use absolute path in img link",
		},
	}

	app.Action = func(c *cli.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return errors.Errorf("command is required")
		}

		image := c.String("image")
		if image == "" {
			return errors.Errorf("image is required")
		}
		switch image {
		case "base64", "b":
			image = "base64"
		case "url", "u":
			image = "url"
		case "save", "s":
			image = "save"
		default:
			return errors.Errorf("invalid image format: %s", image)
		}

		output := c.String("output")
		if output == "" {
			output = "output.md"
		}
		if filepath.Ext(output) != ".md" {
			output = output + ".md"
		}

		divider := string(os.PathSeparator)

		staticFileDir := c.String("static")
		staticFileDir = strings.TrimRight(staticFileDir, divider) + divider
		os.MkdirAll(staticFileDir, 0755)

		dir := c.String("dir")
		dir = strings.TrimRight(dir, divider) + divider
		os.MkdirAll(dir, 0755)

		output = filepath.Join(dir, output)

		imagePolicy := parse.ImageArgValue2ImagePolicy(image)
		fmt.Printf("url: %s, filename: %s, image: %s\n", args[0], output, image)
		article := parse.ParseFromURL(args[0], imagePolicy)
		err := format.FormatAndSave(article, output, staticFileDir)
		if err != nil {
			return err
		}

		if staticFileDir != "" {
			fmt.Println("start to handle image link")
			raw, err := os.ReadFile(output)
			if err != nil {
				return err
			}

			var prefix = ""
			if c.Bool("path-abs") {
				prefix = divider
			}
			raw = bytes.ReplaceAll(raw, []byte("wechat2md-"), []byte(prefix+staticFileDir+"wechat2md-"))
			if err := os.Remove(output); err == nil {
				os.WriteFile(output, raw, 0644)
				return nil
			}
			return errors.Errorf("failed to handle image link")
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("command: [%v] failed: %v\n", strings.Join(os.Args, " "), err)
		os.Exit(1)
		return
	}
}
