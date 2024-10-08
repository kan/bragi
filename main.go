package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kan/bragi/admin"
	"github.com/kan/bragi/config"
	"github.com/kan/bragi/dict"
	"github.com/kan/bragi/server"
	"github.com/pkg/errors"

	"github.com/urfave/cli/v3"
)

func main() {
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "利用方法の表示",
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "バージョン表示",
	}
	cmd := &cli.Command{
		Name:    "bragi",
		Version: "0.0.1",
		Usage:   "tiny skk server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Value:   "config.toml",
				Aliases: []string{"c"},
				Usage:   "設定ファイルパス",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "serve",
				Usage:  "SKKサーバーを実行",
				Action: serve,
			},
			{
				Name:   "update",
				Usage:  "リモート辞書の更新",
				Action: update,
			},
		},
		Action: serve, // デフォルトコマンド
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("%+v", err)
	}
}

func serve(ctx context.Context, cmd *cli.Command) error {
	conf, err := config.LoadConfig(cmd.String("config"))
	if err != nil {
		return errors.WithStack(err)
	}
	log.Printf("%+v", conf)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	restartChan := make(chan struct{}, 1)

	var skkCtx context.Context
	var cancelSKK context.CancelFunc

	runSKK := func(conf *config.Config) {
		skkCtx, cancelSKK = context.WithCancel(context.Background())
		go func() {
			if err := serveSKK(skkCtx, conf); err != nil {
				if errors.Is(err, context.Canceled) {
					log.Println("skk server stopped gracefully")
				} else {
					log.Fatalf("skk server failed: %v", err)
				}
			}
		}()
	}

	restartSKK := func(conf *config.Config) {
		if cancelSKK != nil {
			cancelSKK()
		}
		time.Sleep(100 * time.Millisecond)
		conf, err := config.LoadConfig(cmd.String("config"))
		if err != nil {
			log.Fatal("load config error", err)
		}
		runSKK(conf)
	}

	runSKK(conf)

	go func() {
		if err := serveWeb(conf, cmd.String("config"), restartChan); err != nil {
			log.Fatalf("web server failed: %v", err)
		}
	}()

	go func() {
		for {
			select {
			case <-restartChan:
				log.Println("Restarting SKK server due to config change...")
				restartSKK(conf)
			case <-skkCtx.Done():
				return
			}
		}
	}()

	<-signalChan
	log.Println("Received interrupt signal, shutting down...")

	if cancelSKK != nil {
		cancelSKK()
	}

	return nil
}

func serveSKK(ctx context.Context, conf *config.Config) error {
	l, err := net.Listen("tcp", ":"+conf.Port)
	if err != nil {
		return fmt.Errorf("failed to setup TCP server on port %s: %+v", conf.Port, err)
	}
	defer l.Close()

	s, err := server.LoadServer(conf)
	if err != nil {
		return errors.WithStack(err)
	}

	stopChan := make(chan struct{})

	go func() {
		<-ctx.Done()
		close(stopChan)
		l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-stopChan:
				return context.Canceled
			default:
				log.Println("Failed to accpet connection:", err)
				continue
			}
		}

		log.Println("accept connection")
		go s.Serve(conn)
	}
}

func serveWeb(conf *config.Config, path string, c chan struct{}) error {
	s := admin.LoadServer(conf, path, c)

	if err := s.Serve(); err != nil {
		errors.WithStack(err)
	}

	return nil
}

func update(ctx context.Context, cmd *cli.Command) error {
	conf, err := config.LoadConfig(cmd.String("config"))
	if err != nil {
		return errors.WithStack(err)
	}

	dir, err := conf.GetCacheDir()
	if err != nil {
		panic(err)
	}
	for _, dic := range conf.Dictionary {
		_, up, err := dict.NewSkkDict(dic, dir, true)
		if err != nil {
			panic(err)
		}
		if up {
			log.Printf("Update dictionary: %s ...\n", dic)
		}
	}

	return nil
}
