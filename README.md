uniconfig
=========
[![Build Status](https://travis-ci.org/Babazka/uniconfig.svg?branch=master)](https://travis-ci.org/Babazka/uniconfig)

Simple golang configuration library which works with INI files, cmdline flags and environment variables.

This package is influenced by https://github.com/rakyll/globalconf, but is different: instead of defining flags for every option, you create a ``struct`` with all required fields, and ``uniconfig`` creates all flags for you.

Installation
------------

	go get github.com/Babazka/uniconfig

Usage
-----

First, import the package:

	import "github.com/Babazka/uniconfig"

Define the structure which will hold your configuration. ``uniconfig`` supports two-level structure nesting.

	type MyConfig struct {
		Debug   bool
		Count   int `help:"number of items"`
		Nested1 struct {
			A       string
			B       string
			ignored string
		}
		Nested2 struct {
			Zzz bool
		}
	}
	...
	config := MyConfig{}
	// set some defaults here...
	config.Count = 42

Pass the config structure to the library:

	uniconfig.Load(&config)

This line will perform the following actions:

1. Check the command line for ``--config`` option and fill the ``config`` structure from *INI file*, if specified. The INI file can look like this:

		debug = true
		count = 65535
		; this is a comment
		# also a comment
		
		[Nested1]
		A  = sometag
		b = something else
		irrelevant = ignored option
		number = 14

2. Overwrite ``config`` from *environment variables*. Environment variables are named in uppercase, like this: ``COUNT=65535 NESTED1_A=sometag``.
3. Generate *command line flags* for each of ``config``'s public fields, parse the command line and overwrite ``config`` with parsed values. Command line options are named in lowercase, like this: ``--count 65535 --nested1-a sometag``. The ``--help`` option will be handled here, printing something like this:

		-config="": path to configuration file
		-count=42: number of items
		-debug=false:
		-nested1-a="":
		-nested1-b="":
		-nested2-zzz=false:

	Note that the help text comes from the structure field tags (e.g. ``Count   int `help:"number of items"` ``).

You can dump the final config in the form of INI file:

	fmt.Println(uniconfig.ConfigAsIniFile(config))

Example code
------------

Clone this repository and try running example application in various ways:

	git clone https://github.com/Babazka/uniconfig
	cd uniconfig

	NESTED1_A=bee go run example_app/example.go --count=45 --config=example_app/testconfig.ini
	NESTED1_A=bee go run example_app/example.go --count=45
	go run example_app/example.go --count=45
