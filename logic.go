package main

import (
	"math/rand"
)

// QuestionPool stores questions organized by difficulty level
var QuestionPool = make(map[int][]Question)

// QuestionByID stores questions indexed by ID for quick lookup
var QuestionByID = make(map[int]Question)

func SeedQuestions() {
	questions := GetSeededQuestions()
	for _, q := range questions {
		QuestionPool[q.Difficulty] = append(QuestionPool[q.Difficulty], q)
		QuestionByID[q.ID] = q
	}
}

func GetNextQuestionForUser(currentDifficulty int) Question {

	// Ensure difficulty is within 1-10
	if currentDifficulty < 1 {
		currentDifficulty = 1
	}
	if currentDifficulty > 10 {
		currentDifficulty = 10
	}

	questions, exists := QuestionPool[currentDifficulty]
	if !exists || len(questions) == 0 {
		return QuestionPool[1][0] // Fallback
	}

	return questions[rand.Intn(len(questions))]
}

// GetSeededQuestions returns the static list of 10 questions
func GetSeededQuestions() []Question {
	return []Question{
		{ID: 1, Difficulty: 1, Question: "What is the capital of France?", Options: []string{"Berlin", "Paris", "Madrid", "Rome"}, Answer: "B"},
		{ID: 2, Difficulty: 2, Question: "Which planet is known as the Red Planet?", Options: []string{"Earth", "Venus", "Mars", "Jupiter"}, Answer: "C"},
		{ID: 3, Difficulty: 3, Question: "What is 15 multiplied by 4?", Options: []string{"50", "60", "70", "80"}, Answer: "B"},
		{ID: 4, Difficulty: 4, Question: "Which element has the chemical symbol 'O'?", Options: []string{"Gold", "Silver", "Oxygen", "Iron"}, Answer: "C"},
		{ID: 5, Difficulty: 5, Question: "Who painted the Mona Lisa?", Options: []string{"Van Gogh", "Picasso", "Da Vinci", "Monet"}, Answer: "C"},
		{ID: 6, Difficulty: 6, Question: "What is the square root of 144?", Options: []string{"10", "11", "12", "14"}, Answer: "C"},
		{ID: 7, Difficulty: 7, Question: "Which continent is the Sahara Desert located in?", Options: []string{"Asia", "Africa", "South America", "Australia"}, Answer: "B"},
		{ID: 8, Difficulty: 8, Question: "In what year did the Titanic sink?", Options: []string{"1905", "1912", "1918", "1922"}, Answer: "B"},
		{ID: 9, Difficulty: 9, Question: "What is the largest organ in the human body?", Options: []string{"Heart", "Liver", "Skin", "Lungs"}, Answer: "C"},
		{ID: 10, Difficulty: 10, Question: "Which physicist developed the theory of General Relativity?", Options: []string{"Newton", "Bohr", "Einstein", "Hawking"}, Answer: "C"},
	}
}
