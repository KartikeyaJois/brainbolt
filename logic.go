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

// GetSeededQuestions returns the static list of questions (multiple per difficulty level)
func GetSeededQuestions() []Question {
	return []Question{
		// Difficulty 1 - Basic Geography & General Knowledge
		{ID: 1, Difficulty: 1, Question: "What is the capital of France?", Options: []string{"Berlin", "Paris", "Madrid", "Rome"}, Answer: "B"},
		{ID: 11, Difficulty: 1, Question: "What is the largest ocean on Earth?", Options: []string{"Atlantic", "Indian", "Arctic", "Pacific"}, Answer: "D"},
		{ID: 12, Difficulty: 1, Question: "How many continents are there?", Options: []string{"5", "6", "7", "8"}, Answer: "C"},
		{ID: 13, Difficulty: 1, Question: "What is the capital of Japan?", Options: []string{"Seoul", "Beijing", "Tokyo", "Bangkok"}, Answer: "C"},
		{ID: 14, Difficulty: 1, Question: "Which country is known as the Land of the Rising Sun?", Options: []string{"China", "Japan", "Korea", "Thailand"}, Answer: "B"},

		// Difficulty 2 - Basic Science
		{ID: 2, Difficulty: 2, Question: "Which planet is known as the Red Planet?", Options: []string{"Earth", "Venus", "Mars", "Jupiter"}, Answer: "C"},
		{ID: 15, Difficulty: 2, Question: "What is the chemical symbol for water?", Options: []string{"H2O", "CO2", "O2", "NaCl"}, Answer: "A"},
		{ID: 16, Difficulty: 2, Question: "How many legs does a spider have?", Options: []string{"6", "8", "10", "12"}, Answer: "B"},
		{ID: 17, Difficulty: 2, Question: "What is the closest star to Earth?", Options: []string{"Proxima Centauri", "Sirius", "The Sun", "Alpha Centauri"}, Answer: "C"},
		{ID: 18, Difficulty: 2, Question: "What gas do plants absorb from the atmosphere?", Options: []string{"Oxygen", "Nitrogen", "Carbon Dioxide", "Hydrogen"}, Answer: "C"},

		// Difficulty 3 - Basic Math
		{ID: 3, Difficulty: 3, Question: "What is 15 multiplied by 4?", Options: []string{"50", "60", "70", "80"}, Answer: "B"},
		{ID: 19, Difficulty: 3, Question: "What is 25 divided by 5?", Options: []string{"3", "4", "5", "6"}, Answer: "C"},
		{ID: 20, Difficulty: 3, Question: "What is 12 + 18?", Options: []string{"28", "30", "32", "34"}, Answer: "B"},
		{ID: 21, Difficulty: 3, Question: "What is 100 minus 37?", Options: []string{"61", "63", "65", "67"}, Answer: "B"},
		{ID: 22, Difficulty: 3, Question: "What is 7 times 8?", Options: []string{"54", "56", "58", "60"}, Answer: "B"},

		// Difficulty 4 - Science & Chemistry
		{ID: 4, Difficulty: 4, Question: "Which element has the chemical symbol 'O'?", Options: []string{"Gold", "Silver", "Oxygen", "Iron"}, Answer: "C"},
		{ID: 23, Difficulty: 4, Question: "What is the chemical symbol for gold?", Options: []string{"Go", "Gd", "Au", "Ag"}, Answer: "C"},
		{ID: 24, Difficulty: 4, Question: "What is the hardest natural substance on Earth?", Options: []string{"Gold", "Iron", "Diamond", "Platinum"}, Answer: "C"},
		{ID: 25, Difficulty: 4, Question: "What is the most abundant gas in Earth's atmosphere?", Options: []string{"Oxygen", "Carbon Dioxide", "Nitrogen", "Argon"}, Answer: "C"},
		{ID: 26, Difficulty: 4, Question: "What is the freezing point of water in Celsius?", Options: []string{"-10°C", "0°C", "10°C", "32°C"}, Answer: "B"},

		// Difficulty 5 - Art & History
		{ID: 5, Difficulty: 5, Question: "Who painted the Mona Lisa?", Options: []string{"Van Gogh", "Picasso", "Da Vinci", "Monet"}, Answer: "C"},
		{ID: 27, Difficulty: 5, Question: "In which year did World War II end?", Options: []string{"1943", "1944", "1945", "1946"}, Answer: "C"},
		{ID: 28, Difficulty: 5, Question: "Who wrote 'Romeo and Juliet'?", Options: []string{"Charles Dickens", "William Shakespeare", "Jane Austen", "Mark Twain"}, Answer: "B"},
		{ID: 29, Difficulty: 5, Question: "What is the name of the famous tower in Paris?", Options: []string{"Big Ben", "Eiffel Tower", "Leaning Tower", "Statue of Liberty"}, Answer: "B"},
		{ID: 30, Difficulty: 5, Question: "Which ancient wonder was located in Alexandria?", Options: []string{"Hanging Gardens", "Colossus", "Lighthouse", "Pyramids"}, Answer: "C"},

		// Difficulty 6 - Intermediate Math
		{ID: 6, Difficulty: 6, Question: "What is the square root of 144?", Options: []string{"10", "11", "12", "14"}, Answer: "C"},
		{ID: 31, Difficulty: 6, Question: "What is 15% of 200?", Options: []string{"25", "30", "35", "40"}, Answer: "B"},
		{ID: 32, Difficulty: 6, Question: "What is the value of π (pi) to two decimal places?", Options: []string{"3.12", "3.14", "3.16", "3.18"}, Answer: "B"},
		{ID: 33, Difficulty: 6, Question: "What is 2 to the power of 5?", Options: []string{"16", "32", "64", "128"}, Answer: "B"},
		{ID: 34, Difficulty: 6, Question: "What is the area of a circle with radius 5? (π ≈ 3.14)", Options: []string{"78.5", "31.4", "15.7", "62.8"}, Answer: "A"},

		// Difficulty 7 - Geography & World Knowledge
		{ID: 7, Difficulty: 7, Question: "Which continent is the Sahara Desert located in?", Options: []string{"Asia", "Africa", "South America", "Australia"}, Answer: "B"},
		{ID: 35, Difficulty: 7, Question: "What is the longest river in the world?", Options: []string{"Amazon", "Nile", "Yangtze", "Mississippi"}, Answer: "B"},
		{ID: 36, Difficulty: 7, Question: "Which country is home to the Great Barrier Reef?", Options: []string{"New Zealand", "Australia", "Indonesia", "Philippines"}, Answer: "B"},
		{ID: 37, Difficulty: 7, Question: "What is the smallest country in the world?", Options: []string{"Monaco", "Vatican City", "San Marino", "Liechtenstein"}, Answer: "B"},
		{ID: 38, Difficulty: 7, Question: "Which mountain range separates Europe from Asia?", Options: []string{"Alps", "Himalayas", "Ural Mountains", "Andes"}, Answer: "C"},

		// Difficulty 8 - History
		{ID: 8, Difficulty: 8, Question: "In what year did the Titanic sink?", Options: []string{"1905", "1912", "1918", "1922"}, Answer: "B"},
		{ID: 39, Difficulty: 8, Question: "Who was the first person to walk on the moon?", Options: []string{"Buzz Aldrin", "Neil Armstrong", "Michael Collins", "John Glenn"}, Answer: "B"},
		{ID: 40, Difficulty: 8, Question: "In which year did the Berlin Wall fall?", Options: []string{"1987", "1989", "1991", "1993"}, Answer: "B"},
		{ID: 41, Difficulty: 8, Question: "Who was the first President of the United States?", Options: []string{"Thomas Jefferson", "John Adams", "George Washington", "Benjamin Franklin"}, Answer: "C"},
		{ID: 42, Difficulty: 8, Question: "In which year did World War I begin?", Options: []string{"1912", "1914", "1916", "1918"}, Answer: "B"},

		// Difficulty 9 - Biology & Medicine
		{ID: 9, Difficulty: 9, Question: "What is the largest organ in the human body?", Options: []string{"Heart", "Liver", "Skin", "Lungs"}, Answer: "C"},
		{ID: 43, Difficulty: 9, Question: "How many chambers does the human heart have?", Options: []string{"2", "3", "4", "5"}, Answer: "C"},
		{ID: 44, Difficulty: 9, Question: "What is the powerhouse of the cell?", Options: []string{"Nucleus", "Mitochondria", "Ribosome", "Golgi Apparatus"}, Answer: "B"},
		{ID: 45, Difficulty: 9, Question: "How many bones are in an adult human body?", Options: []string{"196", "206", "216", "226"}, Answer: "B"},
		{ID: 46, Difficulty: 9, Question: "What is the scientific name for the human species?", Options: []string{"Homo erectus", "Homo sapiens", "Homo habilis", "Homo neanderthalensis"}, Answer: "B"},

		// Difficulty 10 - Advanced Science & Physics
		{ID: 10, Difficulty: 10, Question: "Which physicist developed the theory of General Relativity?", Options: []string{"Newton", "Bohr", "Einstein", "Hawking"}, Answer: "C"},
		{ID: 47, Difficulty: 10, Question: "What is the speed of light in vacuum (approximately)?", Options: []string{"300,000 km/s", "150,000 km/s", "450,000 km/s", "600,000 km/s"}, Answer: "A"},
		{ID: 48, Difficulty: 10, Question: "What is the smallest unit of matter?", Options: []string{"Molecule", "Atom", "Electron", "Quark"}, Answer: "B"},
		{ID: 49, Difficulty: 10, Question: "Who discovered the law of gravity?", Options: []string{"Galileo", "Newton", "Einstein", "Kepler"}, Answer: "B"},
		{ID: 50, Difficulty: 10, Question: "What is the formula for energy (Einstein's equation)?", Options: []string{"E = mc", "E = mc²", "E = mv²", "E = mgh"}, Answer: "B"},
	}
}
