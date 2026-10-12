package main

import poster "github.com/alejandropriv/poster/postwriter"

var researchStepConfig = poster.StepConfig{
	LLMProvider: "openai",
	LLMModel:    "gpt-4o-mini",
	PromptTemplate: `Today is 12.10.2026. What are the latest kubernetes trends? return the sources you are using to answer the question. If you don't know the answer, say so. Do not make up sources.`,
}

var writeStepConfig = poster.StepConfig{
	LLMProvider: "openai",
	LLMModel:    "gpt-4o",
	PromptTemplate: `Write a post from the research below.

Research:
{{.research}}

Write a clear, useful post for cloud and Kubernetes practitioners. Preserve factual accuracy and check this source: https://www.linkedin.com/posts/kubernetes-cloudnative-devops-share-7506414279286218753-Bpb1/`,
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
