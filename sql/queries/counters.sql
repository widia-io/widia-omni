-- name: GetCounters :one
SELECT * FROM workspace_counters
WHERE workspace_id = $1;

-- name: IncrementTasksToday :exec
UPDATE workspace_counters
SET tasks_created_today = CASE
        WHEN tasks_today_date = CURRENT_DATE THEN tasks_created_today + 1
        ELSE 1
    END,
    tasks_today_date = CURRENT_DATE
WHERE workspace_id = $1;

-- name: IncrementTransactionsMonth :exec
UPDATE workspace_counters
SET transactions_month_count = CASE
        WHEN transactions_month = to_char(CURRENT_DATE, 'YYYY-MM') THEN transactions_month_count + 1
        ELSE 1
    END,
    transactions_month = to_char(CURRENT_DATE, 'YYYY-MM')
WHERE workspace_id = $1;

-- name: ResetDailyCounters :exec
UPDATE workspace_counters
SET tasks_created_today = 0, tasks_today_date = CURRENT_DATE
WHERE tasks_today_date < CURRENT_DATE;
