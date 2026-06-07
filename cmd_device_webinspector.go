package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/danielpaulus/go-ios/ios/webinspector"
	"github.com/google/uuid"
)

func runWebInspectorCommand(cmdCtx commandContext) {
	timeoutSeconds, _ := cmdCtx.Args.String("--timeout")
	timeout := parseWebInspectorTimeout(timeoutSeconds)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := webinspector.New(cmdCtx.Device)
	exitIfError("failed connecting to webinspector", err)
	defer client.Close()
	exitIfError("failed starting webinspector", webInspectorStartupError(client.Connect(ctx)))

	if list, _ := cmdCtx.Args.Bool("list"); list {
		pages, err := client.ListPages(ctx, 500*time.Millisecond)
		exitIfError("failed listing webinspector pages", err)
		fmt.Println(convertToJSONString(pages))
		return
	}

	if launch, _ := cmdCtx.Args.Bool("launch"); launch {
		url, _ := cmdCtx.Args.String("<url>")
		bundleID, _ := cmdCtx.Args.String("--bundle-id")
		if bundleID == "" {
			bundleID = webinspector.SafariBundleID
		}
		app, err := client.OpenApp(ctx, bundleID)
		exitIfError("failed launching app", err)
		session, err := client.AutomationSession(ctx, app)
		exitIfError("failed starting automation session", err)
		defer func() {
			stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer stopCancel()
			_ = session.Stop(stopCtx)
		}()
		exitIfError("failed starting browsing context", session.Start(ctx))
		exitIfError("failed navigating", session.Navigate(ctx, url))
		currentURL, _ := session.CurrentURL(ctx)
		title, _ := session.Title(ctx)
		fmt.Println(convertToJSONString(map[string]any{"bundleID": bundleID, "url": currentURL, "title": title}))
		return
	}

	if eval, _ := cmdCtx.Args.Bool("eval"); eval {
		pageID, _ := cmdCtx.Args.String("<pageID>")
		expression, _ := cmdCtx.Args.String("<expression>")
		app, page := selectWebInspectorPage(ctx, client, pageID, "")
		if consoleEnable, _ := cmdCtx.Args.Bool("--console-enable"); consoleEnable {
			sessionID := strings.ToUpper(uuid.New().String())
			exitIfError("failed setting up inspector socket", client.SetupInspectorSocket(sessionID, app, page, false))
			_, _ = client.SendCommand(ctx, sessionID, app, page, int(time.Now().UnixNano()&0x7fffffff), "Console.enable", nil)
		}
		result, err := client.Evaluate(ctx, app, page, expression)
		exitIfError("failed evaluating JavaScript", err)
		fmt.Println(convertToJSONString(result))
		return
	}

	if shell, _ := cmdCtx.Args.Bool("js-shell"); shell {
		url, _ := cmdCtx.Args.String("<url>")
		bundleID, _ := cmdCtx.Args.String("--bundle-id")
		openSafari, _ := cmdCtx.Args.Bool("--open-safari")
		if openSafari {
			_, err := client.OpenApp(ctx, webinspector.SafariBundleID)
			exitIfError("failed opening Safari", err)
			if bundleID == "" {
				bundleID = webinspector.SafariBundleID
			}
		}
		app, page := selectWebInspectorPage(ctx, client, "", bundleID)
		if url != "" {
			_, err := client.Evaluate(ctx, app, page, fmt.Sprintf("window.location = %q", url))
			exitIfError("failed navigating", err)
		}
		runWebInspectorJSShell(ctx, client, app, page)
		return
	}

	if cdp, _ := cmdCtx.Args.Bool("cdp"); cdp {
		host, _ := cmdCtx.Args.String("--host")
		port, _ := cmdCtx.Args.Int("--port")
		if port == 0 {
			port = 9222
		}
		server := webinspector.NewCDPServer(client, host, port)
		slog.Info("webinspector CDP server started", "addr", server.Addr())
		serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		err := server.Serve(serverCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			exitIfError("webinspector CDP server failed", err)
		}
		return
	}
}

func webInspectorStartupError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("timed out waiting for Web Inspector pages; enable Settings > Safari > Advanced > Web Inspector, then reconnect the device, or retry with --timeout=<seconds>: %w", err)
	}
	return err
}

func selectWebInspectorPage(ctx context.Context, client *webinspector.Client, pageID string, bundleID string) (webinspector.Application, webinspector.Page) {
	if pageID != "" {
		app, page, ok := client.FindPage(pageID)
		if ok {
			return app, page
		}
	}
	pages, err := client.ListPages(ctx, 500*time.Millisecond)
	exitIfError("failed listing webinspector pages", err)
	var candidates []webinspector.ApplicationPage
	for _, candidate := range pages {
		if candidate.Page.Type != webinspector.WIRTypeWeb && candidate.Page.Type != webinspector.WIRTypeWebPage && candidate.Page.Type != webinspector.WIRTypeJavaScript {
			continue
		}
		if bundleID != "" && candidate.Application.BundleID != bundleID {
			continue
		}
		if pageID != "" && candidate.Page.Key != pageID {
			continue
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 {
		exitIfError("failed finding webinspector page", fmt.Errorf("no matching page found"))
	}
	if len(candidates) > 1 && pageID == "" {
		if JSONdisabled {
			for i, candidate := range candidates {
				fmt.Fprintf(os.Stderr, "[%d] %s %s %s\n", i, candidate.Application.BundleID, candidate.Page.Key, candidate.Page.URL)
			}
		}
	}
	return candidates[0].Application, candidates[0].Page
}

func runWebInspectorJSShell(ctx context.Context, client *webinspector.Client, app webinspector.Application, page webinspector.Page) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println()
				return
			}
			exitIfError("failed reading shell input", err)
		}
		expression := strings.TrimSpace(line)
		if expression == "" {
			continue
		}
		if expression == ".exit" || expression == "exit" || expression == "quit" {
			return
		}
		result, err := client.Evaluate(ctx, app, page, expression)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		fmt.Println(convertToJSONString(result))
	}
}

func parseWebInspectorTimeout(raw string) time.Duration {
	if raw == "" {
		return 5 * time.Second
	}
	seconds, err := strconv.ParseFloat(raw, 64)
	if err != nil || seconds <= 0 {
		return 5 * time.Second
	}
	return time.Duration(seconds * float64(time.Second))
}
