package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/unerror/athenais/internal/db"
	"github.com/unerror/athenais/internal/matrix"
	"github.com/unerror/athenais/pkg/athenais"
	"github.com/unerror/athenais/plugins/openai"
	"github.com/unerror/athenais/plugins/sayhi"
	"github.com/urfave/cli/v2"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/util/dbutil"
)

func main() {
	a := &cli.App{
		Name:  "athenias",
		Usage: "Athenias is a Matrix bot for interacting with OpenAI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "open-ai-key",
				Usage:   "OpenAI API key",
				EnvVars: []string{"OPEN_AI_KEY"},
			},
			&cli.StringFlag{
				Name:    "matrix-homeserver",
				Usage:   "Matrix homeserver URL",
				EnvVars: []string{"MATRIX_HOMESERVER"},
			},
			&cli.StringFlag{
				Name:    "matrix-username",
				Usage:   "Matrix username",
				EnvVars: []string{"MATRIX_USERNAME"},
			},
			&cli.StringFlag{
				Name:    "matrix-password",
				Usage:   "Matrix password",
				EnvVars: []string{"MATRIX_PASSWORD"},
			},
			&cli.StringSliceFlag{
				Name:    "matrix-rooms",
				Usage:   "Matrix rooms to join",
				EnvVars: []string{"MATRIX_ROOMS"},
			},
			&cli.StringFlag{
				Name:    "crypto-pickle-key",
				Usage:   "The crypto pickle key to use for the client crypto helper.",
				Value:   "update-me",
				EnvVars: []string{"CRYPTO_PICKLE_KEY"},
			},
			&cli.StringFlag{
				Name:    "database-dsn",
				Usage:   "Database DSN",
				Value:   "athenias.sqlite3",
				EnvVars: []string{"DATABASE_DSN"},
			},
			&cli.StringFlag{
				Name:    "prompt",
				Usage:   "The prompt to use for the system chat",
				Value:   openai.DefaultPrompt,
				EnvVars: []string{"OPENAI_PROMPT"},
			},
			&cli.IntFlag{
				Name:    "log-level",
				Usage:   "Log level. 0 = Debug, 1 = Info, 2 = Warn, 3 = Error, 4 = Fatal, 5 = Panic",
				Value:   int(zerolog.InfoLevel),
				EnvVars: []string{"LOG_LEVEL"},
			},
			&cli.BoolFlag{
				Name:    "log-pretty",
				Value:   false,
				Usage:   "Log in pretty format",
				EnvVars: []string{"LOG_PRETTY"},
			},
		},
		Action: func(c *cli.Context) error {
			log := zerolog.New(os.Stdout).With().
				Timestamp().
				Logger().
				Level(zerolog.Level(c.Int("log-level")))

			if c.Bool("log-pretty") {
				log = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
			}

			// connect to the database
			conn, err := db.Open(c.String("database-dsn"))
			if err != nil {
				return errors.Wrap(err, "failed to open database")
			}

			dbu, err := dbutil.NewWithDB(conn.DB, db.Driver)
			if err != nil {
				return err
			}
			_ = dbu

			// account data store event type
			adsevt := matrix.NewAccountDataStoreEventType(id.UserID(c.String("matrix-username")))
			// start the matrix client
			mc, err := matrix.NewClient(
				c.String("matrix-homeserver"),
				c.String("matrix-username"),
				c.String("matrix-password"),
				matrix.WithJoinRooms(c.StringSlice("matrix-rooms")),
				matrix.WithLogger(log),
				matrix.WithSyncStore(
					matrix.WithEventType(adsevt),
				),
				matrix.WithPickleKey(c.String("matrix-username")),
				matrix.WithStateStore(
					matrix.WithSQLiteStateStore(dbu),
				),
				matrix.WithCryptoHelperStore(
					matrix.WithSQLCryptoStore(dbu),
				),
			)
			if err != nil {
				return err
			}

			oc := openai.NewClient(
				c.String("open-ai-key"),
				openai.WithLogger(&log),
				openai.WithPrompt(c.String("prompt")),
			)

			b := athenais.New(
				mc,
				athenais.WithLogger(&log),
				athenais.WithPlugins(
					openai.NewPlugin(oc, ""),
					sayhi.NewPlugin(),
				),
			)
			if err := b.Run(c.Context); err != nil {
				return err
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "prompt",
				Usage: "Generate a prompt for the given prompt",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "prompt",
						Usage:   "Prompt to generate a prompt for",
						EnvVars: []string{"PROMPT"},
					},
					&cli.StringFlag{
						Name:    "open-ai-key",
						Usage:   "OpenAI API key",
						EnvVars: []string{"OPEN_AI_KEY"},
					},
				},
				Action: func(c *cli.Context) error {
					log := zerolog.New(os.Stdout).With().
						Timestamp().
						Logger().
						Level(zerolog.Level(c.Int("log-level")))

					oc := openai.NewClient(c.String("open-ai-key"),
						openai.WithLogger(&log),
						openai.WithPrompt(c.String("prompt")),
					)

					response, err := oc.Prompt(c.Context, c.Args().First())

					if err != nil {
						return err
					}

					log.Print(response)

					return nil
				},
			},
		},
	}

	if err := a.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
