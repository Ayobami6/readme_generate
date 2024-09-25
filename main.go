package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Ayobami6/readme_generate/config"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/google/generative-ai-go/genai"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	showBanner()
	startHuhForm(ctx)

}

func generate(ctx context.Context, prompt string) {

	apiKey := config.GetEnv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	defer client.Close()

	// fmt.Printf("this is the prompt from huh %s\n", prompt)

	model := client.GenerativeModel("gemini-1.5-pro")

	model.SetTemperature(1)
	model.SetTopK(64)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(8192)
	model.ResponseMIMEType = "text/plain"
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text("Hey there, you are a helpful assistant that helps write readme documention for my github projects based on some prompt description")},
	}

	session := model.StartChat()

	resp, err := session.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
	tCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	for _, part := range resp.Candidates[0].Content.Parts {
		writeToFile(tCtx, fmt.Sprintf("%v\n", part))
	}
	// fmt.Println(resp.Candidates)

}

func writeToFile(ctx context.Context, content string) error {
	// check if file exist
	_, err := os.Stat("README.md")
	var file *os.File
	if os.IsNotExist(err) {
		// if file not exist
		// create one
		file, err = os.Create("README.md")
		if err != nil {
			return err
		}
	} else {
		// open file
		file, err = os.OpenFile("README.md", os.O_WRONLY, 0600)
		if err != nil {
			return err
		}

	}
	defer file.Close()

	// handle the ctx
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return write(file, content)
	}

}

func write(file *os.File, content string) error {
	_, err := file.WriteString(content)
	if err != nil {
		return err
	}
	return nil

}

func showBanner() {
	// open banner.txt as file
	file, err := os.Open("banner.txt")
	if err == nil {

		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	} else {
		fmt.Print(`
		██████╗ ███████╗ █████╗ ██████╗ ███╗   ███╗███████╗     ██████╗ ███████╗███╗   ██╗███████╗██████╗  █████╗ ████████╗███████╗
		██╔══██╗██╔════╝██╔══██╗██╔══██╗████╗ ████║██╔════╝    ██╔════╝ ██╔════╝████╗  ██║██╔════╝██╔══██╗██╔══██╗╚══██╔══╝██╔════╝
		██████╔╝█████╗  ███████║██║  ██║██╔████╔██║█████╗      ██║  ███╗█████╗  ██╔██╗ ██║█████╗  ██████╔╝███████║   ██║   █████╗  
		██╔══██╗██╔══╝  ██╔══██║██║  ██║██║╚██╔╝██║██╔══╝      ██║   ██║██╔══╝  ██║╚██╗██║██╔══╝  ██╔══██╗██╔══██║   ██║   ██╔══╝  
		██║  ██║███████╗██║  ██║██████╔╝██║ ╚═╝ ██║███████╗    ╚██████╔╝███████╗██║ ╚████║███████╗██║  ██║██║  ██║   ██║   ███████╗
		╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═════╝ ╚═╝     ╚═╝╚══════╝     ╚═════╝ ╚══════╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚══════╝                                                                                                                         
`)
	}

}

func startHuhForm(ctx context.Context) {
	var prompt string
	var projectTitle string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Whats your project name?").
				Prompt("?").
				Value(&projectTitle),
			huh.NewText().
				Title("Provide your readme description").
				Placeholder("Todo list api backend").
				CharLimit(500).
				Value(&prompt),
		),
	)

	err := form.Run()
	instruction := fmt.Sprintf("Project title is %s with description %s", projectTitle, prompt)
	if err != nil {
		log.Fatal(err)
	}
	action := func() {
		generate(ctx, instruction)
	}

	err = spinner.New().
		Title("Preparing your readme......").
		Action(action).
		Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Readme generated!")
}
