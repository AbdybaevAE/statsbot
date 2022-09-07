drop table stats cascade ;
drop table users cascade;
create table if not exists users(
                                    user_id bigserial not null primary key,
                                    tg_user_id int8 not null,
                                    tg_username varchar(1000) not null,
                                    lt_username varchar(1000) not null,
                                    created_at varchar(1000) not null,
                                    chat_id int8 not null,
                                    unique (chat_id, tg_user_id, lt_username)
);
create table if not exists stats(
                                    stat_id bigserial primary key,
                                    user_id int8 not null,
                                    stat_hash varchar(1000) not null unique,
                                    stat_easy_solved int not null,
                                    stat_easy_submissions int not null,
                                    stat_medium_solved int not null,
                                    stat_medium_submissions int not null,
                                    stat_hard_solved int not null,
                                    stat_hard_submissions int not null,
                                    unique(stat_hash),
                                    constraint fk_user_id foreign key (user_id) references users(user_id)
);