package main

import (
	"flag"
	"fmt"
	"os"
)

import "./parser"

// HdrInfoAdd - adds license header to file f
func HdrInfoAdd(f *os.File) {
	f.WriteString(`/*
 * This file is part of the coreboot project.
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; version 2 of
 * the License.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 */

`)
}

// CreateHdrFile - generates include file
func CreateHdrFile() (err error) {
	hrdFile, err := os.Create("generate/gpio.h")
	if err != nil {
		fmt.Printf("Error!\n")
		return err
	}
	defer hrdFile.Close()

	HdrInfoAdd(hrdFile)
	hrdFile.WriteString(`#ifndef PCH_GPIO_H
#define PCH_GPIO_H

#include <soc/gpe.h>
#include <soc/gpio.h>

const struct pad_config *get_gpio_table(size_t *num);
const struct pad_config *get_early_gpio_table(size_t *num);

#endif /* PCH_GPIO_H */
`)
	return nil
}

// CreateGpioFile - generates gpio_raw.c file
// parser          : parser data structure
// showRawDataFlag : raw data flag
//                   in the case when this flag is false, pad information will
//                   be create as macro
func CreateGpioFile(parser *parser.ParserData, showRawDataFlag bool) (err error) {
	var name = "generate/gpio"
	if showRawDataFlag {
		name += "_raw"
	}
	name += ".c"
	gpio, err := os.Create(name)
	if err != nil {
		fmt.Printf("Error!\n")
		return err
	}
	defer gpio.Close()

	HdrInfoAdd(gpio)
	gpio.WriteString(`
#include <commonlib/helpers.h>
#include "include/gpio.h"
`)
	// Add the pads map to gpio.h file
	parser.PadMapFprint(gpio, showRawDataFlag)
	return nil
}

// main
func main() {
	// Command line arguments
	wordPtr := flag.String("file",
		"inteltool.log",
		"the path to the inteltool log file")
	dbgPtr := flag.Bool("dbg", false, "debug flag")
	flag.Parse()

	fmt.Println("dbg:", *dbgPtr)
	fmt.Println("file:", *wordPtr)

	var parser parser.ParserData
	parser.DbgFlag = *dbgPtr
	err := parser.Parse(*wordPtr)
	if err != nil {
		fmt.Printf("Parser: Error!\n")
		os.Exit(1)
	}

	// create dir for output files
	err = os.MkdirAll("generate", os.ModePerm)
	if err != nil {
		fmt.Printf("Create a directory of generated files: Error!\n")
		os.Exit(1)
	}

	// gpio.h
	err = CreateHdrFile()
	if err != nil {
		fmt.Printf("Create pch_gpio.h: Error!\n")
		os.Exit(1)
	}

	// gpio_raw.c
	err = CreateGpioFile(&parser, true)
	if err != nil {
		fmt.Printf("Create gpio_raw.c: Error!\n")
		os.Exit(1)
	}

	// gpio.c with macros
	err = CreateGpioFile(&parser, false)
	if err != nil {
		fmt.Printf("Create gpio.c: Error!\n")
		os.Exit(1)
	}
}
