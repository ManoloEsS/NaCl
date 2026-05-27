package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	apperr "github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/server"
	"github.com/joho/godotenv"
	pkgerr "github.com/pkg/errors"
)

type closeLogger func() error

type stackTracer interface {
	error
	StackTrace() pkgerr.StackTrace
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	status := run(ctx, cancel)
	cancel()

	os.Exit(status)
}

func run(ctx context.Context, cancel context.CancelFunc) int {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load env vars")
	}
	logger, closeLogger, err := initializeLogger(os.Getenv("SALT_LOG_FILE"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		return 1
	}

	//get port
	httpPort, err := strconv.Atoi(os.Getenv("SALT_PORT"))
	if err != nil {
		logger.Error("failed to fetch PORT from .env", "error", err)
	}
	//get db string
	// dbString := os.Getenv("SALT_DB_STRING")
	//initialize db

	defer func() {
		if err := closeLogger(); err != nil {
			fmt.Fprintf(os.Stderr, "logger shutdown with error: %v\n", err)
		}
		//close db
	}()

	s := server.NewServer(httpPort, cancel, logger)
	var serverErr error
	go func() {
		serverErr = s.Start()
	}()

	//blocks
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Debug("Salt is shutting down")
	if err := s.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown server", "error", err)
		return 1
	}
	if serverErr != nil {
		logger.Error("server error", "error", serverErr)
		return 1
	}
	return 0
}

func initializeLogger(logFile string) (*slog.Logger, closeLogger, error) {
	debugHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: replaceAttr,
	})

	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log file: %v", err)
	}

	bufferedWriter := bufio.NewWriterSize(file, 8192)
	closeLogger := func() error {
		err := bufferedWriter.Flush()
		if err != nil {
			return err
		}
		_ = file.Close()
		return nil
	}

	infoHandler := slog.NewJSONHandler(bufferedWriter, &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: replaceAttr,
	})

	logger := slog.New(slog.NewMultiHandler(
		debugHandler,
		infoHandler,
	))

	return logger, closeLogger, nil
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "error" {
		err, ok := a.Value.Any().(error)
		if !ok {
			return a
		}
		errorAttrs := []slog.Attr{
			slog.String("message", err.Error()),
		}
		if stackErr, ok := errors.AsType[stackTracer](err); ok {
			errorAttrs = append(errorAttrs, slog.Any("stack", stackErr.StackTrace()))
		}
		//make app error
		errorAttrs = append(errorAttrs, apperr.Attrs(err)...)
		return slog.GroupAttrs("error", errorAttrs...)
	}
	return a
}

// func main() {
// 	r := chi.NewRouter()
//
// 	// set middleware
// 	// r.Use(middleware.SomeMiddleware)
//
// 	//index routes
// 	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("root."))
// 	})
//
// 	//create subroutes
// 	//r.Route("/users", func(r chi.Router) {
// 	//r.With(somemiddleware).Get("/", ListUsers)
// 	//})
//
// 	//generate docs
// 	// if *routes {
// 	// 	// fmt.Println(docgen.JSONRoutesDoc(r))
// 	// 	fmt.Println(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
// 	// 		ProjectPath: "github.com/go-chi/chi/v5",
// 	// 		Intro:       "Welcome to the chi/_examples/rest generated docs.",
// 	// 	}))
// 	// 	return
// 	// }
//
// 	http.ListenAndServe(":3333", r)
// }
