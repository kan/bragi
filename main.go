package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/kan/bragi/admin"
	"github.com/kan/bragi/config"
	"github.com/kan/bragi/dict"
	"github.com/kan/bragi/server"
	"github.com/pkg/errors"

	"github.com/kardianos/service"
	"github.com/urfave/cli/v3"
)

type app struct {
	logger service.Logger
	exit   chan struct{}
	cancel context.CancelFunc
	config string
}

func (a *app) run() {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	err := serve(ctx, a.config)
	if err != nil && err != context.Canceled {
		a.logger.Errorf("serve error: %+v", err)
	}
}

func (a *app) Start(s service.Service) error {
	a.logger.Info("start bragi service...")

	go a.run()
	return nil
}

func (a *app) Stop(s service.Service) error {
	a.logger.Info("stop bragi service...")
	if a.cancel != nil {
		a.cancel()
	}
	close(a.exit)
	return nil
}

func loadService(conf string) (service.Service, error) {
	appConfig := &service.Config{
		Name:        "Bragi",
		DisplayName: "Bragi SKK Server",
		Description: "It is an implementation of a customizable SKK server built in golang",
	}

	bapp := &app{
		exit:   make(chan struct{}),
		config: conf,
	}
	s, err := service.New(bapp, appConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return s, nil
}

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
				Name:  "install",
				Usage: "SKKサーバーをサービス登録",
				Action: func(ctx context.Context, c *cli.Command) error {
					s, err := loadService(c.String("config"))
					if err != nil {
						return errors.WithStack(err)
					}
					if err := s.Install(); err != nil {
						return errors.WithStack(err)
					}
					fmt.Println("Service installed.")
					return nil
				},
			},
			{
				Name:  "uninstall",
				Usage: "SKKサーバーをサービス登録解除",
				Action: func(ctx context.Context, c *cli.Command) error {
					s, err := loadService(c.String("config"))
					if err != nil {
						return errors.WithStack(err)
					}
					if err := s.Uninstall(); err != nil {
						return errors.WithStack(err)
					}
					fmt.Println("Service uninstalled.")
					return nil
				},
			},
			{
				Name:  "start",
				Usage: "サービス登録したSKKサーバーを開始",
				Action: func(ctx context.Context, c *cli.Command) error {
					s, err := loadService(c.String("config"))
					if err != nil {
						return errors.WithStack(err)
					}
					if err := s.Start(); err != nil {
						return errors.WithStack(err)
					}
					fmt.Println("Service started...")
					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "サービス登録したSKKサーバーを停止",
				Action: func(ctx context.Context, c *cli.Command) error {
					s, err := loadService(c.String("config"))
					if err != nil {
						return errors.WithStack(err)
					}
					if err := s.Stop(); err != nil {
						return errors.WithStack(err)
					}
					fmt.Println("Service stopped.")
					return nil
				},
			},
			{
				Name:  "restart",
				Usage: "サービス登録したSKKサーバーを再起動",
				Action: func(ctx context.Context, c *cli.Command) error {
					s, err := loadService(c.String("config"))
					if err != nil {
						return errors.WithStack(err)
					}
					if err := s.Restart(); err != nil {
						return errors.WithStack(err)
					}
					fmt.Println("Service restarted...")
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "サービス登録したSKKサーバーの状態確認",
				Action: func(ctx context.Context, c *cli.Command) error {
					s, err := loadService(c.String("config"))
					if err != nil {
						return errors.WithStack(err)
					}
					sts, err := s.Status()
					if err != nil {
						return errors.WithStack(err)
					}
					switch sts {
					case service.StatusRunning:
						fmt.Println("Service running.")
					case service.StatusStopped:
						fmt.Println("Service stopped.")
					default:
						fmt.Println("Service unknown.")
					}
					return nil
				},
			},
			{
				Name:  "run",
				Usage: "SKKサーバーを実行",
				Action: func(ctx context.Context, c *cli.Command) error {
					return serve(ctx, c.String("config"))
				},
			},
			{
				Name:   "update",
				Usage:  "リモート辞書の更新",
				Action: update,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			// デフォルトコマンド
			return serve(ctx, c.String("config"))
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("%+v", err)
	}
}

func serve(ctx context.Context, cpath string) error {
	conf, err := config.LoadConfig(cpath)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Printf("%+v", conf)

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

	restartSKK := func() {
		if cancelSKK != nil {
			cancelSKK()
		}
		time.Sleep(100 * time.Millisecond)
		cf, err := config.LoadConfig(cpath)
		if err != nil {
			log.Fatal("load config error", err)
		}
		runSKK(cf)
	}

	runSKK(conf)

	go func() {
		if err := serveWeb(conf, cpath, restartChan); err != nil {
			log.Fatalf("web server failed: %v", err)
		}
	}()

	go func() {
		for {
			select {
			case <-restartChan:
				log.Println("Restarting SKK server due to config change...")
				restartSKK()
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
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
