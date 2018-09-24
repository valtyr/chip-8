# CHIP-8 emulator

### Overview

In this project I'm trying to hit two birds with one stone:

- Try out Go
- Try writing an emulator

This project's a bit messy right now and only semi-functional. It uses SDL2 for graphics, input and, in the future, sound.

### Building

To build the emulator you need to have installed Go and SDL2 for your platform.
After that building should be as simple as:

```
go build
```

_You might have have to manually go get some of the dependencies I'm not totally sure how package management in go is supposed to work yet hehe._

### Running

I've included a ROM to test the emulator out with. The emulator expects a ROM piped in through STDIN.

```
./chip-8 < SYZYGY.ROM
```

### To do

- [x] Implement all instructions needed for basic functionality
- [x] Add in graphics support
- [x] Add in keyboard input support
- [ ] Iron out kinks and quirks
- [ ] Refactor into a better structured project
