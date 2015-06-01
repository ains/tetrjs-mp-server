package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)
import "bufio"
import "os"
import "strconv"
import "unicode"
import "github.com/ains/gotetris"

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter piece: ")
	pieceStr, _ := reader.ReadString('\n')
	pieceRune, _ := utf8.DecodeRuneInString(pieceStr)

	fmt.Println()

	fmt.Print("shift: ")
	shiftStr, _ := reader.ReadString('\n')
	shift, _ := strconv.Atoi(strings.TrimSpace(shiftStr))

	fmt.Print("rot: ")
	rotStr, _ := reader.ReadString('\n')
	rot, _ := strconv.Atoi(strings.TrimSpace(rotStr))

	piece := gotetris.PieceMap[unicode.ToUpper(pieceRune)]
	g := gotetris.Game{}
	game := gotetris.DropPiece(g, piece, shift, rot)
	game.OutputBoard()
}
