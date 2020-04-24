package main

import (
	"flag"
	"fmt"
	"os"
)

import "./parser"
import "./config"

// HdrInfoAdd - adds license header to file f
func HdrInfoAdd(f *os.File) {
	f.WriteString(`/* SPDX-License-Identifier: GPL-2.0-only */
/* This file is part of the coreboot project. */

`)
}

// CreateGpioCfgFile - generates include file
// parser            : parser data structure
func CreateGpioCfgFile(parser *parser.ParserData) (err error) {
	hrdFile, err := os.Create("generate/gpio.h")
	if err != nil {
		fmt.Printf("Error!\n")
		return err
	}
	defer hrdFile.Close()

	HdrInfoAdd(hrdFile)
	hrdFile.WriteString(`#ifndef CFG_GPIO_H
#define CFG_GPIO_H

#include <soc/gpe.h>
#include <soc/gpio.h>
`)
	// Add the pads map
	parser.PadMapFprint(hrdFile)
	hrdFile.WriteString(`

#endif /* CFG_GPIO_H */
`)
	return nil
}

// main
func main() {
	// Command line arguments
	ConfigFile := flag.String("file",
		"inteltool.log",
		"the path to the inteltool log file")
	rawFlag := flag.Bool("raw",
		false,
		"generate macros with raw values of registers DW0, DW1")

	template := flag.Int("t", 0, "template type number\n"+
		"\t0 - inteltool.log (default)\n"+
		"\t1 - gpio.h\n"+
		"\t2 - your template\n\t")

	platform :=  flag.String("p", "snr", "set up a platform\n"+
		"\tsnr - Sunrise PCH or Skylake/Kaby Lake SoC\n"+
		"\tlbg - Lewisburg PCH with Xeon SP\n"+
		"\tapl - Apollo Lake SoC\n")

	flag.Parse()

	if valid := config.PlatformSet(*platform); valid != 0 {
		fmt.Printf("Error: invalid platform!\n")
		os.Exit(1)
	}

	inteltoolConfigFile, err := os.Open(*ConfigFile)
	if err != nil {
		fmt.Printf("Error: inteltool log file was not found!\n")
		os.Exit(1)
	}
	defer inteltoolConfigFile.Close()
	fmt.Println("file:", *ConfigFile)

	parser := parser.ParserData{RawFmt: *rawFlag,
		ConfigFile: inteltoolConfigFile,
		Template:   *template}
	parser.Parse()

	// create dir for output files
	err = os.MkdirAll("generate", os.ModePerm)
	if err != nil {
		fmt.Printf("Error! Can not create a directory for the generated files!\n")
		os.Exit(1)
	}

	// gpio.h
	err = CreateGpioCfgFile(&parser)
	if err != nil {
		fmt.Printf("Error! Can not create the file with GPIO configuration!\n")
		os.Exit(1)
	}
}
