package main

import poster "github.com/alejandropriv/poster/postwriter"

var researchStepConfig = poster.StepConfig{
	LLMProvider: "openai",
	LLMModel:    "gpt-4o-mini",
	PromptTemplate: `Today is {{.date}}.
	Write an article about: {{.topic}}, use only information relevant for Date: {{.date}}
  You are an experienced cloud architect and kubernetes engineer.
  You identify trends, best practices, statistics and practical insights.
  You identify production pain points.
  - Use the provided web-searches.
  - Return at most 10 bullet points:
  - focus on 1 key concept at most 3 if referred to the first 1
  - at most 10 common mistake
  - at most 10 practical tip
  Expected Output: Research concise notes with 5 bullets only and reference links`,
}

var writeStepConfig = poster.StepConfig{
	LLMProvider: "openai",
	LLMModel:    "gpt-4.1-mini",
	InputKey:    "research",
	OutputKey:   "draft",
	PromptTemplate: `Today is {{.date}}.
		you are a writer
		you write engaging LinkedIn posts.
    You are an expert technical writer.
    - Your audience is DevOps Engineers, Platform Engineers and CTOs.
		- Your writing is practical, conversational, educational.
    - Your final goal is to generate engagement and discussion on LinkedIn.
    - Your ultimate purpose is to help the reader solve a problem and sell a consultancy service.
    Using the research, write a LinkedIn post.
    Requirements:
    - 250 words
    - conversational
    - technical
    - hook
    - short paragraphs
    - end with a question
    Topic: {topic}
  	expected_output: Draft LinkedIn post.
 `,
}


var optimizeStepConfig = poster.StepConfig{
	LLMProvider: "openai",
	LLMModel:    "gpt-4o",
	InputKey:    "draft",
	OutputKey:   "post",
	PromptTemplate: `
editor:
  role: >
    Editor
  goal: >
    Improve clarity and readability.
  backstory: >
    You improve grammar, flow, formatting and remove unnecessary words.
    you rephrase sentences to avoid ai writing style and make the text more human-like.
  llm: gpt-4.1-mini

editing_task:
  description: >
    Review the LinkedIn post.
    - ai generated text should be reworded to sound more human and natural
    - Avoid using pretentious or fancy words, keep it simple and clear.
    - Keep the post technical based, intead that selling a product or service.
    - Improve readability, grammar, engagement, clarity

    Keep technical accuracy.
  expected_output: >
    Polished LinkedIn post.

optimizer:
  role: >
    LinkedIn growth expert
  goal: >
    Maximize engagement.
  backstory: >
    You specialize in viral LinkedIn content.
    You create:
    - hashtags
    - CTA
    - image ideas
    - stronger hooks
  llm: gpt-4.1-mini
	optimization_task:
  description: >
    Take the final LinkedIn post.

    Produce:
    1. Final LinkedIn post
    2. Five hashtags
    3. One detailed image-generation prompt
    4. One CTA suggestion
    Format simple, Reduce emojis to the maximum of 2 preferable none

  expected_output: >
    Markdown containing
    - LinkedIn Post
    - Hashtags
    - Image Prompt
    - CTA

	`,
}

var imageStepConfig = poster.StepConfig{
	LLMProvider: "openai",
	LLMModel:    "chatgpt-image-latest",
	PromptTemplate: `Create a polished image to accompany this post. Reflect its main idea accurately and avoid adding text to the image.
		Post:
		{{.post}}`,
	ImageOutputPath: "output_image.png",
}
