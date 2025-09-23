
-- name: GetRawPunch :many
SELECT * FROM raw_punch
WHERE emp_id = ANY(@empIds::text[])
  and punch_ts between @startTs and @endTs;

-- name: CreateRawPunch :one
INSERT INTO raw_punch (emp_id, punch_ts, device, punch_type)
SELECT                 $1     , $2     , $3    , $4
WHERE NOT EXISTS (
    SELECT 1
    FROM raw_punch
    WHERE emp_id = $1
      and punch_ts = $2
)
RETURNING *;
