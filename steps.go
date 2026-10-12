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
	LLMModel:    "gpt-4o",
	PromptTemplate: `Write a post from the research below.

Research:
{{.research}}

Write a clear, useful post for cloud and Kubernetes practitioners. Preserve factual accuracy `,
}

var optimizeStepConfig = poster.StepConfig{
	LLMProvider: "openai",
	LLMModel:    "gpt-4o",
	PromptTemplate: `Improve the draft below for clarity, flow, and concision without changing its meaning or adding unsupported facts.

Draft:
{{.draft}}

Return only the optimized post.`,
}

var imageStepConfig = poster.StepConfig{
	LLMProvider: "openai",
	LLMModel:    "chatgpt-image-latest",
	PromptTemplate: `Create a polished image to accompany this post. Reflect its main idea accurately and avoid adding text to the image.
		Post:
		{{.post}}`,
	ImageOutputPath: "output_image.png",
}
