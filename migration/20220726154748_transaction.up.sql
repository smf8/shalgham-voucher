create table if not exists transactions(
                                           id bigserial primary key,
                                           profile_id bigint not null,
                                           amount real not null,
                                           constraint profile_fk foreign key (profile_id) references profiles(id)
);

