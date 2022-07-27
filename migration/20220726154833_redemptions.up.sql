create table if not exists redemptions(
                                          id bigserial primary key,
                                          voucher_code varchar(255) not null,
                                          redeemer varchar(20) not null,
                                          created_at  timestamp   not null    default now(),
                                          constraint unique_redemption unique (voucher_code, redeemer)
);