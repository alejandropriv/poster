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
	Requirements:
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
    You are an expert technical writer, Your ultimate purpose is to help the reader solve a problem and sell a consultancy service.
		Write an engaging LinkedIn post based on this research:
		{{.research}}
		Use the research and its links as your main source. The topic is the one covered by the research.
    - Your audience is DevOps Engineers, Platform Engineers and CTOs.
    - Your goal is to generate engagement and discussion on LinkedIn.
		- you rephrase sentences to avoid ai writing style and make the text more human-like.
		- You check grammar, flow, formatting and remove unnecessary words.

    Requirements:
    - 250 words
		- practical
    - conversational
    - technical
		- educational
    - hook
    - short paragraphs
    - end with a question
  	expected_output: Draft LinkedIn post.
	Baseline:
		1. Base factual claims on the research and the source URLs provided.
		2. Never invent sources or claim to have read pages you did not access.
		3. If the search tool fails, report the failure honestly.
 `,
}

var optimizeStepConfig = poster.StepConfig{
	LLMProvider: "openai",
	LLMModel:    "gpt-4o",
	InputKey:    "draft",
	OutputKey:   "post",
	PromptTemplate: `
		Improve the draft below to maximize engagement while preserving its facts and technical accuracy.

		Draft:
		{{.draft}}

  You specialize in viral LinkedIn content.
  You create:
    - hashtags
    - CTA
    - image ideas
    - stronger hooks

  Produce:
    1. Final LinkedIn post ready to be posted
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
	LLMModel:    "gpt-image-2.5-flare",
	PromptTemplate: `
	You are an expert Visual artist and art director:
	You receive an Image prompt for a linkedin image
	You will generate an image for the LinkedIn post, according to the given guidelines.
	Take into account the following concepts:
  - composition
  - lighting
  - camera angle
  - environment
	- visual hierarchy
  - color palette
  - realism
  - style
  - important objects
  - things that must NOT appear
`,
ImageOutputPath: "output_image.png",
}
