# github-metrics

In the following examples,`GITHUB_TOKEN` is a [personal access token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token).

Example: Print out a cute table of all currently open PRs for a repository
```shell
$ GITHUB_TOKEN=XXX go run ./cmd/list-open-prs <owner>/<repository>
+------+----------------+--------------------------------+-------------------------------+--------------+
|  ID  |     AUTHOR     |             TITLE              |          CREATED AT           |  OPEN TIME   |
+------+----------------+--------------------------------+-------------------------------+--------------+
| 1054 | usernamedt     | Greenplum backup-fetch MVP     | Fri, 06 Aug 2021 03:43:25 PDT | 40h6m29s     |
| 1037 | krnaveen14     | PostgreSQL - Simple Composer   | Wed, 14 Jul 2021 04:31:17 PDT | 591h18m37s   |
|      |                | for reduced memory usage       |                               |              |
| 1033 | LeGEC          | add a 'flags' subcommand to    | Tue, 06 Jul 2021 09:07:27 PDT | 778h42m28s   |
|      |                | list global flags              |                               |              |
...
```

Example: Load all closed PRs for a repository into a SQLite database.
```shell
$ GITHUB_TOKEN=XXX go run ./cmd/load-prs <path/to/sqlite.db> <owner> <repository>
2021/08/07 19:52:25 inserted records. count: 100
...
```

Analyze closed PR's by month.
```sqlite
select
    datetime(closed_at, 'start of month') as month,
    count(*) num_closed_prs,
    ROUND(avg(ROUND((JULIANDAY(closed_at) - JULIANDAY(created_at)) * 12)), 2) AS avg_hours_open
from prs
where closed_at is not null
group by month
order by closed_at;
```

Analyze closed PR's by author.
```sqlite
select author,
       count(*) as num_closed_prs,
       ROUND(avg(ROUND((JULIANDAY(closed_at) - JULIANDAY(created_at)) * 12)), 2) AS avg_hours_open
from prs
where closed_at is not null
group by author
order by num_closed_prs desc;
```

