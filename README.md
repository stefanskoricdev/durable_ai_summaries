# Durable AI Summaries

## Overview

This is a demonstration project showcasing an innovative approach to content discovery and summarization. It allows users to engage in a conversational flow to refine a topic, dynamically search for relevant YouTube videos, and then process a selected video to extract its core essence. The application automates the video content consumption process by providing concise, easy-to-read markdown summaries.

## How It Works

The core functionality of this project revolves around a multi-step, user-guided process:

1. Topic Initialization: The user provides an initial topic of interest.
2. Conversational Refinement: The application engages in a conversation with the user to gather more specific details and refine the search criteria.
3. YouTube Video Search: Based on the refined information, the application searches YouTube for relevant videos.
4. User Selection: The user reviews the search results and selects a specific video for processing.
5. Video Processing & Summarization: For the selected video, the application performs the following actions:
   - Downloads the video content.
   - Transcribes the audio into text.
   - Utilizes advanced AI models to summarize the transcription into a clear and concise markdown format.

## Technology Stack

This project leverages the following key technologies:

- GO
- Temporal Workflows
- OpenAI Models

## Project Status

This is strictly a demo project designed to showcase the integration and capabilities of Temporal Workflows and OpenAI models for content processing.

It is explicitly not production-ready. This project currently lacks essential features required for production environments, including but not limited to:

- Robust security measures
- Comprehensive input validation
- Extensive automated tests
- Error handling for all edge cases
- Scalability optimizations

This project should be used for demonstration, learning, and proof-of-concept purposes only.This is a demo project currently under active development.
