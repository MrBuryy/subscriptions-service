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
func (s *SubscriptionService) List(
	ctx context.Context,
	userID *string,
	serviceName *string,
	limit int,
	offset int,
) ([]model.Subscription, error) {
	filter := model.SubscriptionFilter{
		Limit:  limit,
		Offset: offset,
	}

	if userID != nil {
		trimmedUserID := strings.TrimSpace(*userID)
		parsedUserID, err := uuid.Parse(trimmedUserID)
		if err != nil {
			return nil, ErrInvalidUserID
		}
		filter.UserID = &parsedUserID
	}

	if serviceName != nil {
		trimmedServiceName := strings.TrimSpace(*serviceName)
		if trimmedServiceName == "" {
			return nil, ErrInvalidService
		}
		filter.ServiceName = &trimmedServiceName
	}

	if limit <= 0 {
		filter.Limit = 10
	}

	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	subs, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return subs, nil
}

func (s *SubscriptionService) Update(
	ctx context.Context,
	id string,
	serviceName string,
	price int,
	userID string,
	startDate string,
	endDate *string,
) (*model.Subscription, error) {
	id = strings.TrimSpace(id)
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return nil, ErrInvalidService
	}

	if price <= 0 {
		return nil, ErrInvalidPrice
	}

	userID = strings.TrimSpace(userID)
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
		trimmedEndDate := strings.TrimSpace(*endDate)
		if trimmedEndDate == "" {
			return nil, ErrInvalidEndDate
		}

		end, err := ParseMonthYear(trimmedEndDate)
		if err != nil {
			return nil, ErrInvalidEndDate
		}

		if end.Before(parsedStartDate) {
			return nil, ErrInvalidEndDate
		}

		parsedEndDate = &end
	}

	sub := &model.Subscription{
		ID:          parsedID,
		ServiceName: serviceName,
		Price:       price,
		UserID:      parsedUserID,
		StartDate:   parsedStartDate,
		EndDate:     parsedEndDate,
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidID
	}

	if err := s.repo.Delete(ctx, parsedID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *SubscriptionService) Summary(
	ctx context.Context,
	from string,
	to string,
	userID *string,
	serviceName *string,
) (int, error) {
	parsedFrom, err := ParseMonthYear(from)
	if err != nil {
		return 0, ErrInvalidFrom
	}

	parsedTo, err := ParseMonthYear(to)
	if err != nil {
		return 0, ErrInvalidTo
	}

	if parsedFrom.After(parsedTo) {
		return 0, ErrInvalidPeriod
	}

	filter := model.SubscriptionFilter{
		Limit:  1000,
		Offset: 0,
	}

	if userID != nil {
		trimmedUserID := strings.TrimSpace(*userID)
		parsedUserID, err := uuid.Parse(trimmedUserID)
		if err != nil {
			return 0, ErrInvalidUserID
		}
		filter.UserID = &parsedUserID
	}

	if serviceName != nil {
		trimmedServiceName := strings.TrimSpace(*serviceName)
		if trimmedServiceName == "" {
			return 0, ErrInvalidService
		}
		filter.ServiceName = &trimmedServiceName
	}

	subs, err := s.repo.List(ctx, filter)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, sub := range subs {
		months := CountMonthsInIntersection(
			sub.StartDate,
			sub.EndDate,
			parsedFrom,
			parsedTo,
		)
		total += months * sub.Price
	}

	return total, nil
}