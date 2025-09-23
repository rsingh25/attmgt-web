
-- name: GetShiftByCd :one
SELECT * FROM shift_def
WHERE shift_cd = $1;

-- name: GetShiftByGroup :many
SELECT * FROM shift_def
WHERE shift_group = ANY(@shiftGroups::text[]);;


-- name: CreateShiftDef :one
INSERT INTO shift_def (
  shift_cd,
  shift_group,
  in_time_start,
  in_time_end,
  min_halfday_duration,
  min_present_duration,
  double_shift_allowed,
  min_double_shift_duration
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8
)
RETURNING *;
