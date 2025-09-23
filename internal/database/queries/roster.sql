-- name: GetRoster :many
SELECT * FROM roster
WHERE emp_id = ANY(@empIds::text[])
  and roster_date between @startDate and @endDate;

-- name: CreateRoster :one
INSERT INTO roster (
  emp_id, roster_date, shift_cd
) VALUES (
  $1      , $2       , $3
)
RETURNING *;

-- name: UpdateRoster :one
WITH rows AS (
  UPDATE roster
    set  shift_cd = $3
  WHERE emp_id = $1
    and roster_date = $2
    and roster_date > @now
    RETURNING 1
)
SELECT count(*) FROM rows;

