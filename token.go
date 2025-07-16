package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const tokenEnvName = "YANDEX_IAM_TOKEN"

var envVars struct {
	token string
}

func getToken() string {
	if envVars.token == "" {
		if token := os.Getenv(tokenEnvName); token != "" {
			envVars.token = token
		} else {
			askToken()
		}
	}
	return envVars.token
}

func askToken() {
	fmt.Println("Необходим IAM-токен для федеративного аккаунта")
	fmt.Println("Если у вас еще нет интерфейса командной строки Yandex Cloud, установите и инициализируйте его. https://yandex.cloud/ru/docs/cli/quickstart#install")
	fmt.Println("Затем выполните команду: yc init --federation-id=<federation_ID>")
	fmt.Println("где <federation_ID> - идентификатор федеративного аккаунта, который вы можете найти в разделе 'Управление федерацией' в Yandex Cloud Console.")
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
	case "W", "w":
		saveToken(inputToken())
	case "N", "n":
		fmt.Println("=(")
		os.Exit(0)
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

func checkEnvFile() string {
	file, err := os.Open(".env")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		re := regexp.MustCompile(tokenEnvName + `=(?P<value>.+)`)
		matches := re.FindStringSubmatch(scanner.Text())
		valueIndex := re.SubexpIndex("value")
		if valueIndex > -1 && valueIndex < len(matches) {
			return matches[valueIndex]
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	return ""
}
