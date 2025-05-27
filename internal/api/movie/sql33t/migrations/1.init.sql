create table if not exists movies (
    title_en text,
    overview text,
    vote_avg float,
    genre_ids_csv text,
    poster_src text,
    release_year int
);

create table if not exists genres(
    id int,
    name text
);