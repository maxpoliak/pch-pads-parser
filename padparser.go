package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

type padInfo struct {
	offset   uint16
	id       string
	function string
	dw0      uint32
	dw1      uint32
}

type InteltoolData struct {
	padmap  []padInfo
	dbgFlag bool
}

func (info *padInfo) Add(line string) {
	var val uint64
	/* ------- GPIO Group GPP_A ------- */
	if strings.HasPrefix(line, "----") {
		/* Add header to define GPIO group */
		info.function = line
		return
	}
	/* 0x0520: 0x0000003c44000600 GPP_B12  SLP_S0#  */
	/* 0x0438: 0xffffffffffffffff GPP_C7   RESERVED */
	fmt.Sscanf(line,
		"0x%x: 0x%x %s %s",
		&info.offset,
		&val,
		&info.id,
		&info.function)

	info.dw1 = uint32(val >> 32)
	info.dw0 = uint32(val & 0xffffffff)
}

func (info *padInfo) TitleFprint(gpio *os.File) {
	fmt.Fprintf(gpio, "\t/* %s */\n", info.function)
}

func (info *padInfo) ReservedFprint(gpio *os.File) {
	/* small comment about reserved port */
	fmt.Fprintf(gpio, "\t/* %s */\n", info.function)
}

func (info *padInfo) Fprint(gpio *os.File) {
	fmt.Fprintf(gpio,
		"\tPCH_PAD_DW0_DW1_CFG(%s, 0x%0.8x, 0x%0.8x), /* %s */\n",
		info.id,
		info.dw0,
		info.dw1,
		info.function)
}

func (inteltool *InteltoolData) AddEntry(line string) {
	var pad padInfo
	pad.Add(line)
	inteltool.padmap = append(inteltool.padmap, pad)
}

func (inteltool *InteltoolData) PadMapFprint(gpio *os.File) {
	gpio.WriteString("\n/* Pad configuration in ramstage */\n")
	gpio.WriteString("static const struct pad_config gpio_table[] = {\n")
	for _, pad := range inteltool.padmap {
		switch pad.dw0 {
		case 0:
			pad.TitleFprint(gpio)
		case 0xffffffff:
			pad.ReservedFprint(gpio)
		default:
			pad.Fprint(gpio)
		}
	}
	gpio.WriteString("};\n")

	/* FIXME: need to add early configuration */
	gpio.WriteString("\n/* Early pad configuration in romstage. */\n")
	gpio.WriteString("static const struct pad_config early_gpio_table[] = {\n")
	gpio.WriteString("\t/* TODO: Add early pad configuration */\n")
	gpio.WriteString("};\n")
}

func (inteltool *InteltoolData) Parse(logFile string) (err error) {
	file, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer file.Close()

	/* Read all lines from inteltool log file */
	fmt.Println("Parse IntelTool Log File...")
	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		/* Use only the string that contains the GPP information */
		if !strings.Contains(line, "GPP_") &&
			!strings.Contains(line, "GPD") {
			continue
		}
		inteltool.AddEntry(line)
	}
	fmt.Println("...done!")
	return nil
}

func HdrInfoAdd(gpio *os.File) {
	gpio.WriteString(`/*
 * This file is part of the coreboot project.
 *
 * Copyright (C) 2019 Maxim Polyakov <max.senia.poliak@gmail.com>
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

#ifndef _GPIO_H
#define _GPIO_H

#include <soc/gpe.h>
#include <soc/gpio.h>

#define PCH_PAD_DW0_DW1_CFG(val, config0, config1)  \
		_PAD_CFG_STRUCT(val, config0, config1)
`)
}

func PostfixAdd(gpio *os.File) {
	gpio.WriteString("\n#endif\n")
}

func main() {
	/* Command line arguments */
	wordPtr := flag.String("file",
		"inteltool.log",
		"the path to the inteltool log file")
	dbgPtr := flag.Bool("dbg", false, "debug flag")
	flag.Parse()

	fmt.Println("d:", *dbgPtr)
	fmt.Println("f:", *wordPtr)

	var inteltool InteltoolData
	inteltool.dbgFlag = *dbgPtr
	err := inteltool.Parse(*wordPtr)
	if err != nil {
		fmt.Printf("Parser: Error!\n")
		os.Exit(1)
	}

	gpio, err := os.Create("gpio.h")
	if err != nil {
		fmt.Printf("Error!\n")
		os.Exit(1)
	}
	defer gpio.Close()

	HdrInfoAdd(gpio)
	/* Add the pads map to gpio.h file */
	inteltool.PadMapFprint(gpio)
	PostfixAdd(gpio)
}
