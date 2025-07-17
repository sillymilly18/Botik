-- +goose Up
-- +goose StatementBegin
create table if not exists users(
    id bigint not null primary key,
    is_notifications_enabled boolean not null default false,
    created_at timestamp not null default clock_timestamp()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists users;
-- +goose StatementEnd
