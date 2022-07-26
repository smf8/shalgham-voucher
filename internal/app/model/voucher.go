package model

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var ErrRecordNotFound = errors.New("record not found")

type Voucher struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Amount    float64   `json:"amount"`
	Limit     int       `json:"limit"`
	CreatedAt time.Time `json:"created_at"`
}

type Redemption struct {
	ID          int64     `json:"id"`
	VoucherCode string    `json:"voucher_code"`
	Redeemer    string    `json:"redeemer"`
	CreatedAt   time.Time `json:"created_at"`
}

type VoucherRemainderRepo interface {
	Revert(ctx context.Context, voucherCode string) error
	Use(ctx context.Context, voucherCode string) (bool, error)
	Create(ctx context.Context, voucherCode string, remainder int) error
	Get(ctx context.Context, voucherCode string) (int, error)
}

type VoucherRepo interface {
	Save(ctx context.Context, voucher *Voucher) error
	Find(ctx context.Context, voucherCode string) (*Voucher, error)
	FindAll(ctx context.Context) ([]Voucher, error)
	Delete(ctx context.Context, voucherCode string) error
}

type RedemptionRepo interface {
	Delete(ctx context.Context, redemption *Redemption) error
	Create(ctx context.Context, redemption *Redemption) error
	FindRedemptions(ctx context.Context, voucherCode string, limit, offset int) ([]Redemption, error)
}

type SQLVoucherRepo struct {
	DB *gorm.DB
}

type SQLRedemptionRepo struct {
	DB *gorm.DB
}

type RedisVoucherRemainderRepo struct {
	Redis redis.Cmdable
}

func (r *RedisVoucherRemainderRepo) voucherRemainderKey(voucherCode string) string {
	return "voucher:remainder:" + voucherCode
}

func (r *RedisVoucherRemainderRepo) Use(ctx context.Context, voucherCode string) (bool, error) {
	result, err := r.Redis.Decr(ctx, r.voucherRemainderKey(voucherCode)).Result()
	if err != nil {
		return false, err
	}

	return result >= 0, nil
}

func (r *RedisVoucherRemainderRepo) Get(ctx context.Context, voucherCode string) (int, error) {
	res, err := r.Redis.Get(ctx, r.voucherRemainderKey(voucherCode)).Result()
	if err != nil {
		return 0, err
	}

	intResult, err := strconv.ParseInt(res, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(intResult), nil
}

func (r *RedisVoucherRemainderRepo) Revert(ctx context.Context, voucherCode string) error {
	return r.Redis.Incr(ctx, r.voucherRemainderKey(voucherCode)).Err()
}

func (r *RedisVoucherRemainderRepo) Create(ctx context.Context, voucherCode string, remainder int) error {
	return r.Redis.Set(ctx, r.voucherRemainderKey(voucherCode), remainder, 0).Err()
}

func (r *SQLRedemptionRepo) Delete(ctx context.Context, redemption *Redemption) error {
	return r.DB.WithContext(ctx).Delete(redemption).Error
}

func (r *SQLRedemptionRepo) Create(ctx context.Context, redemption *Redemption) error {
	return r.DB.WithContext(ctx).Create(redemption).Error
}

func (r *SQLRedemptionRepo) FindRedemptions(ctx context.Context, voucherCode string, limit, offset int) ([]Redemption, error) {
	var result []Redemption

	err := r.DB.WithContext(ctx).Where("voucher_code = ?", voucherCode).Limit(limit).Offset(offset).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (v *SQLVoucherRepo) Save(ctx context.Context, voucher *Voucher) error {
	return v.DB.WithContext(ctx).Save(voucher).Error
}

func (v *SQLVoucherRepo) Find(ctx context.Context, voucherCode string) (*Voucher, error) {
	var result Voucher

	if err := v.DB.WithContext(ctx).Where("code = ?", voucherCode).First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("sql voucher find error: %w", ErrRecordNotFound)
		}

		return nil, err
	}

	return &result, nil
}

func (v *SQLVoucherRepo) FindAll(ctx context.Context) ([]Voucher, error) {
	var result []Voucher

	if err := v.DB.WithContext(ctx).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (v *SQLVoucherRepo) Delete(ctx context.Context, voucherCode string) error {
	return v.DB.WithContext(ctx).Where("code = ?", voucherCode).Delete(&Voucher{}).Error
}
