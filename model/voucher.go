package model

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type Voucher struct {
	ID     int64   `json:"id"`
	Code   string  `json:"code"`
	Amount float64 `json:"amount"`
	Limit  int     `json:"limit"`
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
	Save(voucher *Voucher) error
	Find(voucherCode string) (*Voucher, error)
	FindAll() ([]Voucher, error)
	Delete(voucherCode string) error
}

type RedemptionRepo interface {
	Create(redemption *Redemption) error
	FindRedemptions(voucherCode string, limit, offset int) ([]Redemption, error)
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

	return result > 0, nil
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

func (r *SQLRedemptionRepo) Create(redemption *Redemption) error {
	return r.DB.Create(redemption).Error
}

func (r *SQLRedemptionRepo) FindRedemptions(voucherCode string, limit, offset int) ([]Redemption, error) {
	var result []Redemption

	err := r.DB.Where("code = ?", voucherCode).Limit(limit).Offset(offset).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (v *SQLVoucherRepo) Save(voucher *Voucher) error {
	return v.DB.Save(voucher).Error
}

func (v *SQLVoucherRepo) Find(voucherCode string) (*Voucher, error) {
	var result Voucher

	if err := v.DB.Where("code = ?", voucherCode).Find(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (v *SQLVoucherRepo) FindAll() ([]Voucher, error) {
	var result []Voucher

	if err := v.DB.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (v *SQLVoucherRepo) Delete(voucherCode string) error {
	return v.DB.Where("code = ?", voucherCode).Delete(&Voucher{}).Error
}
