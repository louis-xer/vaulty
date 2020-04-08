package main

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
	"github.com/vaulty/proxy/core"
	"github.com/vaulty/proxy/proxy"
	"github.com/vaulty/proxy/storage"
	"github.com/vaulty/proxy/transformer"
)

var proxyCommand = &cli.Command{
	Name:  "proxy",
	Usage: "run proxy server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   "8080",
		},
	},
	Action: func(c *cli.Context) error {
		port := c.String("port")
		environment := c.String("environment")
		config := core.LoadConfig(fmt.Sprintf("config/%s.yml", environment))
		redisClient := core.NewRedisClient(config)
		storage := storage.NewRedisStorage(redisClient)
		transformer := transformer.NewSidekiqTransformer(redisClient)

		proxy := proxy.NewProxy(storage, transformer, config)

		fmt.Printf("==> Vaulty proxy server started on port %v! in %v environment\n", port, environment)
		http.ListenAndServe(":"+port, proxy)
		return nil
	},
}
