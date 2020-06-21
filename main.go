package main

import (
	"flag"
	"fmt"
	"os"
)

import "./parser"
import "./config"

// generateOutputFile - generates include file
// parser            : parser data structure
func generateOutputFile(parser *parser.ParserData) (err error) {

	config.OutputGenFile.WriteString(`/* SPDX-License-Identifier: GPL-2.0-only */

#ifndef CFG_GPIO_H
#define CFG_GPIO_H

#include <soc/gpio.h>

/* Pad configuration  */
static const struct pad_config gpio_table[] = {
`)
	// Add the pads map
	parser.PadMapFprint()
	config.OutputGenFile.WriteString(`};

#endif /* CFG_GPIO_H */
`)
	return nil
}

// main
func main() {
	// Command line arguments
	inputFileName := flag.String("file",
		"inteltool.log",
		"the path to the inteltool log file")

	outputFileName := flag.String("o",
		"generate/gpio.h",
		"the path to the generated file with GPIO configuration")

	rawFlag := flag.Bool("raw",
		false,
		"generate macros with raw values of registers DW0, DW1")

	advFlag := flag.Bool("adv",
		false,
		"generate advanced macros only")

	nonCheckFlag := flag.Bool("n",
		false,
		"generate macros without checking.\n" +
		"In this case, some fields of the configuration registers\n" +
		"DW0 will be ignored.")

	template := flag.Int("t", 0, "template type number\n"+
		"\t0 - inteltool.log (default)\n"+
		"\t1 - gpio.h\n"+
		"\t2 - your template\n\t")

	platform :=  flag.String("p", "snr", "set up a platform\n"+
		"\tsnr - Sunrise PCH or Skylake/Kaby Lake SoC\n"+
		"\tlbg - Lewisburg PCH with Xeon SP\n"+
		"\tapl - Apollo Lake SoC\n")

	flag.Parse()

	config.RawFormatFlagSet(*rawFlag)
	config.AdvancedFormatFlagSet(*advFlag)
	config.NonCheckingFlagSet(*nonCheckFlag)

	if valid := config.PlatformSet(*platform); valid != 0 {
		fmt.Printf("Error: invalid platform!\n")
		os.Exit(1)
	}

	fmt.Println("Log file:", *inputFileName)
	fmt.Println("Output generated file:", *outputFileName)

	inputRegDumpFile, err := os.Open(*inputFileName)
	if err != nil {
		fmt.Printf("Error: inteltool log file was not found!\n")
		os.Exit(1)
	}
	outputGenFile, err := os.Create(*outputFileName)
	if err != nil {
		fmt.Printf("Error: unable to generate GPIO config file!\n")
		os.Exit(1)
	}

	defer inputRegDumpFile.Close()
	defer outputGenFile.Close()

	config.OutputGenFile = outputGenFile
	config.InputRegDumpFile = inputRegDumpFile

	parser := parser.ParserData{Template: *template}
	parser.Parse()

	// create dir for output files
	err = os.MkdirAll("generate", os.ModePerm)
	if err != nil {
		fmt.Printf("Error! Can not create a directory for the generated files!\n")
		os.Exit(1)
	}

	// gpio.h
	err = generateOutputFile(&parser)
	if err != nil {
		fmt.Printf("Error! Can not create the file with GPIO configuration!\n")
		os.Exit(1)
	}
}
