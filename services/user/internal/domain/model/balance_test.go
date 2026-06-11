package model

import (
	"errors"
	"testing"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

func TestPointsBalance_Add(t *testing.T) {
	tests := []struct {
		name    string
		start   int64
		amount  int64
		wantPts int64
		wantErr error
	}{
		{"adds positive", 100, 50, 150, nil},
		{"rejects zero", 100, 0, 100, domain.ErrInvalidInput},
		{"rejects negative", 100, -10, 100, domain.ErrInvalidInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &PointsBalance{Points: tt.start}
			err := b.Add(tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Add() err = %v, want %v", err, tt.wantErr)
			}
			if b.Points != tt.wantPts {
				t.Errorf("Points = %d, want %d", b.Points, tt.wantPts)
			}
		})
	}
}

func TestPointsBalance_Deduct(t *testing.T) {
	tests := []struct {
		name    string
		start   int64
		amount  int64
		wantPts int64
		wantErr error
	}{
		{"deducts within balance", 100, 40, 60, nil},
		{"deducts exact balance", 100, 100, 0, nil},
		{"rejects over balance", 100, 101, 100, domain.ErrInsufficientBalance},
		{"rejects zero", 100, 0, 100, domain.ErrInvalidInput},
		{"rejects negative", 100, -10, 100, domain.ErrInvalidInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &PointsBalance{Points: tt.start}
			err := b.Deduct(tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Deduct() err = %v, want %v", err, tt.wantErr)
			}
			if b.Points != tt.wantPts {
				t.Errorf("Points = %d, want %d", b.Points, tt.wantPts)
			}
		})
	}
}
