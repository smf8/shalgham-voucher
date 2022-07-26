create table if not exists redemptions(
                                          id bigserial primary key,
                                          voucher_code varchar(255) not null,
                                          redeemer varchar(20) unique not null,
                                          created_at  timestamp   not null    default now()
);