package main

import (
	"log"
)

// QuestionService handles question-related business logic (next question, recording asked).
type QuestionService struct {
	questionRepo *QuestionRepository
	userRepo     *UserRepository
	userService  *UserService
}

// NewQuestionService creates a new question service.
func NewQuestionService(questionRepo *QuestionRepository, userRepo *UserRepository, userService *UserService) *QuestionService {
	return &QuestionService{
		questionRepo: questionRepo,
		userRepo:     userRepo,
		userService:  userService,
	}
}

// GetNextQuestionForUser returns the next question for a user.
// Uses a single join query to return the question directly (no second lookup).
func (s *QuestionService) GetNextQuestionForUser(userID int) (*Question, int, error) {
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		return nil, 0, err
	}

	currentDifficulty := user.CurrentDifficulty
	if currentDifficulty == 0 {
		currentDifficulty = 1
	}

	question, err := s.questionRepo.GetRandomQuestionForUser(userID, currentDifficulty)
	if err != nil {
		return nil, 0, err
	}
	if question == nil {
		return nil, 0, ErrQuestionNotFound
	}

	if err := s.userRepo.RecordQuestionAsked(userID, question.ID); err != nil {
		log.Printf("Failed to record question asked for userID %d, questionID %d: %v", userID, question.ID, err)
	}

	return question, currentDifficulty, nil
}
