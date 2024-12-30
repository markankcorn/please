# please
command line tool written in Go to use Gemini for help with command syntax

Based on the [please.bash](https://github.com/pmarreck/dotfiles/blob/master/bin/functions/please.bash) project from [Peter Marreck](https://github.com/pmarreck) I asked Gemini to write a Go executable program to get assistance on commands for Zshell. It reads the previous ten commands from the `.zsh_history` file but we can modify the program later using the `HISTFILE` environment variable. 

This is my attempt to replace some of the functionality of the [Warp](warp.dev) terminal emulator, which does a very fine job of getting me out of syntax frustration, but it's a massive memory hog. It's also not yet (Dec 2024) available on Windows and I very much enjoy using my GPUs on my gaming pc for running local models. I'd like to be able to use this both on my MacBook and on my Win11 rig running Ubuntu via WSL.

The tool will ask Gemini to present three suggestions for commands to the user, who can use the Tab key to cycle through the suggestions, Enter to accept and run a suggestion, or Escape key to reject all and return to the command line.

### Install Go
First off, we need to install Go if you don't already have it, along with 

`sudo apt install golang-go`

then verify that we did it right

`go version`

Now let's make a director for go and our tools, then export it to our path variable

`mkdir -p ~/go && cd ~/go && echo "export GOPATH=$HOME/go" >> ~/.zshrc`

Again, let's make sure we did this right.

`echo $GOPATH`

Now let's install our tools

`go install golang.org/x/tools/cmd/goimports@latest`

Make a directory for the project, I just called it Please

`mkdir please && cd please`

The initialize a new Go module to create a `go.mod` file which will track the project's dependencies.

`go mod init please`

Now for our first dependency, the Google AI Go SDK. T his will add the necessary packages to the `go.mod` and `go.sum` files.

`go get github.com/google/generative-ai-go`

To handle the Tab / Enter / Escape functionality, we'll need a third-party library.

`go get github.com/eiannone/keyboard`

We'll also need to get an API key from Google. Go (here)[https://aistudio.google.com/app/apikey] and fetch one and be sure to edit your `go.main` file to include this. You can also change the history look-back to be any number of commands in the `/zsh_history` using the `n` variable in line 65 of the `main.go` file.

Once that's all set up, let's download and build the program.

`go build`

To run the program, use `./ please <your request>`
