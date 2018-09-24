package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/valtyr/chip-8/font"
	"github.com/valtyr/chip-8/graphics"
)

const screenWidth = 64
const screenHeight = 32
const pixelCount = screenWidth * screenHeight

var (
	opcode          uint16
	memory          [4096]byte
	vRegisters      [16]byte
	iRegister       uint16
	programCounter  uint16
	graphicsBuffer  [pixelCount]bool
	delayTimer      byte
	soundTimer      byte
	stack           [16]uint16
	stackPointer    uint16
	keypadRegisters [16]bool
)

func printExplanation(explanation string) {
	fmt.Printf("%x\t%s\n", opcode, explanation)
}

func initialize() {
	// Load FontData into memory
	for i := range font.FontData {
		memory[i] = font.FontData[i]
	}

	graphics.Init()
	programCounter = 512
}

func loadProgram() {
	// READ PROGRAM INTO MEMORY
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		fmt.Println("The CHIP-8 emulator accepts programs through stdin.")
		fmt.Println("Usage: ./chip-8 < ./program.rom")
		os.Exit(2)
	}
	if info.Size() > 4096 {
		fmt.Println("The program you provided was too large.")
		fmt.Println("The CHIP-8 memory is limited to 4096 bytes.")
		os.Exit(2)
	}

	var i = 0
	reader := bufio.NewReader(os.Stdin)
	for {
		inputByte, err := reader.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		memory[i+512] = inputByte
		i++
	}

}

func readIntoKeypadRegister() {
	keypadRegisters = graphics.GetKeyboardState()
}

func fetchOpcode() {
	var firstHalfShifted = uint16(memory[programCounter]) << 8
	var secondHalf = uint16(memory[programCounter+1])
	opcode = firstHalfShifted | secondHalf
}

func bmc(bitmask uint16, value uint16) bool {
	// Bitmask compare
	return opcode&bitmask == value
}

func incrementCounter() {
	programCounter += 2
}

func skipNextInstruction() {
	programCounter += 4
}

func renderFromBuffer() {
	graphics.RenderBuffer(graphicsBuffer[:], 64, 16)

	// RENDERING IN TERMINAL
	// for i := range graphicsBuffer {
	// 	if i%screenWidth == 0 {
	// 		fmt.Print("\n")
	// 	}
	// 	if graphicsBuffer[i] {
	// 		fmt.Print("█")
	// 	} else {
	// 		fmt.Print(" ")
	// 	}
	// }
}

func performOperation() {
	if opcode == 0x00E0 {
		printExplanation("CLEAR")
		// Clears screen.
		for i := range graphicsBuffer {
			graphicsBuffer[i] = false
		}
		incrementCounter()
		return
	}
	if opcode == 0x00EE {
		printExplanation("RETURN")
		// Returns from a subroutine.
		var address = stack[stackPointer-1]
		stackPointer--
		programCounter = address + 2
		return
	}
	if bmc(0xF000, 0x1000) {
		printExplanation("JUMP")
		// Jump to address
		var address = opcode & 0x0FFF
		programCounter = address
		return
	}
	if bmc(0xF000, 0x2000) {
		printExplanation("SUBROUTINE")
		var address = opcode & 0x0FFF
		// Call subroutine at address
		stack[stackPointer] = programCounter
		stackPointer++
		programCounter = address
		return
	}
	if bmc(0xF000, 0x3000) {
		printExplanation("SKIP IF")
		// Skip next instruction if VX == value
		var value = uint8(opcode & 0x00FF)
		var registerIndexX = opcode & 0x0F00 >> 8
		if vRegisters[registerIndexX] == value {
			skipNextInstruction()
		} else {
			incrementCounter()
		}
		return
	}
	if bmc(0xF000, 0x4000) {
		printExplanation("SKIP IF NOT")
		// Skip next instruction if VX != value
		var value = uint8(opcode & 0x00FF)
		var registerIndexX = opcode & 0x0F00 >> 8
		if vRegisters[registerIndexX] != value {
			skipNextInstruction()
		} else {
			incrementCounter()
		}
		return
	}
	if bmc(0xF000, 0x5000) {
		printExplanation("SKIP IF")
		// Skip next instruction if VX == VY
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4
		if vRegisters[registerIndexX] == vRegisters[registerIndexY] {
			skipNextInstruction()
		} else {
			incrementCounter()
		}
		return
	}
	if bmc(0xF000, 0x6000) {
		printExplanation("SET")
		// Sets VX to value
		var registerIndexX = opcode & 0x0F00 >> 8
		var value = uint8(opcode & 0x00FF)
		vRegisters[registerIndexX] = value
		incrementCounter()
		return
	}
	if bmc(0xF000, 0x7000) {
		printExplanation("ADD W/O CARRY")
		// Add value to x (no carry bit)
		var registerIndexX = opcode & 0x0F00 >> 8
		var value = uint8(opcode & 0x00FF)
		vRegisters[registerIndexX] = vRegisters[registerIndexX] + value
		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x8000) {
		printExplanation("COPY")
		// VY -> VX
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4
		vRegisters[registerIndexX] = vRegisters[registerIndexY]
		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x8001) {
		printExplanation("OR")
		// OR VX, VY -> VX
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4
		var result = vRegisters[registerIndexX] | vRegisters[registerIndexY]
		vRegisters[registerIndexX] = result
		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x8002) {
		printExplanation("AND")
		// AND VX, VY -> VX
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4
		var result = vRegisters[registerIndexX] & vRegisters[registerIndexY]
		vRegisters[registerIndexX] = result
		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x8003) {
		printExplanation("XOR")
		// XOR VX, VY -> VX
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4
		var result = vRegisters[registerIndexX] ^ vRegisters[registerIndexY]
		vRegisters[registerIndexX] = result
		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x8004) {
		printExplanation("SUM")
		// VX + VY -> VX
		// 1 -> VF if carry
		// 0 -> VF if no carry
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4

		var rawSum = vRegisters[registerIndexX] + vRegisters[registerIndexY]
		var carryFlag byte
		if rawSum > 0xFF {
			carryFlag = 1
		}
		vRegisters[0xF] = carryFlag
		vRegisters[registerIndexX] = byte(rawSum)

		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x8005) {
		printExplanation("DIFFERENCE")
		// VX - VY -> VX
		// 0 -> VF if borrow
		// 1 -> VF if no borrow
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4

		if vRegisters[registerIndexX] < vRegisters[registerIndexY] {
			vRegisters[0xF] = 1
		} else {
			vRegisters[0xF] = 0
		}
		vRegisters[registerIndexX] = vRegisters[registerIndexX] - vRegisters[registerIndexY]

		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x8006) {
		printExplanation("SHIFT >>")
		// Store least signif. bit of VX in VF
		// Shift VX to the right by 1 (>> 1)
		var registerIndexX = opcode & 0x0F00 >> 8
		vRegisters[0xF] = vRegisters[registerIndexX] & 0x01
		vRegisters[registerIndexX] = vRegisters[registerIndexX] >> 1
		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x8007) {
		printExplanation("DIFFERENCE SWAPPED")
		// VY - VX -> VX
		// 0 -> VF if borrow
		// 1 -> VF if no borrow
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4

		if vRegisters[registerIndexY] < vRegisters[registerIndexX] {
			vRegisters[0xF] = 1
		} else {
			vRegisters[0xF] = 0
		}
		vRegisters[registerIndexX] = vRegisters[registerIndexY] - vRegisters[registerIndexX]

		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x800E) {
		printExplanation("SHIFT <<")
		// Store most signif. bit of VX in VF
		// Shift VX to the left by 1 (<< 1)
		var registerIndexX = opcode & 0x0F00 >> 8
		vRegisters[0xF] = vRegisters[registerIndexX] & 0x80 >> 7
		vRegisters[registerIndexX] = vRegisters[registerIndexX] << 1
		incrementCounter()
		return
	}
	if bmc(0xF00F, 0x9000) {
		printExplanation("SKIP IF NOT EQUAL")
		// Skip next instruction if VX != VY
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4
		if vRegisters[registerIndexX] != vRegisters[registerIndexY] {
			skipNextInstruction()
		} else {
			incrementCounter()
		}
		return
	}
	if bmc(0xF000, 0xA000) {
		printExplanation("WRITE TO I")
		// value -> I register
		var value = opcode & 0x0FFF
		iRegister = value
		incrementCounter()
		return
	}
	if bmc(0xF000, 0xC000) {
		printExplanation("RAND BITMASK")
		// Set register VX to bitwise and of bitmask
		// and random number from 0 to 255
		var registerIndexX = opcode & 0x0F00 >> 8
		var bitmask = uint8(opcode & 0x00FF)
		var random = uint8(rand.Intn(256))
		vRegisters[registerIndexX] = bitmask & random
		incrementCounter()
		return
	}
	if bmc(0xF000, 0xD000) {
		// Draws a sprite at coordinate (VX, VY) that
		// has a width of 8 pixels and a height of
		// N pixels. Each row of 8 pixels is read as
		// bit-coded starting from memory location I;
		// I value doesn’t change after the execution
		// of this instruction. As described above,
		// VF is set to 1 if any screen pixels are
		// flipped from set to unset when the sprite
		// is drawn, and to 0 if that doesn’t happen
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerIndexY = opcode & 0x00F0 >> 4
		var height = int(opcode & 0x000F)

		var xValue = vRegisters[registerIndexX]
		var yValue = vRegisters[registerIndexY]

		var originLocation = int(yValue)*screenWidth + int(xValue)

		var explanation = fmt.Sprintf("DRAW SPRITE AT %d %d; ORIGIN %d", xValue, yValue, originLocation)
		printExplanation(explanation)

		for row := 0; row < height; row++ {
			var memByte = memory[int(iRegister)+row]
			for bitIndex := 7; bitIndex >= 0; bitIndex-- {
				var bit = memByte >> uint8(7-bitIndex) & 1

				var memoryLocation = originLocation + row*screenWidth + int(bitIndex)
				var oldValue = graphicsBuffer[memoryLocation]
				var newValue = bit == 1 || oldValue
				graphicsBuffer[memoryLocation] = newValue
				if oldValue && !newValue {
					vRegisters[0xF] = 1
				} else {
					vRegisters[0xF] = 0
				}
			}
		}

		incrementCounter()
		return
	}
	if bmc(0xF0FF, 0xE09E) {
		printExplanation("SKIP IF KEY")
		// Skip next instruction if key pressed in
		// VX is pressed.
		var registerIndexX = opcode & 0x0F00 >> 8
		var keypadNumber = vRegisters[registerIndexX]
		if keypadRegisters[keypadNumber] {
			skipNextInstruction()
		} else {
			incrementCounter()
		}
		return
	}
	if bmc(0xF0FF, 0xE0A1) {
		// Skip next instruction if key with code in
		// VX is not pressed.
		var registerIndexX = opcode & 0x0F00 >> 8
		var keypadNumber = vRegisters[registerIndexX]
		if !keypadRegisters[keypadNumber] {
			skipNextInstruction()
		} else {
			incrementCounter()
		}
		var explanation = fmt.Sprintf("SKIP IF NOT KEY %d", keypadNumber)
		printExplanation(explanation)
		return
	}
	if bmc(0xF0FF, 0xF007) {
		printExplanation("WRITE DELAY TIMER TO REG")
		// Set VX to value of delay timer
		var registerIndexX = opcode & 0x0F00 >> 8
		vRegisters[registerIndexX] = delayTimer
		incrementCounter()
		return
	}
	if bmc(0xF0FF, 0xF00A) {
		printExplanation("AWAIT KEYPRESS")
		// A key press is awaited, and then stored in
		// VX. (Blocking Operation. All instruction
		// halted until next key event)
		var registerIndexX = opcode & 0x0F00 >> 8
		var searching = true
		for searching {
			readIntoKeypadRegister()
			for keyIndex := range keypadRegisters {
				if keypadRegisters[keyIndex] {
					vRegisters[registerIndexX] = uint8(keyIndex)
					searching = false
					break
				}
			}
		}
		incrementCounter()
		return
	}
	if bmc(0xF0FF, 0xF015) {
		printExplanation("SET DELAY")
		// Sets the delay timer to VX.
		var registerIndexX = opcode & 0x0F00 >> 8
		delayTimer = vRegisters[registerIndexX]
		incrementCounter()
		return
	}
	if bmc(0xF0FF, 0xF018) {
		printExplanation("SET SOUND")
		// Sets the sound timer to VX.
		var registerIndexX = opcode & 0x0F00 >> 8
		soundTimer = vRegisters[registerIndexX]
		incrementCounter()
		return
	}
	if bmc(0xF0FF, 0xF01E) {
		printExplanation("ADD REG TO I")
		// Adds VX to I.
		var registerIndexX = opcode & 0x0F00 >> 8
		iRegister = uint16(vRegisters[registerIndexX]) + iRegister
		incrementCounter()
		return
	}
	if bmc(0xF0FF, 0xF029) {
		printExplanation("SPRITELOC")
		// Set I to location of sprite for char in VX.
		// Characters 0-F (in hexadecimal) are
		// represented by a 4x5 font.
		var registerIndexX = opcode & 0x0F00 >> 8
		iRegister = uint16(vRegisters[registerIndexX]) * 5
		incrementCounter()
		return
	}
	if bmc(0xF0FF, 0xF033) {
		printExplanation("DIGIT WRITE")
		// Stores the binary-coded decimal representation of
		// VX, with the most significant of three digits at
		// the address in I, the middle digit at I plus 1,
		// and the least significant digit at I plus 2. (In
		// other words, take the decimal representation of
		// VX, place the hundreds digit in memory at
		// location in I, the tens digit at location I+1,
		// and the ones digit at location I+2.)

		var address = int(iRegister)
		var registerIndexX = opcode & 0x0F00 >> 8
		var registerValue = vRegisters[registerIndexX]

		memory[address] = registerValue / 100 % 10
		memory[address+1] = registerValue / 10 % 10
		memory[address+2] = registerValue % 10

		return
	}
	if bmc(0xF0FF, 0xF055) {
		printExplanation("STORE TO MEM")
		// Stores V0 to VX (including VX) in memory starting
		// at address I. The offset from I is increased by 1
		// for each value written, but I itself is
		// left unmodified.
		var registerIndexX = opcode & 0x0F00 >> 8
		var iRegisterValue = int(iRegister)
		for i := 0; i > int(registerIndexX); i++ {
			memory[iRegisterValue+i] = uint8(vRegisters[i])
		}
		incrementCounter()
		return
	}
	if bmc(0xF0FF, 0xF065) {
		printExplanation("RAND BITMASK")
		// Fills V0 to VX (including VX) with values from
		// memory starting at address I. The offset from I
		// is increased by 1 for each value written, but I
		// itself is left unmodified.
		var registerIndexX = opcode & 0x0F00 >> 8
		var iRegisterValue = int(iRegister)
		for i := 0; i > int(registerIndexX); i++ {
			vRegisters[i] = memory[iRegisterValue+i]
		}
		incrementCounter()
		return
	}
}

func emulateCycle() {
	fetchOpcode()
	performOperation()
	renderFromBuffer()
	readIntoKeypadRegister()

	// BEEP IF AT 0

	if delayTimer > 0 {
		delayTimer--
	}
	if soundTimer > 0 {
		soundTimer--
	}
}

func main() {
	loadProgram()
	initialize()
	defer graphics.Teardown()

	fmt.Println("CHIP-8 baby!")

	for t := range time.NewTicker(17 * time.Millisecond).C {
		// for t := range time.NewTicker(200 * time.Millisecond).C {
		if false {
			fmt.Println(t)
		}
		emulateCycle()
	}

}
