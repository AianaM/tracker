package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const tokenEnvName = "YANDEX_IAM_TOKEN"

var envVars struct {
	token string
}

func init() {
	envVars.token = os.Getenv(tokenEnvName)
	if envVars.token == "" {
		askToken()
	}
}

func getToken() string {
	return envVars.token
}

func askToken() {
	fmt.Println("Необходим IAM-токен для федеративного аккаунта")
	fmt.Println("Если у вас еще нет интерфейса командной строки Yandex Cloud, установите и инициализируйте его. https://yandex.cloud/ru/docs/cli/quickstart#install")
	fmt.Println("\n")
	fmt.Println("Получите IAM-токен: yc iam create-token")
	fmt.Println("Нажми Y, чтобы получить токен")
	fmt.Println("Нажми W, чтобы ввести токен вручную")
	fmt.Println("Нажми N, чтобы выйти")
	fmt.Print("Вводи: ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	switch input {
	case "Y", "y":
		saveToken(createToken())
		break
	case "W", "w":
		saveToken(inputToken())
		break
	case "N", "n":
		fmt.Println("=(")
		os.Exit(0)
		break
	}
}

func inputToken() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Введите токен: ")
	for {
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if text != "" {
			return text
		}

		fmt.Println("=(")
		os.Exit(0)
	}
}

func createToken() string {
	cmd := exec.Command("yc", "iam", "create-token")
	stdout, err := cmd.Output()

	if err != nil {
		log.Fatal(err.Error())
	}

	return strings.Replace(string(stdout), "\n", "", -1)
}

func saveToken(value string) {
	fmt.Println("Сохраняю токен: ", value)
	envVars.token = value
	os.Setenv(tokenEnvName, value)
}
