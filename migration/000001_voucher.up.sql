create table if not exists vouchers (
    id bigserial primary key,
    code varchar(255) unique not null,
    amount real not null,
    "limit" int not null,
    created_at  timestamp   not null    default now()
);

create index vouchers_code_idx on vouchers(code);