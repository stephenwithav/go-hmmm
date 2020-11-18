# hmmm

A TUI to simplify my [#PapersThatMakeYouGoHmmm](https://twitter.com/search?q=%23PapersThatMakeYouGoHmmm) browsing of weekly AI research papers.

Copyright 2020, Steven Edwards
steven@stephenwithav.com

## About the Project

Package hmmm provides a simple interface to easily browse the latest arXiv papers and, if necessary, generate links to the arXiv and ScienceWise abstracts.

## About the CLI

`hmmm` is a TUI to automate the posting of selected papers to Twitter.

`hmmm` requires a `config.yaml` file in either the current directory or `$HOME/.hmmm`.

This file MUST contain your Twitter OAuth credentials.  (Create [here](https://developer.twitter.com/en/apps).)

A [sample YAML configuration](https://github.com/stephenwithav/go-hmmm/hmmm/config.yaml) file is provided.  (JSON and TOML, along with environment variable variants, are also supported.)
