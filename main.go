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
	"github.com/urfave/cli/v2"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/sqlstatestore"
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
		},
		Action: func(c *cli.Context) error {
			log := zerolog.New(os.Stdout).With().Timestamp().Logger()

			// connect to the database
			conn, err := db.Open(c.String("database-dsn"))
			if err != nil {
				return errors.Wrap(err, "failed to open database")
			}

			db, err := dbutil.NewWithDB(conn.DB, db.Driver)
			if err != nil {
				return err
			}

			// start the matrix client
			mc, err := matrix.NewClient(
				c.String("matrix-homeserver"),
				c.String("matrix-username"),
				c.String("matrix-password"),
				matrix.WithJoinRooms(c.StringSlice("matrix-rooms")),
				matrix.WithLogger(&log),
				matrix.WithSyncStore(matrix.NewSQLLiteStore(conn.DB)),
				matrix.WithStateStore(func() mautrix.StateStore {
					s := sqlstatestore.NewSQLStateStore(db, dbutil.ZeroLogger(log))

					if err := s.Upgrade(); err != nil {
						panic(err)
					}

					return s
				}()),
				matrix.WithDatabaseDSN(c.String("database-dsn")),
			)
			if err != nil {
				return err
			}

			oc := openai.NewClient(
				c.String("open-ai-key"),
				openai.WithLogger(&log),
				openai.WithPrompt(c.String("prompt")),
			)

			b := athenais.New(mc, athenais.WithLogger(&log), athenais.WithPlugins(openai.NewPlugin(oc, "")))
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
					oc := openai.NewClient(c.String("open-ai-key"))
					prompt, err := oc.Prompt(c.Context, c.String("prompt"))
					if err != nil {
						return err
					}
					log.Println(prompt)
					return nil
				},
			},
		},
	}

	if err := a.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
