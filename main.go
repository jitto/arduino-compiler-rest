package main

import (
    "io"
    "os"
    "fmt"
    "log"
    "io/ioutil"
	// "bytes"
	"context"
    "net/http"
    "encoding/json"

    "github.com/arduino/arduino-cli/cli"
    "github.com/arduino/arduino-cli/cli/instance"
    "github.com/arduino/arduino-cli/commands/board"
    "github.com/arduino/arduino-cli/commands/compile"
	// "github.com/arduino/arduino-cli/cli/arguments"
    "github.com/arduino/arduino-cli/cli/errorcodes"
    "github.com/arduino/arduino-cli/configuration"
    "github.com/arduino/arduino-cli/i18n"
    rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
    )

var inst *rpc.Instance

func listBoards(w http.ResponseWriter, req *http.Request) {
	ports, err := board.List(&rpc.BoardListRequest{
		Instance: inst,
		Timeout:  2000,
	})
    if (err != nil) {
        fmt.Println("Error", err)
    } else {
        portsJson, err1 := json.Marshal(ports)
        if (err1 != nil) {
            fmt.Println("Error", err1)
        }
        fmt.Fprintf(w, "ports", portsJson)
    }
}

func compileSketch(w http.ResponseWriter, req *http.Request) {
	dir, err := ioutil.TempDir("/tmp", "arduino")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := dir + "/source/"

	err1 := os.Mkdir(path, 0755)
	if err1 != nil {
		panic(err)
	}

	outFile, err := os.Create(path + "source.ino")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, req.Body)
	outFile.Close()

	if err != nil {
		panic(err)
	}
	compileRequest := &rpc.CompileRequest{
		Instance:                      inst,
		Fqbn:                          "arduino:avr:uno",
		SketchPath:                    path,
		ShowProperties:                false,
		Preprocess:                    false,
		BuildCachePath:                dir + "/build-cache",
		BuildPath:                     dir + "/build",
		BuildProperties:               []string{},
		Warnings:                      "default",
		Verbose:                       false,
		Quiet:                         false,
		VidPid:                        "",
		ExportDir:                     "",
		Libraries:                     []string{},
		OptimizeForDebug:              false,
		Clean:                         false,
		CreateCompilationDatabaseOnly: false,
		SourceOverride:                nil,
		Library:                       []string{},
	}
	compileRes, compileError := compile.Compile(context.Background(), compileRequest, os.Stdout, os.Stderr, true)
	fmt.Println("err", compileError)
	fmt.Println("res", compileRes)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.ServeFile(w, req, dir + "/build/source.ino.hex")
}

func main() {
	configuration.Settings = configuration.Init(configuration.FindConfigFileInArgsOrWorkingDirectory(os.Args))
	i18n.Init(configuration.Settings.GetString("locale"))
	if (len(os.Args) > 1) {
		arduinoCmd := cli.NewCommand()
		if err := arduinoCmd.Execute(); err != nil {
			os.Exit(errorcodes.ErrGeneric)
		}
	} else {
		inst = instance.CreateAndInit()

		http.HandleFunc("/board", listBoards)
		http.HandleFunc("/compile", compileSketch)

		http.ListenAndServe(":8090", nil)
	}
}
