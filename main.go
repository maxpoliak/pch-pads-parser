package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

import "./sunrise"

// Information about pad
// id       : pad id string
// offset   : the offset of the register address relative to the base
// function : the string that means the pad function
// data     : dw configuration register data
type padInfo struct {
	id       string
	offset   uint16
	function string
	dw       [sunrise.MAX_DW]uint32
}

// Add information about pad to data structure
// line : string from inteltool log file
func (info *padInfo) Add(line string) {
	var val uint64
	// ------- GPIO Group GPP_A -------
	if strings.HasPrefix(line, "----") {
		// Add header to define GPIO group
		info.function = line
		return
	}
	// 0x0520: 0x0000003c44000600 GPP_B12  SLP_S0#
	// 0x0438: 0xffffffffffffffff GPP_C7   RESERVED
	fmt.Sscanf(line,
		"0x%x: 0x%x %s %s",
		&info.offset,
		&val,
		&info.id,
		&info.function)
	info.dw[0] = uint32(val & 0xffffffff)
	info.dw[1] = uint32(val >> 32)
}

// Print GPIO group title to file
// gpio : gpio.c file descriptor
func (info *padInfo) TitleFprint(gpio *os.File) {
	fmt.Fprintf(gpio, "\n\t/* %s */\n", info.function)
}

// Print Reserved GPIO to file as comment
// gpio : gpio.c file descriptor
func (info *padInfo) ReservedFprint(gpio *os.File) {
	// small comment about reserved port
	fmt.Fprintf(gpio, "\t/* %s - %s */\n", info.id, info.function)
}

// Print information about current pad to file using raw format:
// _PAD_CFG_STRUCT(GPP_F1, 0x84000502, 0x00003026), /* SATAXPCIE4 */
// gpio : gpio.c file descriptor
func (info *padInfo) FprintPadInfoRaw(gpio *os.File) {
	fmt.Fprintf(gpio,
		"\t_PAD_CFG_STRUCT(%s, 0x%0.8x, 0x%0.8x), /* %s */\n",
		info.id,
		info.dw[0],
		(info.dw[1] & 0xffffff00), // Interrupt Select - RO
		info.function)
}

// Print information about current pad to file using special macros:
// PAD_CFG_NF(GPP_F1, 20K_PU, PLTRST, NF1), /* SATAXPCIE4 */
// gpio : gpio.c file descriptor
func (info *padInfo) FprintPadInfoMacro(gpio *os.File) {
	fmt.Fprintf(gpio, "\t/* %s - %s */\n\t%s\n",
		info.id,
		info.function,
		sunrise.GetMacro(info.id, info.dw[0], info.dw[1]))
}

// InteltoolData
// padmap  : pad info map
// dbgFlag : gebug flag, currently not used
type InteltoolData struct {
	padmap  []padInfo
	dbgFlag bool
}

// Adds a new entry to pad info map
// line - string/line from the inteltool log file
func (inteltool *InteltoolData) AddEntry(line string) {
	var pad padInfo
	pad.Add(line)
	inteltool.padmap = append(inteltool.padmap, pad)
}

// Print pad info map to file
// gpio : gpio.c descriptor file
// raw  : in the case when this flag is false, pad information will be print
//        as macro
func (inteltool *InteltoolData) PadMapFprint(gpio *os.File, raw bool) {
	gpio.WriteString("\n/* Pad configuration in ramstage */\n")
	gpio.WriteString("static const struct pad_config gpio_table[] = {\n")
	for _, pad := range inteltool.padmap {
		switch pad.dw[0] {
		case 0:
			pad.TitleFprint(gpio)
		case 0xffffffff:
			pad.ReservedFprint(gpio)
		default:
			if raw {
				pad.FprintPadInfoRaw(gpio)
			} else {
				pad.FprintPadInfoMacro(gpio)
			}
		}
	}
	gpio.WriteString("};\n")

	// FIXME: need to add early configuration
	gpio.WriteString(`/* Early pad configuration in romstage. */
static const struct pad_config early_gpio_table[] = {
	/* TODO: Add early pad configuration */
};

const struct pad_config *get_gpio_table(size_t *num)
{
	*num = ARRAY_SIZE(gpio_table);
	return gpio_table;
}

const struct pad_config *get_early_gpio_table(size_t *num)
{
	*num = ARRAY_SIZE(early_gpio_table);
	return early_gpio_table;
}

`)
}

// Parse pads groupe information in the inteltool log file
// logFile : name of inteltool log file
// return
// err : error
func (inteltool *InteltoolData) Parse(logFile string) (err error) {
	file, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read all lines from inteltool log file
	fmt.Println("Parse IntelTool Log File...")
	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		// Use only the string that contains the GPP information
		if !strings.Contains(line, "GPP_") && !strings.Contains(line, "GPD") {
			continue
		}
		inteltool.AddEntry(line)
	}
	fmt.Println("...done!")
	return nil
}

// HdrInfoAdd adds license header to file
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

// Generates include file
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

// Generates gpio_raw.c file
// inteltool       : inteltool data structure
// showRawDataFlag : raw data flag
//                   in the case when this flag is false, pad information will
//                   be create as macro
func CreateGpioFile(inteltool *InteltoolData, showRawDataFlag bool) (err error) {
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
	inteltool.PadMapFprint(gpio, showRawDataFlag)
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

	var inteltool InteltoolData
	inteltool.dbgFlag = *dbgPtr
	err := inteltool.Parse(*wordPtr)
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
	err = CreateGpioFile(&inteltool, true)
	if err != nil {
		fmt.Printf("Create gpio_raw.c: Error!\n")
		os.Exit(1)
	}

	// gpio.c with macros
	err = CreateGpioFile(&inteltool, false)
	if err != nil {
		fmt.Printf("Create gpio.c: Error!\n")
		os.Exit(1)
	}
}
