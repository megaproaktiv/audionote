package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const model = "anthropic.claude-3-5-sonnet-20240620-v1:0"

func CallBedrock(prompt string) (string, error) {
	fmt.Printf("Calling Bedrock model 'anthropic.claude-3-5-sonnet-20240620-v1:0'...\n")
	// This simulates an API call to Bedrock.
	input := "Processed result from Bedrock with prompt: " + prompt
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("configuration error, " + err.Error())
		os.Exit(1)
	}
	client := bedrockruntime.NewFromConfig(cfg)
	ctx := context.TODO()
	result, err := Converse(ctx, client, input)
	if err != nil {
		return "", err
	}
	return result, nil
}

// Plain Converse
func Converse(ctx context.Context, client *bedrockruntime.Client, input string) (string, error) {

	converseInput := &bedrockruntime.ConverseInput{
		ModelId: aws.String(model),
	}

	// create user message
	userMsg := types.Message{
		Role: types.ConversationRoleUser,
		Content: []types.ContentBlock{
			&types.ContentBlockMemberText{
				Value: input,
			},
		},
	}

	converseInput.Messages = append(converseInput.Messages, userMsg)

	output, err := client.Converse(ctx, converseInput)

	if err != nil {
		fmt.Println("Converse API Call", "error", err)
		os.Exit(1)
	}

	if output == nil {
		fmt.Println("Empty response from Bedrock API")
		os.Exit(2)
	}

	response, ok := output.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		fmt.Println("Unexpected response type from Bedrock API")
		os.Exit(3)
	}

	if len(response.Value.Content) == 0 {
		fmt.Println("Empty content in Bedrock API response")
		os.Exit(4)
	}

	responseContentBlock := response.Value.Content[0]
	text, ok := responseContentBlock.(*types.ContentBlockMemberText)
	if !ok {
		fmt.Println("Unexpected content block type in Bedrock API response")
		os.Exit(5)
	}

	if text.Value == "" {
		fmt.Println("Empty text value in Bedrock API response")
		os.Exit(6)
	}

	return text.Value, nil
}
