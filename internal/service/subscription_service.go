package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MrBuryy/subscriptions-service/internal/model"
	"github.com/MrBuryy/subscriptions-service/internal/repository"
)

type SubscriptionService struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) Create(
	ctx context.Context,
	serviceName string,
	price int,
	userID string,
	startDate string,
	endDate *string,
) (*model.Subscription, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return nil, ErrInvalidService
	}

	if price <= 0 {
		return nil, ErrInvalidPrice
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	parsedStartDate, err := ParseMonthYear(startDate)
	if err != nil {
		return nil, ErrInvalidStartDate
	}

	var parsedEndDate *time.Time
	if endDate != nil {
		t, err := ParseMonthYear(*endDate)
		if err != nil {
			return nil, ErrInvalidEndDate
		}

		if t.Before(parsedStartDate) {
			return nil, ErrInvalidEndDate
		}

		parsedEndDate = &t
	}

	sub := &model.Subscription{
		ServiceName: serviceName,
		Price:       price,
		UserID:      parsedUserID,
		StartDate:   parsedStartDate,
		EndDate:     parsedEndDate,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	sub, err := s.repo.GetByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return sub, nil
}