package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/hoisie/web"
	"github.com/spf13/cobra"
)

var (
	globalFlags struct {
		debug   bool
		verbose bool
	}

	mainFlags struct {
		BinaryPath string
		Config     string
		Port       int
	}

	mainCmd = &cobra.Command{
		Use:   "esp8266-updater",
		Short: "Distribute binaries for the ESP8266",
		Long:  "Manage multiple ESP8266 and distribute binaries OTA.",
		Run:   mainRun,
	}

	projectVersion string
	projectBuild   string
)

func init() {
	mainCmd.PersistentFlags().BoolVarP(&globalFlags.debug, "debug", "d", false, "Print debug output")
	mainCmd.PersistentFlags().BoolVarP(&globalFlags.verbose, "verbose", "v", false, "Print verbose output")
	mainCmd.PersistentFlags().StringVar(&mainFlags.Config, "config", "/etc/esp8266-updater/config.yml", "Configuration file")
	mainCmd.PersistentFlags().StringVar(&mainFlags.BinaryPath, "binary-path", "/var/lib/esp8266-updater", "Path where the binaries are stored")
	mainCmd.PersistentFlags().IntVar(&mainFlags.Port, "port", 8266, "Updater port")
}

func assert(err error) {
	if err != nil {
		if globalFlags.debug {
			fmt.Printf("%#v\n", err)
			os.Exit(1)
		} else {
			log.Fatal(err)
		}
	}
}

func debugHeaders(ctx *web.Context) {
	log.Println("STA-MAC: ", ctx.Request.Header.Get("x-ESP8266-STA-MAC"))
	log.Println("AP-MAC: ", ctx.Request.Header.Get("x-ESP8266-AP-MAC"))
	log.Println("free-space: ", ctx.Request.Header.Get("x-ESP8266-free-space"))
	log.Println("sketch-size: ", ctx.Request.Header.Get("x-ESP8266-sketch-size"))
	log.Println("chip-size: ", ctx.Request.Header.Get("x-ESP8266-chip-size"))
	log.Println("sdk-version: ", ctx.Request.Header.Get("x-ESP8266-sdk-version"))
	log.Println("mode: ", ctx.Request.Header.Get("x-ESP8266-mode"))
	log.Println("version: ", ctx.Request.Header.Get("x-ESP8266-version"))
}

func validHeaders(ctx *web.Context) bool {
	return ctx.Request.UserAgent() == "ESP8266-http-Update" &&
		ctx.Request.Header.Get("x-ESP8266-STA-MAC") != "" &&
		ctx.Request.Header.Get("x-ESP8266-AP-MAC") != "" &&
		ctx.Request.Header.Get("x-ESP8266-free-space") != "" &&
		ctx.Request.Header.Get("x-ESP8266-sketch-size") != "" &&
		ctx.Request.Header.Get("x-ESP8266-chip-size") != "" &&
		ctx.Request.Header.Get("x-ESP8266-sdk-version") != "" &&
		ctx.Request.Header.Get("x-ESP8266-mode") != "" &&
		ctx.Request.Header.Get("x-ESP8266-version") != ""
}

func sendFile(ctx *web.Context, filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("500 binary not found")
		debugHeaders(ctx)
		ctx.ResponseWriter.WriteHeader(500)
		return "500 binary not found"
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Println("500 binary not found")
		debugHeaders(ctx)
		ctx.ResponseWriter.WriteHeader(500)
		return "500 binary not found"
	}

	hash := md5.New()
	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, io.TeeReader(file, hash)); err != nil {
		log.Println("500 binary not found")
		debugHeaders(ctx)
		ctx.ResponseWriter.WriteHeader(500)
		return "500 binary not found"
	}
	payload := string(buf.Bytes())
	hashInBytes := hash.Sum(nil)[:16]
	hashString := hex.EncodeToString(hashInBytes)

	ctx.ContentType("application/octet-stream")
	ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s", path.Base(filename)), true)
	ctx.SetHeader("Content-Length", fmt.Sprintf("%d", fi.Size()), true)
	ctx.SetHeader("x-MD5", hashString, true)

	return payload
}

func update(ctx *web.Context, val string) string {
	ctx.ContentType("text/plain")

	if !validHeaders(ctx) {
		log.Println("403 Only for ESP8266 updater")
		debugHeaders(ctx)
		ctx.Forbidden()

		return "403 Only for ESP8266 updater"
	}

	conf, err := loadConfig(mainFlags.Config)
	if err == nil {
		if version, ok := conf.Versions[ctx.Request.Header.Get("X-ESP8266-STA-MAC")]; ok {
			if version != ctx.Request.Header.Get("X-ESP8266-VERSION") {
				return sendFile(ctx, fmt.Sprintf("%s/%s.bin", mainFlags.BinaryPath, version))
			} else {
				log.Println("304 Not modified: ", version)
				ctx.NotModified()

				return ""
			}
		}
	}

	log.Println("500 No version found")
	debugHeaders(ctx)
	ctx.ResponseWriter.WriteHeader(500)
	return "500 No version found"
}

func mainRun(cmd *cobra.Command, args []string) {
	web.Get("/update/(.*)", update)
	go web.Run(fmt.Sprintf("0.0.0.0:%d", mainFlags.Port))

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	web.Close()
}

func main() {
	mainCmd.AddCommand(versionCmd)

	mainCmd.Execute()
}
