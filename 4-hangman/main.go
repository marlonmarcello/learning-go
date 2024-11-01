package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"
)

func main() {
	resp, err := http.Get("https://random-word-api.herokuapp.com/word?number=1")

	if err != nil {
		fmt.Println("Couldn't get a random word from API, using basic stored one...")
		fmt.Println(err)
	}

	if resp.StatusCode >= 300 {
		fmt.Println("Couldn't get a random word from API, not sure why, using basic stored word...")
	}

	defer resp.Body.Close()

	var randomWord []string

	err = json.NewDecoder(resp.Body).Decode(&randomWord)
	if err != nil {
		fmt.Println("Error parsing the request body, using basic stored word...")
		fmt.Println(err)
	}

	word := "golang"
	if len(randomWord) > 0 {
		word = randomWord[0]
	}

	attempts := 6
	currentWordState := initializeCurrentWordState(word)
	scanner := bufio.NewScanner(os.Stdin)
	guessedLetters := make(map[string]bool)

	fmt.Println("Welcome to Hangman!")
	fmt.Println("The word has ", len(word), " letters. Good luck!")

	for attempts > 0 {
		displayHangman(6 - attempts)
		displayCurrentState(currentWordState, attempts)

		userInput := getUserInput(scanner)

		fmt.Println("---------------------------------------------------------------------------------------")

		if !isValidInput(userInput) {
			fmt.Println("Invalid input. Only single letters are accepted")
			continue
		}

		if guessedLetters[userInput] {
			fmt.Println("You already guessed the letter ", userInput)
			continue
		}

		fmt.Println("You guessed: ", userInput)

		guessedLetters[userInput] = true

		correctGuess := updateGuessed(word, currentWordState, userInput)

		if !correctGuess {
			attempts--
			fmt.Println("You guessed wrong! No letter ", userInput)
		}

		if isWordGuessed(currentWordState, word) {
			fmt.Println("Congratulations! You've guessed the word!")
			displayCurrentState(currentWordState, attempts)
			return
		}

		if attempts == 0 {
			fmt.Println("Game over. The word was: ", word)
			return
		}

	}

	displayCurrentState(currentWordState, attempts)

}

func isWordGuessed(guessed []string, word string) bool {
	return strings.Join(guessed, "") == word
}

func displayHangman(incorrectGuesses int) {
	if incorrectGuesses >= 0 && incorrectGuesses < len(hangmanStates) {
		fmt.Println(hangmanStates[incorrectGuesses])
	}
}

func updateGuessed(word string, guessed []string, input string) bool {
	correctGuess := false

	for i, char := range word {
		if string(char) == input {
			guessed[i] = input
			correctGuess = true
		}
	}

	return correctGuess
}

func isValidInput(input string) bool {
	return utf8.RuneCountInString(input) == 1
}

func getUserInput(scanner *bufio.Scanner) string {
	fmt.Println("Type a letter to guess: ")
	scanner.Scan()
	return scanner.Text()
}

func initializeCurrentWordState(word string) []string {
	currentWordState := make([]string, len(word))

	for i := range currentWordState {
		currentWordState[i] = "_"
	}

	return currentWordState
}

func displayCurrentState(currentWordState []string, attempts int) {
	fmt.Println("Current word state:", strings.Join(currentWordState, " "))
	fmt.Println("Attempts left: ", attempts)
}
